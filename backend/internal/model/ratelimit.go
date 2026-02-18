package model

import "time"

// RateLimit 限额配置模型
// 采用基于小时的灵活周期机制，统计最小粒度为小时。
// 周期从用户首次产生 token 时开始计时，到期后清零，
// 新周期需等待用户再次产生 token 才开启。
type RateLimit struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType     string    `gorm:"size:20;not null;uniqueIndex:idx_rate_limits_target" json:"target_type"` // global | department | user
	TargetID       int64     `gorm:"not null;default:0;uniqueIndex:idx_rate_limits_target" json:"target_id"`
	Period         string    `gorm:"size:20;not null" json:"period"`                                         // daily | weekly | monthly | custom（显示标签）
	PeriodHours    int       `gorm:"not null;default:24;uniqueIndex:idx_rate_limits_target" json:"period_hours"` // 实际周期时长（小时）
	MaxTokens      int64     `gorm:"not null" json:"max_tokens"`
	MaxRequests    int       `gorm:"not null;default:0" json:"max_requests"`
	MaxConcurrency int       `gorm:"not null;default:5" json:"max_concurrency"`
	AlertThreshold int16     `gorm:"not null;default:80" json:"alert_threshold"`
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

// 限额周期常量（显示标签）
const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
	PeriodCustom  = "custom"
)

// PeriodLabel 根据小时数返回友好的周期标签
func PeriodLabel(hours int) string {
	switch hours {
	case 24:
		return PeriodDaily
	case 168:
		return PeriodWeekly
	case 720:
		return PeriodMonthly
	default:
		return PeriodCustom
	}
}

// PeriodHoursFromLabel 从标签获取默认小时数
func PeriodHoursFromLabel(period string) int {
	switch period {
	case PeriodDaily:
		return 24
	case PeriodWeekly:
		return 168
	case PeriodMonthly:
		return 720
	default:
		return 24
	}
}

// EffectiveHours 返回有效的周期小时数
// 兼容旧数据：如果 PeriodHours 未正确设置（0），从 Period 标签推导
func (r *RateLimit) EffectiveHours() int {
	if r.PeriodHours > 0 {
		return r.PeriodHours
	}
	return PeriodHoursFromLabel(r.Period)
}
