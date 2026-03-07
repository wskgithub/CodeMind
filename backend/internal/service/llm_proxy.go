package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"
	"codemind/pkg/llm"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// LLMProxyService LLM 代理业务逻辑
// 集成负载均衡器和基于周期的限额系统
type LLMProxyService struct {
	providerManager  *llm.ProviderManager
	loadBalancer     *llm.LoadBalancer
	usageRepo        *repository.UsageRepository
	limitRepo        *repository.RateLimitRepository
	keyRepo          *repository.APIKeyRepository
	trainingDataRepo *repository.TrainingDataRepository
	sysConfigRepo    *repository.SystemRepository
	limitService     *LimitService
	rdb              *redis.Client
	logger           *zap.Logger

	// 训练数据采集开关缓存
	trainingEnabled     bool
	trainingEnabledAt   time.Time
	trainingEnabledOnce sync.Once
	trainingEnabledMu   sync.RWMutex
}

// NewLLMProxyService 创建 LLM 代理服务
func NewLLMProxyService(
	providerManager *llm.ProviderManager,
	loadBalancer *llm.LoadBalancer,
	usageRepo *repository.UsageRepository,
	limitRepo *repository.RateLimitRepository,
	keyRepo *repository.APIKeyRepository,
	trainingDataRepo *repository.TrainingDataRepository,
	sysConfigRepo *repository.SystemRepository,
	limitService *LimitService,
	rdb *redis.Client,
	logger *zap.Logger,
) *LLMProxyService {
	return &LLMProxyService{
		providerManager:  providerManager,
		loadBalancer:     loadBalancer,
		usageRepo:        usageRepo,
		limitRepo:        limitRepo,
		keyRepo:          keyRepo,
		trainingDataRepo: trainingDataRepo,
		sysConfigRepo:    sysConfigRepo,
		limitService:     limitService,
		rdb:              rdb,
		logger:           logger,
	}
}

// GetProviderManager 获取 Provider 管理器
func (s *LLMProxyService) GetProviderManager() *llm.ProviderManager {
	return s.providerManager
}

// GetProviderForModel 根据模型名称获取合适的 Provider
// 优先使用负载均衡器选择；如果无可用节点，回退到 ProviderManager 静态路由
func (s *LLMProxyService) GetProviderForModel(ctx context.Context, userID int64, modelName string) (llm.Provider, error) {
	if s.loadBalancer != nil && s.loadBalancer.NodeCount() > 0 {
		provider, err := s.loadBalancer.SelectProvider(ctx, userID, modelName)
		if err == nil {
			return provider, nil
		}
		s.logger.Warn("负载均衡选择失败，回退到静态路由",
			zap.String("model", modelName), zap.Error(err))
	}
	return s.providerManager.RouteByModel(modelName)
}

// AcquireConcurrency 获取并发槽位
// 从所有生效的限额规则中取最大并发值；长周期规则的并发值优先级更高
func (s *LLMProxyService) AcquireConcurrency(ctx context.Context, userID int64, deptID *int64) (bool, error) {
	maxConcurrency := 5

	limits, err := s.limitRepo.GetAllEffectiveLimits(userID, deptID)
	if err == nil && len(limits) > 0 {
		// 取所有规则中最宽松的并发上限（任意一条允许即可）
		best := 0
		for _, l := range limits {
			if l.MaxConcurrency > best {
				best = l.MaxConcurrency
			}
		}
		if best > 0 {
			maxConcurrency = best
		}
	}

	key := fmt.Sprintf("codemind:concurrency:%d", userID)
	current, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		s.logger.Error("Redis INCR 失败", zap.Error(err))
		return true, nil
	}

	s.rdb.Expire(ctx, key, 5*time.Minute)

	if current > int64(maxConcurrency) {
		s.rdb.Decr(ctx, key)
		return false, nil
	}

	return true, nil
}

// ReleaseConcurrency 释放并发槽位
func (s *LLMProxyService) ReleaseConcurrency(ctx context.Context, userID int64) {
	key := fmt.Sprintf("codemind:concurrency:%d", userID)
	result, err := s.rdb.Decr(ctx, key).Result()
	if err != nil {
		s.logger.Error("Redis DECR 失败", zap.Error(err))
		return
	}
	if result < 0 {
		s.rdb.Set(ctx, key, 0, 5*time.Minute)
	}
}

// CheckTokenQuota 检查 Token 用量配额
// 使用新版基于周期的限额系统
func (s *LLMProxyService) CheckTokenQuota(ctx context.Context, userID int64, deptID *int64) (bool, error) {
	return s.limitService.CheckAllQuotas(ctx, userID, deptID)
}

