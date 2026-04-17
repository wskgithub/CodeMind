package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"codemind/internal/model"
	"codemind/internal/pkg/timezone"
	"codemind/internal/repository"
	"codemind/pkg/llm"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var acquireConcurrencyScript = redis.NewScript(`
	local current = redis.call('INCR', KEYS[1])
	redis.call('EXPIRE', KEYS[1], ARGV[1])
	return current
`)

// LLMProxyService handles LLM proxy operations with load balancing and quota management
type LLMProxyService struct {
	providerManager       *llm.ProviderManager
	loadBalancer          *llm.LoadBalancer
	usageRepo             *repository.UsageRepository
	limitRepo             *repository.RateLimitRepository
	keyRepo               *repository.APIKeyRepository
	trainingDataBuffer    *TrainingDataBuffer
	sysConfigRepo         *repository.SystemRepository
	limitService          *LimitService
	thirdPartyService     *ThirdPartyProviderService
	rdb                   *redis.Client
	logger                *zap.Logger
	sanitizer             *TrainingDataSanitizer
	conversationExtractor *ConversationExtractor
	deduplicator          *TrainingDataDeduplicator
	qualityScorer         *TrainingDataQualityScorer
	trainingEnabled       bool
	trainingEnabledAt     time.Time
	trainingEnabledOnce   sync.Once
	trainingEnabledMu     sync.RWMutex
}

// NewLLMProxyService creates an LLM proxy service
func NewLLMProxyService(
	providerManager *llm.ProviderManager,
	loadBalancer *llm.LoadBalancer,
	usageRepo *repository.UsageRepository,
	limitRepo *repository.RateLimitRepository,
	keyRepo *repository.APIKeyRepository,
	trainingDataBuffer *TrainingDataBuffer,
	sysConfigRepo *repository.SystemRepository,
	limitService *LimitService,
	rdb *redis.Client,
	logger *zap.Logger,
) *LLMProxyService {
	sanitizer := NewTrainingDataSanitizer(sysConfigRepo, logger)
	conversationExtractor := NewConversationExtractor()
	deduplicator := NewTrainingDataDeduplicator(rdb, sysConfigRepo, logger)
	qualityScorer := NewTrainingDataQualityScorer(sysConfigRepo, logger)

	return &LLMProxyService{
		providerManager:       providerManager,
		loadBalancer:          loadBalancer,
		usageRepo:             usageRepo,
		limitRepo:             limitRepo,
		keyRepo:               keyRepo,
		trainingDataBuffer:    trainingDataBuffer,
		sysConfigRepo:         sysConfigRepo,
		limitService:          limitService,
		rdb:                   rdb,
		logger:                logger,
		sanitizer:             sanitizer,
		conversationExtractor: conversationExtractor,
		deduplicator:          deduplicator,
		qualityScorer:         qualityScorer,
	}
}

// SetThirdPartyService injects third-party service (delayed to avoid circular dependency)
func (s *LLMProxyService) SetThirdPartyService(tps *ThirdPartyProviderService) {
	s.thirdPartyService = tps
}

// GetThirdPartyService returns the third-party service instance
func (s *LLMProxyService) GetThirdPartyService() *ThirdPartyProviderService {
	return s.thirdPartyService
}

// GetProviderManager returns the provider manager
func (s *LLMProxyService) GetProviderManager() *llm.ProviderManager {
	return s.providerManager
}

// GetProviderForModel selects a provider for the model (load balancer first, then static routing)
func (s *LLMProxyService) GetProviderForModel(ctx context.Context, userID int64, modelName string) (llm.Provider, error) {
	if s.loadBalancer != nil && s.loadBalancer.NodeCount() > 0 {
		provider, err := s.loadBalancer.SelectProvider(ctx, userID, modelName)
		if err == nil {
			return provider, nil
		}
		s.logger.Warn("load balancer selection failed, falling back to static routing",
			zap.String("model", modelName), zap.Error(err))
	}
	return s.providerManager.RouteByModel(modelName)
}

