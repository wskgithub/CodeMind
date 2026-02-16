package repository

import (
	"time"

	"codemind/internal/model"

	"gorm.io/gorm"
)

// AuditRepository 审计日志数据访问层
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 创建审计日志 Repository
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create 创建审计日志
func (r *AuditRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// List 分页查询审计日志
func (r *AuditRepository) List(page, pageSize int, filters map[string]interface{}) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{}).Preload("Operator")

	// 应用过滤条件
	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}
	if operatorID, ok := filters["operator_id"].(*int64); ok && operatorID != nil {
		query = query.Where("operator_id = ?", *operatorID)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
