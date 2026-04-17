package model

import "time"

// SystemConfig represents a system configuration entry.
type SystemConfig struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey   string    `gorm:"size:100;not null;uniqueIndex" json:"config_key"`
	ConfigValue string    `gorm:"type:text;not null" json:"config_value"`
	Description *string   `gorm:"size:500" json:"description"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name.
func (SystemConfig) TableName() string {
	return "system_configs"
}

const (
	ConfigLLMBaseURL         = "llm.base_url"
	ConfigLLMAPIKey          = "llm.api_key"
	ConfigLLMModels          = "llm.models"
	ConfigLLMDefaultModel    = "llm.default_model"
	ConfigMaxKeysPerUser     = "system.max_keys_per_user"
	ConfigDefaultConcurrency     = "system.default_concurrency"
	ConfigForceChangePwd         = "system.force_change_password"
	ConfigTrainingDataCollection = "system.training_data_collection"
	ConfigPlatformServiceURL     = "platform.service_url"
	ConfigTrainingSanitizeEnabled       = "training.sanitize_enabled"
	ConfigTrainingSanitizePatterns      = "training.sanitize_patterns"
	ConfigTrainingDedupEnabled          = "training.dedup_enabled"
	ConfigTrainingQualityScoringEnabled = "training.quality_scoring_enabled"
)
