package model

import "time"

// LLMBackend represents an LLM backend service node for load balancing.
type LLMBackend struct {
	CreatedAt            time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
	HealthCheckURL       string    `gorm:"size:500;not null;default:''" json:"health_check_url"`
	BaseURL              string    `gorm:"size:500;not null" json:"base_url"`
	APIKey               string    `gorm:"size:500;not null;default:''" json:"-"`
	Format               string    `gorm:"size:20;not null;default:openai" json:"format"`
	ModelPatterns        string    `gorm:"type:text;not null;default:*" json:"model_patterns"`
	DisplayName          string    `gorm:"size:200;not null;default:''" json:"display_name"`
	Name                 string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Weight               int       `gorm:"not null;default:100" json:"weight"`
	MaxConcurrency       int       `gorm:"not null;default:100" json:"max_concurrency"`
	ID                   int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TimeoutSeconds       int       `gorm:"not null;default:300" json:"timeout_seconds"`
	StreamTimeoutSeconds int       `gorm:"not null;default:600" json:"stream_timeout_seconds"`
	Status               int16     `gorm:"not null;default:1" json:"status"`
}

// TableName returns the table name.
func (LLMBackend) TableName() string {
	return "llm_backends"
}

// LLM 后端状态常量。
const (
	LLMBackendDisabled = 0
	LLMBackendEnabled  = 1
	LLMBackendDraining = 2
)
