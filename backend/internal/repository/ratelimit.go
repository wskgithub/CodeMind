package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RateLimitRepository handles rate limit data access.
type RateLimitRepository struct {
	db *gorm.DB
}

// NewRateLimitRepository creates a new rate limit repository.
func NewRateLimitRepository(db *gorm.DB) *RateLimitRepository {
	return &RateLimitRepository{db: db}
}

// Upsert creates or updates rate limit config.
func (r *RateLimitRepository) Upsert(limit *model.RateLimit) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "target_type"},
			{Name: "target_id"},
			{Name: "period_hours"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"period", "max_tokens", "max_requests", "max_concurrency",
			"alert_threshold", "status", "updated_at",
		}),
	}).Create(limit).Error
}

// FindByID returns rate limit by ID.
func (r *RateLimitRepository) FindByID(id int64) (*model.RateLimit, error) {
	var limit model.RateLimit
	err := r.db.First(&limit, id).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// FindByTarget returns rate limit by target type and ID.
func (r *RateLimitRepository) FindByTarget(targetType string, targetID int64, period string) (*model.RateLimit, error) {
	var limit model.RateLimit
	err := r.db.Where("target_type = ? AND target_id = ? AND period = ?",
		targetType, targetID, period).First(&limit).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// ListByTarget returns all rate limits for specified target.
func (r *RateLimitRepository) ListByTarget(targetType string, targetID int64) ([]model.RateLimit, error) {
	var limits []model.RateLimit
	err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("period ASC").Find(&limits).Error
	return limits, err
}

// ListAll returns all rate limit configs.
func (r *RateLimitRepository) ListAll(filters map[string]interface{}) ([]model.RateLimit, error) {
	var limits []model.RateLimit
	query := r.db.Model(&model.RateLimit{})

	if targetType, ok := filters["target_type"].(string); ok && targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if targetID, ok := filters["target_id"].(*int64); ok && targetID != nil {
		query = query.Where("target_id = ?", *targetID)
	}

	err := query.Order("target_type ASC, target_id ASC, period ASC").Find(&limits).Error
	return limits, err
}

// Delete removes rate limit config.
func (r *RateLimitRepository) Delete(id int64) error {
	return r.db.Delete(&model.RateLimit{}, id).Error
}

// GetEffectiveLimit returns effective rate limit (priority: user > department > global).
func (r *RateLimitRepository) GetEffectiveLimit(userID int64, deptID *int64, period string) (*model.RateLimit, error) {
	var limit model.RateLimit
	err := r.db.Where("target_type = ? AND target_id = ? AND period = ? AND status = 1",
		model.TargetTypeUser, userID, period).First(&limit).Error
	if err == nil {
		return &limit, nil
	}

	if deptID != nil {
		err = r.db.Where("target_type = ? AND target_id = ? AND period = ? AND status = 1",
			model.TargetTypeDepartment, *deptID, period).First(&limit).Error
		if err == nil {
			return &limit, nil
		}
	}

	err = r.db.Where("target_type = ? AND target_id = 0 AND period = ? AND status = 1",
		model.TargetTypeGlobal, period).First(&limit).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// GetAllEffectiveLimits returns all effective rate limits for user.
// Priority: user > department > global (only keeps highest priority for same period hours).
func (r *RateLimitRepository) GetAllEffectiveLimits(userID int64, deptID *int64) ([]model.RateLimit, error) {
	collected := make(map[int]model.RateLimit)

	var globalLimits []model.RateLimit
	r.db.Where("target_type = ? AND target_id = 0 AND status = 1",
		model.TargetTypeGlobal).Find(&globalLimits)
	for _, l := range globalLimits {
		collected[l.EffectiveHours()] = l
	}

	if deptID != nil {
		var deptLimits []model.RateLimit
		r.db.Where("target_type = ? AND target_id = ? AND status = 1",
			model.TargetTypeDepartment, *deptID).Find(&deptLimits)
		for _, l := range deptLimits {
			collected[l.EffectiveHours()] = l
		}
	}

	var userLimits []model.RateLimit
	r.db.Where("target_type = ? AND target_id = ? AND status = 1",
		model.TargetTypeUser, userID).Find(&userLimits)
	for _, l := range userLimits {
		collected[l.EffectiveHours()] = l
	}

	result := make([]model.RateLimit, 0, len(collected))
	for _, l := range collected {
		result = append(result, l)
	}
	return result, nil
}
