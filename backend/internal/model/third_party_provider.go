package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringSlice represents a JSON string array that maps to PostgreSQL JSONB.
type StringSlice []string

// Value implements driver.Valuer for database writes.
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database reads.
func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = StringSlice{}
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("StringSlice: cannot scan non []byte/string type")
	}
	return json.Unmarshal(data, s)
}

// ThirdPartyProviderTemplate represents an admin-configured third-party service template.
type ThirdPartyProviderTemplate struct {
	UpdatedAt        time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	CreatedAt        time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	Description      *string        `gorm:"size:500" json:"description"`
	Icon             *string        `gorm:"size:100" json:"icon"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	AnthropicBaseURL string         `gorm:"column:anthropic_base_url;size:500" json:"anthropic_base_url"`
	Format           string         `gorm:"size:20;not null;default:openai" json:"format"`
	OpenAIBaseURL    string         `gorm:"column:openai_base_url;size:500" json:"openai_base_url"`
	Name             string         `gorm:"size:100;not null" json:"name"`
	Models           StringSlice    `gorm:"type:jsonb;not null;default:'[]'" json:"models"`
	SortOrder        int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedBy        int64          `gorm:"not null" json:"created_by"`
	ID               int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Status           int16          `gorm:"not null;default:1" json:"status"`
}

// TableName returns the table name.
func (ThirdPartyProviderTemplate) TableName() string {
	return "third_party_provider_templates"
}

// UserThirdPartyProvider represents a user's bound third-party model service.
type UserThirdPartyProvider struct {
	CreatedAt        time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	TemplateID       *int64         `json:"template_id"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Name             string         `gorm:"size:100;not null" json:"name"`
	OpenAIBaseURL    string         `gorm:"column:openai_base_url;size:500" json:"openai_base_url"`
	AnthropicBaseURL string         `gorm:"column:anthropic_base_url;size:500" json:"anthropic_base_url"`
	APIKeyEncrypted  string         `gorm:"column:api_key_encrypted;type:text;not null" json:"-"`
	Format           string         `gorm:"size:20;not null;default:openai" json:"format"`
	Models           StringSlice    `gorm:"type:jsonb;not null;default:'[]'" json:"models"`
	ID               int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64          `gorm:"not null;index" json:"user_id"`
	Status           int16          `gorm:"not null;default:1" json:"status"`
}

// TableName returns the table name.
func (UserThirdPartyProvider) TableName() string {
	return "user_third_party_providers"
}

// IsActive returns whether the service is enabled.
func (p *UserThirdPartyProvider) IsActive() bool {
	return p.Status == StatusEnabled
}

// ContainsModel checks if the service includes the specified model.
func (p *UserThirdPartyProvider) ContainsModel(model string) bool {
	for _, m := range p.Models {
		if m == model {
			return true
		}
	}
	return false
}

// ThirdPartyTokenUsage represents third-party model service usage.
type ThirdPartyTokenUsage struct {
	CreatedAt                time.Time `gorm:"not null;autoCreateTime;index:idx_tptu_user_created" json:"created_at"`
	DurationMs               *int      `json:"duration_ms"`
	Model                    string    `gorm:"size:100;not null" json:"model"`
	RequestType              string    `gorm:"size:30;not null" json:"request_type"`
	CompletionTokens         int       `gorm:"not null;default:0" json:"completion_tokens"`
	PromptTokens             int       `gorm:"not null;default:0" json:"prompt_tokens"`
	ID                       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TotalTokens              int       `gorm:"not null;default:0" json:"total_tokens"`
	CacheCreationInputTokens int       `gorm:"not null;default:0" json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int       `gorm:"not null;default:0" json:"cache_read_input_tokens"`
	APIKeyID                 int64     `gorm:"not null" json:"api_key_id"`
	ProviderID               int64     `gorm:"not null;index:idx_tptu_provider" json:"provider_id"`
	UserID                   int64     `gorm:"not null;index:idx_tptu_user_created" json:"user_id"`
}

// TableName returns the table name.
func (ThirdPartyTokenUsage) TableName() string {
	return "third_party_token_usage"
}

// ThirdPartyRouteInfo contains third-party model routing information.
type ThirdPartyRouteInfo struct {
	ProviderName     string `json:"provider_name"`
	OpenAIBaseURL    string `json:"openai_base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url"`
	APIKeyEncrypted  string `json:"api_key_encrypted"`
	Format           string `json:"format"`
	ProviderID       int64  `json:"provider_id"`
}

// BaseURLForFormat returns the base URL for the specified request format.
func (r *ThirdPartyRouteInfo) BaseURLForFormat(requestFormat string) string {
	if requestFormat == "anthropic" {
		return r.AnthropicBaseURL
	}
	return r.OpenAIBaseURL
}

// IsFormatCompatible checks if the provider format is compatible with the request format.
func (r *ThirdPartyRouteInfo) IsFormatCompatible(requestFormat string) bool {
	return r.Format == "all" || r.Format == requestFormat
}
