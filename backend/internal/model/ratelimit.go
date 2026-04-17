package model

import "time"

// RateLimit represents a rate limit configuration.
type RateLimit struct {
	CreatedAt      time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
	TargetType     string    `gorm:"size:20;not null;uniqueIndex:idx_rate_limits_target" json:"target_type"`
	Period         string    `gorm:"size:20;not null" json:"period"`
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetID       int64     `gorm:"not null;default:0;uniqueIndex:idx_rate_limits_target" json:"target_id"`
	PeriodHours    int       `gorm:"not null;default:24;uniqueIndex:idx_rate_limits_target" json:"period_hours"`
	MaxTokens      int64     `gorm:"not null" json:"max_tokens"`
	MaxRequests    int       `gorm:"not null;default:0" json:"max_requests"`
	MaxConcurrency int       `gorm:"not null;default:5" json:"max_concurrency"`
	AlertThreshold int16     `gorm:"not null;default:80" json:"alert_threshold"`
	Status         int16     `gorm:"not null;default:1" json:"status"`
}

// TableName returns the table name.
func (RateLimit) TableName() string {
	return "rate_limits"
}

// Rate limit target type constants.
const (
	TargetTypeGlobal     = "global"
	TargetTypeDepartment = "department"
	TargetTypeUser       = "user"
)

// Rate limit period constants.
const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
	PeriodCustom  = "custom"
)

// PeriodLabel returns a friendly period label based on hours.
func PeriodLabel(hours int) string {
	switch hours {
	case 24: //nolint:mnd // intentional constant.
		return PeriodDaily
	case 168: //nolint:mnd // intentional constant.
		return PeriodWeekly
	case 720: //nolint:mnd // intentional constant.
		return PeriodMonthly
	default:
		return PeriodCustom
	}
}

// PeriodHoursFromLabel returns default hours from a period label.
func PeriodHoursFromLabel(period string) int {
	switch period {
	case PeriodDaily:
		return 24 //nolint:mnd // intentional constant.
	case PeriodWeekly:
		return 168 //nolint:mnd // intentional constant.
	case PeriodMonthly:
		return 720 //nolint:mnd // intentional constant.
	default:
		return 24 //nolint:mnd // intentional constant.
	}
}

// EffectiveHours returns the effective period hours.
func (r *RateLimit) EffectiveHours() int {
	if r.PeriodHours > 0 {
		return r.PeriodHours
	}
	return PeriodHoursFromLabel(r.Period)
}
