package model

import "time"

// RateLimit represents a rate limit configuration.
type RateLimit struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetType     string    `gorm:"size:20;not null;uniqueIndex:idx_rate_limits_target" json:"target_type"` // global | department | user
	TargetID       int64     `gorm:"not null;default:0;uniqueIndex:idx_rate_limits_target" json:"target_id"`
	Period         string    `gorm:"size:20;not null" json:"period"`
	PeriodHours    int       `gorm:"not null;default:24;uniqueIndex:idx_rate_limits_target" json:"period_hours"`
	MaxTokens      int64     `gorm:"not null" json:"max_tokens"`
	MaxRequests    int       `gorm:"not null;default:0" json:"max_requests"`
	MaxConcurrency int       `gorm:"not null;default:5" json:"max_concurrency"`
	AlertThreshold int16     `gorm:"not null;default:80" json:"alert_threshold"`
	Status         int16     `gorm:"not null;default:1" json:"status"`
	CreatedAt      time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name.
func (RateLimit) TableName() string {
	return "rate_limits"
}

const (
	TargetTypeGlobal     = "global"
	TargetTypeDepartment = "department"
	TargetTypeUser       = "user"
)

const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
	PeriodCustom  = "custom"
)

// PeriodLabel returns a friendly period label based on hours.
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

// PeriodHoursFromLabel returns default hours from a period label.
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

// EffectiveHours returns the effective period hours.
func (r *RateLimit) EffectiveHours() int {
	if r.PeriodHours > 0 {
		return r.PeriodHours
	}
	return PeriodHoursFromLabel(r.Period)
}
