package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// LimitService 限额管理业务逻辑
type LimitService struct {
	limitRepo *repository.RateLimitRepository
	usageRepo *repository.UsageRepository
	auditRepo *repository.AuditRepository
	rdb       *redis.Client
	logger    *zap.Logger
}

// NewLimitService 创建限额服务
func NewLimitService(
	limitRepo *repository.RateLimitRepository,
	usageRepo *repository.UsageRepository,
	auditRepo *repository.AuditRepository,
	rdb *redis.Client,
	logger *zap.Logger,
) *LimitService {
	return &LimitService{
		limitRepo: limitRepo,
		usageRepo: usageRepo,
		auditRepo: auditRepo,
		rdb:       rdb,
		logger:    logger,
	}
}

// List 获取限额配置列表
func (s *LimitService) List(query *dto.LimitListQuery) ([]model.RateLimit, error) {
	filters := map[string]interface{}{
		"target_type": query.TargetType,
		"target_id":   query.TargetID,
	}
	limits, err := s.limitRepo.ListAll(filters)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	return limits, nil
}

// Upsert 创建或更新限额配置
func (s *LimitService) Upsert(req *dto.UpsertRateLimitRequest, operatorID int64, clientIP string) error {
	limit := &model.RateLimit{
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		Period:         req.Period,
		MaxTokens:      req.MaxTokens,
		MaxRequests:    req.MaxRequests,
		MaxConcurrency: req.MaxConcurrency,
		AlertThreshold: req.AlertThreshold,
		Status:         model.StatusEnabled,
	}

	if limit.MaxConcurrency == 0 {
		limit.MaxConcurrency = 5
	}
	if limit.AlertThreshold == 0 {
		limit.AlertThreshold = 80
	}

	if err := s.limitRepo.Upsert(limit); err != nil {
		return errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionUpdateLimit, model.AuditTargetRateLimit, nil,
		map[string]interface{}{"target_type": req.TargetType, "target_id": req.TargetID, "period": req.Period}, clientIP)

	return nil
}

// GetMyLimits 获取当前用户的限额和用量信息
func (s *LimitService) GetMyLimits(userID int64, deptID *int64) (*dto.MyLimitResponse, error) {
	ctx := context.Background()
	resp := &dto.MyLimitResponse{
		Limits: make(map[string]dto.LimitDetail),
	}

	// 查询每日限额
	dailyLimit, err := s.limitRepo.GetEffectiveLimit(userID, deptID, model.PeriodDaily)
	if err == nil && dailyLimit != nil {
		todayStr := time.Now().Format("2006-01-02")
		usedKey := fmt.Sprintf("codemind:usage:%d:daily:%s", userID, todayStr)
		used, _ := s.rdb.Get(ctx, usedKey).Int64()
		if used == 0 {
			used, _ = s.usageRepo.GetPeriodUsage(userID, "daily", todayStr)
		}

		remaining := dailyLimit.MaxTokens - used
		if remaining < 0 {
			remaining = 0
		}
		percent := 0
		if dailyLimit.MaxTokens > 0 {
			percent = int(used * 100 / dailyLimit.MaxTokens)
		}

		resp.Limits["daily"] = dto.LimitDetail{
			MaxTokens:       dailyLimit.MaxTokens,
			UsedTokens:      used,
			RemainingTokens: remaining,
			UsagePercent:    percent,
		}
	}

	// 查询每月限额
	monthlyLimit, err := s.limitRepo.GetEffectiveLimit(userID, deptID, model.PeriodMonthly)
	if err == nil && monthlyLimit != nil {
		monthStr := time.Now().Format("2006-01")
		usedKey := fmt.Sprintf("codemind:usage:%d:monthly:%s", userID, monthStr)
		used, _ := s.rdb.Get(ctx, usedKey).Int64()
		if used == 0 {
			monthStart := time.Now().Format("2006-01") + "-01"
			used, _ = s.usageRepo.GetPeriodUsage(userID, "monthly", monthStart)
		}

		remaining := monthlyLimit.MaxTokens - used
		if remaining < 0 {
			remaining = 0
		}
		percent := 0
		if monthlyLimit.MaxTokens > 0 {
			percent = int(used * 100 / monthlyLimit.MaxTokens)
		}

		resp.Limits["monthly"] = dto.LimitDetail{
			MaxTokens:       monthlyLimit.MaxTokens,
			UsedTokens:      used,
			RemainingTokens: remaining,
			UsagePercent:    percent,
		}
	}

	// 并发信息
	concurrencyKey := fmt.Sprintf("codemind:concurrency:%d", userID)
	current, _ := s.rdb.Get(ctx, concurrencyKey).Int()
	maxConcurrency := 5
	if dailyLimit != nil {
		maxConcurrency = dailyLimit.MaxConcurrency
	}

	resp.Concurrency = dto.ConcurrencyInfo{
		Max:     maxConcurrency,
		Current: current,
	}

	return resp, nil
}

// Delete 删除限额配置
func (s *LimitService) Delete(id int64, operatorID int64, clientIP string) error {
	limit, err := s.limitRepo.FindByID(id)
	if err != nil {
		return errcode.ErrRecordNotFound
	}

	if err := s.limitRepo.Delete(id); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionDeleteLimit, model.AuditTargetRateLimit, &id,
		map[string]interface{}{"target_type": limit.TargetType, "target_id": limit.TargetID}, clientIP)

	return nil
}

// recordAudit 记录审计日志
func (s *LimitService) recordAudit(operatorID int64, action, targetType string, targetID *int64, detail interface{}, clientIP string) {
	var detailJSON json.RawMessage
	if detail != nil {
		data, _ := json.Marshal(detail)
		detailJSON = data
	}

	log := &model.AuditLog{
		OperatorID: operatorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detailJSON,
		ClientIP:   &clientIP,
	}

	if err := s.auditRepo.Create(log); err != nil {
		s.logger.Error("记录审计日志失败", zap.Error(err), zap.String("action", action))
	}
}