// AcquireConcurrency acquires a concurrency slot
func (s *LLMProxyService) AcquireConcurrency(ctx context.Context, userID int64, deptID *int64) (bool, error) {
	maxConcurrency := 5

	limits, err := s.limitRepo.GetAllEffectiveLimits(userID, deptID)
	if err == nil && len(limits) > 0 {
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
	current, err := acquireConcurrencyScript.Run(ctx, s.rdb, []string{key}, int(5*time.Minute.Seconds())).Int64()
	if err != nil {
		s.logger.Error("failed to acquire concurrency slot", zap.Error(err))
		return false, fmt.Errorf("concurrency control service unavailable")
	}

	if current > int64(maxConcurrency) {
		s.rdb.Decr(ctx, key)
		return false, nil
	}

	return true, nil
}

// ReleaseConcurrency releases a concurrency slot
func (s *LLMProxyService) ReleaseConcurrency(ctx context.Context, userID int64) {
	key := fmt.Sprintf("codemind:concurrency:%d", userID)
	result, err := s.rdb.Decr(ctx, key).Result()
	if err != nil {
		s.logger.Error("Redis DECR failed", zap.Error(err))
		return
	}
	if result < 0 {
		s.rdb.Set(ctx, key, 0, 5*time.Minute)
	}
}

// CheckTokenQuota checks token usage quota
func (s *LLMProxyService) CheckTokenQuota(ctx context.Context, userID int64, deptID *int64) (bool, error) {
	return s.limitService.CheckAllQuotas(ctx, userID, deptID)
}

// RecordUsage records request usage (async, parallel writes)
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

	var cacheCreationTokens, cacheReadTokens int
	if usage.PromptTokensDetails != nil {
		cacheCreationTokens = usage.PromptTokensDetails.CacheCreationInputTokens
		cacheReadTokens = usage.PromptTokensDetails.CacheReadInputTokens
	}

	ctx := context.Background()
	today := timezone.Today()

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		record := &model.TokenUsage{
			UserID:                   userID,
			APIKeyID:                 keyID,
			Model:                    modelName,
			PromptTokens:             usage.PromptTokens,
			CompletionTokens:         usage.CompletionTokens,
			TotalTokens:              usage.TotalTokens,
			CacheCreationInputTokens: cacheCreationTokens,
			CacheReadInputTokens:     cacheReadTokens,
			RequestType:              requestType,
			DurationMs:               &durationMs,
		}
		if err := s.usageRepo.CreateUsage(record); err != nil {
			s.logger.Error("failed to write usage record", zap.Error(err), zap.Int64("user_id", userID))
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.usageRepo.UpsertDaily(userID, today, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, cacheCreationTokens, cacheReadTokens); err != nil {
			s.logger.Error("failed to update daily summary", zap.Error(err), zap.Int64("user_id", userID))
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.usageRepo.UpsertDailyKey(keyID, userID, today, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, cacheCreationTokens, cacheReadTokens); err != nil {
			s.logger.Error("failed to update key daily summary", zap.Error(err), zap.Int64("api_key_id", keyID))
		}
	}()

	go func() {
		defer wg.Done()
		s.limitService.RecordCycleUsage(ctx, userID, deptID, usage.TotalTokens)
	}()

	go func() {
		defer wg.Done()
		_ = s.keyRepo.UpdateLastUsed(keyID)
	}()

	wg.Wait()
}

// RecordRequestLog records request log
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
		s.logger.Error("failed to write request log", zap.Error(err))
	}
}

