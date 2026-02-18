package model

import "time"

// LLMBackend LLM 后端服务节点模型
// 用于多后端负载均衡，平台将用户请求动态调度到不同的 LLM 服务上
type LLMBackend struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                 string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	DisplayName          string    `gorm:"size:200;not null;default:''" json:"display_name"`
	BaseURL              string    `gorm:"size:500;not null" json:"base_url"`
	APIKey               string    `gorm:"size:500;not null;default:''" json:"-"` // API Key 不返回给前端
	Format               string    `gorm:"size:20;not null;default:openai" json:"format"` // openai | anthropic
	Weight               int       `gorm:"not null;default:100" json:"weight"`
	MaxConcurrency       int       `gorm:"not null;default:100" json:"max_concurrency"`
	Status               int16     `gorm:"not null;default:1" json:"status"` // 0=禁用, 1=启用, 2=排空
	HealthCheckURL       string    `gorm:"size:500;not null;default:''" json:"health_check_url"`
	TimeoutSeconds       int       `gorm:"not null;default:300" json:"timeout_seconds"`
	StreamTimeoutSeconds int       `gorm:"not null;default:600" json:"stream_timeout_seconds"`
	ModelPatterns        string    `gorm:"type:text;not null;default:*" json:"model_patterns"` // 逗号分隔的模型匹配模式
	CreatedAt            time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (LLMBackend) TableName() string {
	return "llm_backends"
}

// LLM 后端状态常量
const (
	LLMBackendDisabled = 0
	LLMBackendEnabled  = 1
	LLMBackendDraining = 2
)
