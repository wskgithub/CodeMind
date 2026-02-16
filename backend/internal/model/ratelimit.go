package model

import "time"

// RateLimit 限额配置模型
type RateLimit struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType     string    `gorm:"size:20;not null;uniqueIndex:idx_rate_limits_target" json:"target_type"` // global | department | user
	TargetID       int64     `gorm:"not null;default:0;uniqueIndex:idx_rate_limits_target" json:"target_id"`
	Period         string    `gorm:"size:20;not null;uniqueIndex:idx_rate_limits_target" json:"period"` // daily | weekly | monthly
	MaxTokens      int64     `gorm:"not null" json:"max_tokens"`
	MaxRequests    int       `gorm:"not null;default:0" json:"max_requests"`       // 0 表示不限制
	MaxConcurrency int       `gorm:"not null;default:5" json:"max_concurrency"`
	AlertThreshold int16     `gorm:"not null;default:80" json:"alert_threshold"`   // 告警阈值百分比
	Status         int16     `gorm:"not null;default:1" json:"status"`
	CreatedAt      time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (RateLimit) TableName() string {
	return "rate_limits"
}

// 限额目标类型常量
const (
	TargetTypeGlobal     = "global"
	TargetTypeDepartment = "department"
	TargetTypeUser       = "user"
)

// 限额周期常量
const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
)