// RecordUsage 记录请求用量（异步调用）
// 同时更新：用量明细表、每日汇总表、周期限额计数器
func (s *LLMProxyService) RecordUsage(
	userID, keyID int64,
	deptID *int64,
	modelName, requestType string,
	usage *llm.Usage,
	durationMs int,
) {
	if usage == nil {
		return
	}

	ctx := context.Background()

	// 1. 写入明细表
	record := &model.TokenUsage{
		UserID:           userID,
		APIKeyID:         keyID,
		Model:            modelName,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		RequestType:      requestType,
		DurationMs:       &durationMs,
	}
	if err := s.usageRepo.CreateUsage(record); err != nil {
		s.logger.Error("写入用量明细失败", zap.Error(err), zap.Int64("user_id", userID))
	}

	// 2. 更新每日汇总
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := s.usageRepo.UpsertDaily(userID, today, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); err != nil {
		s.logger.Error("更新每日汇总失败", zap.Error(err), zap.Int64("user_id", userID))
	}

	// 3. 更新周期限额计数器
	s.limitService.RecordCycleUsage(ctx, userID, deptID, usage.TotalTokens)

	// 4. 更新 Key 最后使用时间
	_ = s.keyRepo.UpdateLastUsed(keyID)
}

// RecordRequestLog 记录请求日志
func (s *LLMProxyService) RecordRequestLog(
	userID, keyID int64,
	requestType string,
	modelName string,
	statusCode int,
	errMsg string,
	clientIP, userAgent string,
	durationMs int,
) {
	var errMsgPtr *string
	if errMsg != "" {
		errMsgPtr = &errMsg
	}
	var modelPtr *string
	if modelName != "" {
		modelPtr = &modelName
	}

	log := &model.RequestLog{
		UserID:       userID,
		APIKeyID:     keyID,
		RequestType:  requestType,
		Model:        modelPtr,
		StatusCode:   statusCode,
		ErrorMessage: errMsgPtr,
		ClientIP:     &clientIP,
		UserAgent:    &userAgent,
		DurationMs:   &durationMs,
	}

	if err := s.usageRepo.CreateRequestLog(log); err != nil {
		s.logger.Error("写入请求日志失败", zap.Error(err))
	}
}

// RecordTrainingData 记录 LLM 请求/响应用于模型训练（异步调用）
// 内部检查系统配置开关，关闭时直接跳过
func (s *LLMProxyService) RecordTrainingData(
	userID, keyID int64,
	requestType, modelName string,
	isStream bool,
	requestBody, responseBody json.RawMessage,
	usage *llm.Usage,
	statusCode, durationMs int,
	clientIP string,
) {
	if !s.isTrainingDataEnabled() {
		return
	}

	var promptTokens, completionTokens, totalTokens int
	if usage != nil {
		promptTokens = usage.PromptTokens
		completionTokens = usage.CompletionTokens
		totalTokens = usage.TotalTokens
	}

	record := &model.LLMTrainingData{
		UserID:           userID,
		APIKeyID:         keyID,
		RequestType:      requestType,
		Model:            modelName,
		IsStream:         isStream,
		RequestBody:      requestBody,
		ResponseBody:     responseBody,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		DurationMs:       &durationMs,
		StatusCode:       statusCode,
		ClientIP:         &clientIP,
	}

	if err := s.trainingDataRepo.Create(record); err != nil {
		s.logger.Error("写入训练数据失败", zap.Error(err), zap.Int64("user_id", userID))
	}
}

// isTrainingDataEnabled 检查训练数据采集是否开启（带 60 秒缓存）
func (s *LLMProxyService) isTrainingDataEnabled() bool {
	const cacheTTL = 60 * time.Second

	s.trainingEnabledMu.RLock()
	if !s.trainingEnabledAt.IsZero() && time.Since(s.trainingEnabledAt) < cacheTTL {
		enabled := s.trainingEnabled
		s.trainingEnabledMu.RUnlock()
		return enabled
	}
	s.trainingEnabledMu.RUnlock()

	s.trainingEnabledMu.Lock()
	defer s.trainingEnabledMu.Unlock()

	// 双重检查
	if !s.trainingEnabledAt.IsZero() && time.Since(s.trainingEnabledAt) < cacheTTL {
		return s.trainingEnabled
	}

	enabled := true // 默认开启
	if s.sysConfigRepo != nil {
		cfg, err := s.sysConfigRepo.GetByKey(model.ConfigTrainingDataCollection)
		if err == nil {
			enabled = cfg.ConfigValue == "true"
		}
	}
	s.trainingEnabled = enabled
	s.trainingEnabledAt = time.Now()
	return enabled
}
