package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringSlice JSON 字符串数组类型，映射 PostgreSQL JSONB
type StringSlice []string

// Value 实现 driver.Valuer 接口，写入数据库
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口，从数据库读取
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
		return errors.New("StringSlice: 无法扫描非 []byte/string 类型")
	}
	return json.Unmarshal(data, s)
}

// ──────────────────────────────────
// 第三方模型服务模板
// ──────────────────────────────────

// ThirdPartyProviderTemplate 管理员配置的第三方服务模板
type ThirdPartyProviderTemplate struct {
	ID               int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name             string         `gorm:"size:100;not null" json:"name"`
	OpenAIBaseURL    string         `gorm:"column:openai_base_url;size:500" json:"openai_base_url"`
	AnthropicBaseURL string         `gorm:"column:anthropic_base_url;size:500" json:"anthropic_base_url"`
	Models           StringSlice    `gorm:"type:jsonb;not null;default:'[]'" json:"models"`
	Format           string         `gorm:"size:20;not null;default:openai" json:"format"`
	Description      *string        `gorm:"size:500" json:"description"`
	Icon             *string        `gorm:"size:100" json:"icon"`
	Status           int16          `gorm:"not null;default:1" json:"status"`
	SortOrder        int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedBy        int64          `gorm:"not null" json:"created_by"`
	CreatedAt        time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ThirdPartyProviderTemplate) TableName() string {
	return "third_party_provider_templates"
}

// ──────────────────────────────────
// 用户第三方模型服务
// ──────────────────────────────────

// UserThirdPartyProvider 用户绑定的第三方模型服务
type UserThirdPartyProvider struct {
	ID               int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64          `gorm:"not null;index" json:"user_id"`
	Name             string         `gorm:"size:100;not null" json:"name"`
	OpenAIBaseURL    string         `gorm:"column:openai_base_url;size:500" json:"openai_base_url"`
	AnthropicBaseURL string         `gorm:"column:anthropic_base_url;size:500" json:"anthropic_base_url"`
	APIKeyEncrypted  string         `gorm:"column:api_key_encrypted;type:text;not null" json:"-"` // 永不序列化
	Models           StringSlice    `gorm:"type:jsonb;not null;default:'[]'" json:"models"`
	Format           string         `gorm:"size:20;not null;default:openai" json:"format"`
	TemplateID       *int64         `json:"template_id"`
	Status           int16          `gorm:"not null;default:1" json:"status"`
	CreatedAt        time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (UserThirdPartyProvider) TableName() string {
	return "user_third_party_providers"
}

// IsActive 服务是否启用
func (p *UserThirdPartyProvider) IsActive() bool {
	return p.Status == StatusEnabled
}

// ContainsModel 检查是否包含指定模型名称
func (p *UserThirdPartyProvider) ContainsModel(model string) bool {
	for _, m := range p.Models {
		if m == model {
			return true
		}
	}
	return false
}

// ──────────────────────────────────
// 第三方服务用量记录
// ──────────────────────────────────

// ThirdPartyTokenUsage 第三方模型服务用量（仅供参考）
type ThirdPartyTokenUsage struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64     `gorm:"not null;index:idx_tptu_user_created" json:"user_id"`
	ProviderID       int64     `gorm:"not null;index:idx_tptu_provider" json:"provider_id"`
	APIKeyID         int64     `gorm:"not null" json:"api_key_id"`
	Model            string    `gorm:"size:100;not null" json:"model"`
	PromptTokens     int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int       `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens      int       `gorm:"not null;default:0" json:"total_tokens"`
	RequestType      string    `gorm:"size:30;not null" json:"request_type"`
	DurationMs       *int      `json:"duration_ms"`
	CreatedAt        time.Time `gorm:"not null;autoCreateTime;index:idx_tptu_user_created" json:"created_at"`
}

// TableName 指定表名
func (ThirdPartyTokenUsage) TableName() string {
	return "third_party_token_usage"
}

// ──────────────────────────────────
// 第三方服务路由信息（运行时，非持久化）
// ──────────────────────────────────

// ThirdPartyRouteInfo 第三方模型路由信息（从缓存/DB 解析后传递到代理层）
type ThirdPartyRouteInfo struct {
	ProviderID       int64  `json:"provider_id"`
	ProviderName     string `json:"provider_name"`
	OpenAIBaseURL    string `json:"openai_base_url"`
	AnthropicBaseURL string `json:"anthropic_base_url"`
	APIKeyEncrypted  string `json:"api_key_encrypted"`
	Format           string `json:"format"`
}

// BaseURLForFormat 根据请求协议格式返回对应的 Base URL
func (r *ThirdPartyRouteInfo) BaseURLForFormat(requestFormat string) string {
	if requestFormat == "anthropic" {
		return r.AnthropicBaseURL
	}
	return r.OpenAIBaseURL
}

// IsFormatCompatible 检查服务商协议格式是否兼容当前请求格式
// providerFormat: openai / anthropic / all
// requestFormat:  openai / anthropic
func (r *ThirdPartyRouteInfo) IsFormatCompatible(requestFormat string) bool {
	return r.Format == "all" || r.Format == requestFormat
}