// RecordTrainingData records LLM request/response for model training (async)
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

	sanitizedRequest := requestBody
	sanitizedResponse := responseBody
	if s.sanitizer != nil {
		sanitizedRequest = s.sanitizer.SanitizeRequestBody(requestBody)
		sanitizedResponse = s.sanitizer.SanitizeResponseBody(responseBody)
	}

	var contentHash *string
	if s.deduplicator != nil {
		hash := s.deduplicator.ComputeContentHash(sanitizedRequest, sanitizedResponse)
		if hash != "" {
			contentHash = &hash
			if s.deduplicator.IsDuplicate(hash) {
				s.logger.Debug("skipping duplicate training data",
					zap.Int64("user_id", userID),
					zap.String("hash", hash),
				)
				return
			}
			s.deduplicator.MarkAsSeen(hash)
		}
	}

	var conversationID *string
	if s.conversationExtractor != nil {
		convID := s.conversationExtractor.ExtractConversationIDFromMetadata(sanitizedRequest)
		if convID != "" {
			conversationID = &convID
		}
	}

	var qualityScore *int
	var promptTokens, completionTokens, totalTokens int
	if usage != nil {
		promptTokens = usage.PromptTokens
		completionTokens = usage.CompletionTokens
		totalTokens = usage.TotalTokens
	}
	if s.qualityScorer != nil {
		qualityScore = s.qualityScorer.Score(
			sanitizedRequest, sanitizedResponse,
			promptTokens, completionTokens,
			statusCode, &durationMs,
		)
	}

	record := &model.LLMTrainingData{
		UserID:           userID,
		APIKeyID:         keyID,
		RequestType:      requestType,
		Model:            modelName,
		IsStream:         isStream,
		RequestBody:      sanitizedRequest,
		ResponseBody:     sanitizedResponse,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		DurationMs:       &durationMs,
		StatusCode:     statusCode,
		ClientIP:       &clientIP,
		IsSanitized:    s.sanitizer != nil && s.sanitizer.IsEnabled(),
		ConversationID: conversationID,
		ContentHash:    contentHash,
		QualityScore:   qualityScore,
	}

	s.trainingDataBuffer.Add(record)
}

// RecordTrainingDataWithSource records training data with source tracking
func (s *LLMProxyService) RecordTrainingDataWithSource(
	userID, keyID int64,
	requestType, modelName string,
	isStream bool,
	requestBody, responseBody json.RawMessage,
	usage *llm.Usage,
	statusCode, durationMs int,
	clientIP string,
	source string,
	thirdPartyProviderID *int64,
) {
	if !s.isTrainingDataEnabled() {
		return
	}

	sanitizedRequest := requestBody
	sanitizedResponse := responseBody
	if s.sanitizer != nil {
		sanitizedRequest = s.sanitizer.SanitizeRequestBody(requestBody)
		sanitizedResponse = s.sanitizer.SanitizeResponseBody(responseBody)
	}

	var contentHash *string
	if s.deduplicator != nil {
		hash := s.deduplicator.ComputeContentHash(sanitizedRequest, sanitizedResponse)
		if hash != "" {
			contentHash = &hash
			if s.deduplicator.IsDuplicate(hash) {
				s.logger.Debug("skipping duplicate training data",
					zap.Int64("user_id", userID),
					zap.String("hash", hash),
				)
				return
			}
			s.deduplicator.MarkAsSeen(hash)
		}
	}

	var conversationID *string
	if s.conversationExtractor != nil {
		convID := s.conversationExtractor.ExtractConversationIDFromMetadata(sanitizedRequest)
		if convID != "" {
			conversationID = &convID
		}
	}

	var qualityScore *int
	var promptTokens, completionTokens, totalTokens int
	if usage != nil {
		promptTokens = usage.PromptTokens
		completionTokens = usage.CompletionTokens
		totalTokens = usage.TotalTokens
	}
	if s.qualityScorer != nil {
		qualityScore = s.qualityScorer.Score(
			sanitizedRequest, sanitizedResponse,
			promptTokens, completionTokens,
			statusCode, &durationMs,
		)
	}

	record := &model.LLMTrainingData{
		UserID:               userID,
		APIKeyID:             keyID,
		RequestType:          requestType,
		Model:                modelName,
		IsStream:             isStream,
		RequestBody:          sanitizedRequest,
		ResponseBody:         sanitizedResponse,
		PromptTokens:         promptTokens,
		CompletionTokens:     completionTokens,
		TotalTokens:          totalTokens,
		DurationMs:           &durationMs,
		StatusCode:           statusCode,
		ClientIP:             &clientIP,
		Source:               source,
		ThirdPartyProviderID: thirdPartyProviderID,
		IsSanitized:          s.sanitizer != nil && s.sanitizer.IsEnabled(),
		ConversationID: conversationID,
		ContentHash:    contentHash,
		QualityScore:   qualityScore,
	}

	s.trainingDataBuffer.Add(record)
}

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

	if !s.trainingEnabledAt.IsZero() && time.Since(s.trainingEnabledAt) < cacheTTL {
		return s.trainingEnabled
	}

	enabled := true
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
