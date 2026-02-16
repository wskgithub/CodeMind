package model

import "time"

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey   string    `gorm:"size:100;not null;uniqueIndex" json:"config_key"`
	ConfigValue string    `gorm:"type:text;not null" json:"config_value"` // JSON 格式
	Description *string   `gorm:"size:500" json:"description"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// 预定义配置键
const (
	ConfigLLMBaseURL         = "llm.base_url"
	ConfigLLMAPIKey          = "llm.api_key"
	ConfigLLMModels          = "llm.models"
	ConfigLLMDefaultModel    = "llm.default_model"
	ConfigMaxKeysPerUser     = "system.max_keys_per_user"
	ConfigDefaultConcurrency = "system.default_concurrency"
	ConfigForceChangePwd     = "system.force_change_password"
)
