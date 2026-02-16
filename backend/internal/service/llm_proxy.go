package service

import (
	"context"
	"fmt"
	"time"

	"codemind/internal/model"
	"codemind/internal/repository"
	"codemind/pkg/llm"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// LLMProxyService LLM 代理业务逻辑
type LLMProxyService struct {
	providerManager *llm.ProviderManager
	usageRepo       *repository.UsageRepository
	limitRepo       *repository.RateLimitRepository
	keyRepo         *repository.APIKeyRepository
	rdb             *redis.Client
	logger          *zap.Logger
}

// NewLLMProxyService 创建 LLM 代理服务
func NewLLMProxyService(
	providerManager *llm.ProviderManager,
	usageRepo *repository.UsageRepository,
	limitRepo *repository.RateLimitRepository,
	keyRepo *repository.APIKeyRepository,
	rdb *redis.Client,
	logger *zap.Logger,
) *LLMProxyService {
	return &LLMProxyService{
		providerManager: providerManager,
		usageRepo:       usageRepo,
		limitRepo:       limitRepo,
		keyRepo:         keyRepo,
		rdb:             rdb,
		logger:          logger,
	}
}

// GetProviderManager 获取 Provider 管理器
func (s *LLMProxyService) GetProviderManager() *llm.ProviderManager {
	return s.providerManager
}

// GetProviderForModel 根据模型名称获取合适的 Provider
func (s *LLMProxyService) GetProviderForModel(modelName string) (llm.Provider, error) {
	return s.providerManager.RouteByModel(modelName)
}

// AcquireConcurrency 获取并发槽位
// 返回 true 表示获取成功，false 表示已达并发上限
func (s *LLMProxyService) AcquireConcurrency(ctx context.Context, userID int64, deptID *int64) (bool, error) {
	// 获取用户的并发限制
	maxConcurrency := 5 // 默认值
	limit, err := s.limitRepo.GetEffectiveLimit(userID, deptID, model.PeriodDaily)
	if err == nil && limit != nil {
		maxConcurrency = limit.MaxConcurrency
	}

	key := fmt.Sprintf("codemind:concurrency:%d", userID)

	// 原子操作：INCR + 检查 + 条件 DECR
	current, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		s.logger.Error("Redis INCR 失败", zap.Error(err))
		return true, nil // Redis 故障时降级放行
	}

	// 设置 TTL 防止计数泄漏
	s.rdb.Expire(ctx, key, 5*time.Minute)

	if current > int64(maxConcurrency) {
		// 超出限制，回退计数
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
	// 防止计数器变为负数
	if result < 0 {
		s.rdb.Set(ctx, key, 0, 5*time.Minute)
	}
}

// CheckTokenQuota 检查 Token 用量配额
// 返回 true 表示配额充足，false 表示已超限
func (s *LLMProxyService) CheckTokenQuota(ctx context.Context, userID int64, deptID *int64) (bool, error) {
	// 检查每日配额
	dailyLimit, err := s.limitRepo.GetEffectiveLimit(userID, deptID, model.PeriodDaily)
	if err == nil && dailyLimit != nil && dailyLimit.MaxTokens > 0 {
		today := time.Now().Format("2006-01-02")
		usedKey := fmt.Sprintf("codemind:usage:%d:daily:%s", userID, today)

		used, err := s.rdb.Get(ctx, usedKey).Int64()
		if err != nil && err != redis.Nil {
			s.logger.Warn("读取 Redis 用量失败，回退到数据库", zap.Error(err))
			// 回退到数据库查询
			used, _ = s.usageRepo.GetPeriodUsage(userID, "daily", today)
		}

		if used >= dailyLimit.MaxTokens {
			return false, nil
		}
	}

	// 检查每月配额
	monthlyLimit, err := s.limitRepo.GetEffectiveLimit(userID, deptID, model.PeriodMonthly)
	if err == nil && monthlyLimit != nil && monthlyLimit.MaxTokens > 0 {
		month := time.Now().Format("2006-01")
		usedKey := fmt.Sprintf("codemind:usage:%d:monthly:%s", userID, month)

		used, err := s.rdb.Get(ctx, usedKey).Int64()
		if err != nil && err != redis.Nil {
			monthStart := time.Now().Format("2006-01") + "-01"
			used, _ = s.usageRepo.GetPeriodUsage(userID, "monthly", monthStart)
		}

		if used >= monthlyLimit.MaxTokens {
			return false, nil
		}
	}

	return true, nil
}

// RecordUsage 记录请求用量（异步调用）
func (s *LLMProxyService) RecordUsage(
	userID, keyID int64,
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

	// 2. 更新每日汇总（UPSERT）
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.usageRepo.UpsertDaily(userID, today, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); err != nil {
		s.logger.Error("更新每日汇总失败", zap.Error(err), zap.Int64("user_id", userID))
	}

	// 3. 更新 Redis 计数器
	todayStr := time.Now().Format("2006-01-02")
	monthStr := time.Now().Format("2006-01")

	dailyKey := fmt.Sprintf("codemind:usage:%d:daily:%s", userID, todayStr)
	monthlyKey := fmt.Sprintf("codemind:usage:%d:monthly:%s", userID, monthStr)

	s.rdb.IncrBy(ctx, dailyKey, int64(usage.TotalTokens))
	s.rdb.Expire(ctx, dailyKey, 48*time.Hour)

	s.rdb.IncrBy(ctx, monthlyKey, int64(usage.TotalTokens))
	s.rdb.Expire(ctx, monthlyKey, 35*24*time.Hour)

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
