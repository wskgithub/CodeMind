package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// LimitService 限额管理业务逻辑
// 采用基于小时的弹性周期机制：
//   - 周期从用户首次产生 token 时开始计时
//   - 到期后清零，等待用户再次产生 token 才开启新周期
//   - 不同规则的周期独立运作，互不干扰
//   - 长周期限额（如月）达到时，无论短周期是否达到均暂停服务
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

// ──────────────────────────────────
// 管理类方法
// ──────────────────────────────────

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
	// 确定实际周期小时数
	periodHours := req.PeriodHours
	if periodHours == 0 {
		periodHours = model.PeriodHoursFromLabel(req.Period)
	}

	limit := &model.RateLimit{
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		Period:         req.Period,
		PeriodHours:    periodHours,
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

	s.recordAudit(operatorID, model.AuditActionUpdateLimit, model.AuditTargetRateLimit, nil,
		map[string]interface{}{
			"target_type":  req.TargetType,
			"target_id":    req.TargetID,
			"period":       req.Period,
			"period_hours": periodHours,
		}, clientIP)

	return nil
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

// ──────────────────────────────────
// 周期限额核心逻辑
// ──────────────────────────────────

// cycleStartKey 周期起始时间的 Redis 键
func cycleStartKey(userID int64, ruleID int64) string {
	return fmt.Sprintf("codemind:cycle:%d:%d:start", userID, ruleID)
}

// cycleUsageKey 当前周期用量的 Redis 键
func cycleUsageKey(userID int64, ruleID int64) string {
	return fmt.Sprintf("codemind:cycle:%d:%d:usage", userID, ruleID)
}

// CheckAllQuotas 检查用户的所有限额规则
// 返回 true 表示配额充足，false 表示至少有一条规则超限
func (s *LimitService) CheckAllQuotas(ctx context.Context, userID int64, deptID *int64) (bool, error) {
	limits, err := s.limitRepo.GetAllEffectiveLimits(userID, deptID)
	if err != nil {
		s.logger.Warn("获取限额规则失败，降级放行", zap.Error(err))
		return true, nil
	}

	for _, limit := range limits {
		if limit.MaxTokens <= 0 {
			continue
		}
		exceeded, err := s.isRuleExceeded(ctx, userID, &limit)
		if err != nil {
			s.logger.Warn("检查限额失败，降级放行",
				zap.Int64("rule_id", limit.ID), zap.Error(err))
			continue
		}
		if exceeded {
			return false, nil
		}
	}
	return true, nil
}

// isRuleExceeded 检查单条规则是否超限
func (s *LimitService) isRuleExceeded(ctx context.Context, userID int64, rule *model.RateLimit) (bool, error) {
	startKey := cycleStartKey(userID, rule.ID)

	startStr, err := s.rdb.Get(ctx, startKey).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	startUnix, err := parseInt64(startStr)
	if err != nil {
		return false, err
	}

	periodDuration := time.Duration(rule.EffectiveHours()) * time.Hour
	if time.Now().After(time.Unix(startUnix, 0).Add(periodDuration)) {
		s.rdb.Del(ctx, startKey, cycleUsageKey(userID, rule.ID))
		return false, nil
	}

	usageKey := cycleUsageKey(userID, rule.ID)
	used, err := s.rdb.Get(ctx, usageKey).Int64()
	if err != nil && err != redis.Nil {
		return false, err
	}

	return used >= rule.MaxTokens, nil
}

// RecordCycleUsage 记录用量到各规则的当前周期
// 在每次请求完成后异步调用
func (s *LimitService) RecordCycleUsage(ctx context.Context, userID int64, deptID *int64, tokens int) {
	limits, err := s.limitRepo.GetAllEffectiveLimits(userID, deptID)
	if err != nil {
		s.logger.Error("获取限额规则失败（记录用量）", zap.Error(err))
		return
	}

	for _, limit := range limits {
		if limit.MaxTokens <= 0 {
			continue
		}
		s.recordRuleUsage(ctx, userID, &limit, int64(tokens))
	}
}

// recordRuleUsage 为单条规则记录用量
func (s *LimitService) recordRuleUsage(ctx context.Context, userID int64, rule *model.RateLimit, tokens int64) {
	startKey := cycleStartKey(userID, rule.ID)
	usageKey := cycleUsageKey(userID, rule.ID)
	periodDuration := time.Duration(rule.EffectiveHours()) * time.Hour
	ttl := periodDuration + 1*time.Hour

	startStr, err := s.rdb.Get(ctx, startKey).Result()
	if err == redis.Nil {
		// 无活跃周期 → 开启新周期
		now := time.Now().Unix()
		s.rdb.Set(ctx, startKey, fmt.Sprintf("%d", now), ttl)
		s.rdb.Set(ctx, usageKey, tokens, ttl)
		return
	}
	if err != nil {
		s.logger.Error("读取周期起始时间失败", zap.Error(err))
		return
	}

	// 检查周期是否已过期
	startUnix, _ := parseInt64(startStr)
	if time.Now().After(time.Unix(startUnix, 0).Add(periodDuration)) {
		// 旧周期已过期 → 开启新周期
		now := time.Now().Unix()
		s.rdb.Set(ctx, startKey, fmt.Sprintf("%d", now), ttl)
		s.rdb.Set(ctx, usageKey, tokens, ttl)
		return
	}

	// 在当前周期内累加用量
	s.rdb.IncrBy(ctx, usageKey, tokens)
}

// ──────────────────────────────────
// 进度查询
// ──────────────────────────────────

// GetLimitProgress 获取用户的限额进度信息（包含重置时间）
func (s *LimitService) GetLimitProgress(userID int64, deptID *int64) (*dto.LimitProgressResponse, error) {
	ctx := context.Background()

	limits, err := s.limitRepo.GetAllEffectiveLimits(userID, deptID)
	if err != nil {
		return nil, errcode.ErrDatabase
	}

	resp := &dto.LimitProgressResponse{
		Limits: make([]dto.LimitProgressItem, 0, len(limits)),
	}

	for _, limit := range limits {
		effectiveHours := limit.EffectiveHours()
		item := dto.LimitProgressItem{
			RuleID:      limit.ID,
			Period:      limit.Period,
			PeriodHours: effectiveHours,
			MaxTokens:   limit.MaxTokens,
		}

		if limit.MaxTokens <= 0 {
			resp.Limits = append(resp.Limits, item)
			continue
		}

		startKey := cycleStartKey(userID, limit.ID)
		startStr, err := s.rdb.Get(ctx, startKey).Result()
		if err == nil {
			startUnix, _ := parseInt64(startStr)
			periodEnd := time.Unix(startUnix, 0).Add(time.Duration(effectiveHours) * time.Hour)

			if time.Now().Before(periodEnd) {
				// 活跃周期
				cycleStart := startUnix
				resetAt := periodEnd.Unix()
				hoursLeft := time.Until(periodEnd).Hours()
				hoursLeft = math.Round(hoursLeft*10) / 10

				item.CycleStartAt = &cycleStart
				item.ResetAt = &resetAt
				item.ResetInHours = &hoursLeft

				// 获取当前周期用量
				usageKey := cycleUsageKey(userID, limit.ID)
				used, _ := s.rdb.Get(ctx, usageKey).Int64()
				item.UsedTokens = used
			}
			// 如果周期已过期，used 保持 0
		}

		item.RemainingTokens = limit.MaxTokens - item.UsedTokens
		if item.RemainingTokens < 0 {
			item.RemainingTokens = 0
		}
		if limit.MaxTokens > 0 {
			item.UsagePercent = int(item.UsedTokens * 100 / limit.MaxTokens)
			if item.UsagePercent > 100 {
				item.UsagePercent = 100
			}
		}
		item.Exceeded = item.UsedTokens >= limit.MaxTokens

		if item.Exceeded {
			resp.AnyExceeded = true
		}

		resp.Limits = append(resp.Limits, item)
	}

	// 并发信息
	concurrencyKey := fmt.Sprintf("codemind:concurrency:%d", userID)
	current, _ := s.rdb.Get(ctx, concurrencyKey).Int()
	maxConcurrency := 5
	if len(limits) > 0 {
		maxConcurrency = limits[0].MaxConcurrency
	}
	resp.Concurrency = dto.ConcurrencyInfo{
		Max:     maxConcurrency,
		Current: current,
	}

	return resp, nil
}

// GetMyLimits 兼容旧接口：获取当前用户的限额和用量信息
func (s *LimitService) GetMyLimits(userID int64, deptID *int64) (*dto.MyLimitResponse, error) {
	progress, err := s.GetLimitProgress(userID, deptID)
	if err != nil {
		return nil, err
	}

	resp := &dto.MyLimitResponse{
		Limits:      make(map[string]dto.LimitDetail),
		Concurrency: progress.Concurrency,
	}

	for _, item := range progress.Limits {
		resp.Limits[item.Period] = dto.LimitDetail{
			MaxTokens:       item.MaxTokens,
			UsedTokens:      item.UsedTokens,
			RemainingTokens: item.RemainingTokens,
			UsagePercent:    item.UsagePercent,
		}
	}

	return resp, nil
}

// ──────────────────────────────────
// 辅助方法
// ──────────────────────────────────

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

// parseInt64 安全解析 int64
func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
