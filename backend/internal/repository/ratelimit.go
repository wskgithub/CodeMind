package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RateLimitRepository 限额配置数据访问层
type RateLimitRepository struct {
	db *gorm.DB
}

// NewRateLimitRepository 创建限额 Repository
func NewRateLimitRepository(db *gorm.DB) *RateLimitRepository {
	return &RateLimitRepository{db: db}
}

// Upsert 创建或更新限额配置
func (r *RateLimitRepository) Upsert(limit *model.RateLimit) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "target_type"},
			{Name: "target_id"},
			{Name: "period"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"max_tokens", "max_requests", "max_concurrency",
			"alert_threshold", "status", "updated_at",
		}),
	}).Create(limit).Error
}

// FindByID 根据 ID 查找限额配置
func (r *RateLimitRepository) FindByID(id int64) (*model.RateLimit, error) {
	var limit model.RateLimit
	err := r.db.First(&limit, id).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// FindByTarget 根据目标类型和 ID 查找限额配置
func (r *RateLimitRepository) FindByTarget(targetType string, targetID int64, period string) (*model.RateLimit, error) {
	var limit model.RateLimit
	err := r.db.Where("target_type = ? AND target_id = ? AND period = ?",
		targetType, targetID, period).First(&limit).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// ListByTarget 查询指定目标的所有限额配置
func (r *RateLimitRepository) ListByTarget(targetType string, targetID int64) ([]model.RateLimit, error) {
	var limits []model.RateLimit
	err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("period ASC").Find(&limits).Error
	return limits, err
}

// ListAll 查询所有限额配置
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

// Delete 删除限额配置
func (r *RateLimitRepository) Delete(id int64) error {
	return r.db.Delete(&model.RateLimit{}, id).Error
}

// GetEffectiveLimit 获取用户的有效限额配置（按优先级：用户 > 部门 > 全局）
func (r *RateLimitRepository) GetEffectiveLimit(userID int64, deptID *int64, period string) (*model.RateLimit, error) {
	// 1. 优先查用户级限额
	var limit model.RateLimit
	err := r.db.Where("target_type = ? AND target_id = ? AND period = ? AND status = 1",
		model.TargetTypeUser, userID, period).First(&limit).Error
	if err == nil {
		return &limit, nil
	}

	// 2. 其次查部门级限额
	if deptID != nil {
		err = r.db.Where("target_type = ? AND target_id = ? AND period = ? AND status = 1",
			model.TargetTypeDepartment, *deptID, period).First(&limit).Error
		if err == nil {
			return &limit, nil
		}
	}

	// 3. 最后查全局限额
	err = r.db.Where("target_type = ? AND target_id = 0 AND period = ? AND status = 1",
		model.TargetTypeGlobal, period).First(&limit).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}
