package model

import "time"

// TokenUsage 单次请求 Token 用量记录
type TokenUsage struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64     `gorm:"not null;index:idx_token_usage_user_created" json:"user_id"`
	APIKeyID         int64     `gorm:"not null;index" json:"api_key_id"`
	Model            string    `gorm:"size:100;not null;index" json:"model"`
	PromptTokens     int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int       `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens      int       `gorm:"not null;default:0" json:"total_tokens"`
	RequestType      string    `gorm:"size:30;not null" json:"request_type"` // chat_completion | completion
	DurationMs       *int      `json:"duration_ms"`
	CreatedAt        time.Time `gorm:"not null;autoCreateTime;index:idx_token_usage_user_created;index" json:"created_at"`
}

// TableName 指定表名
func (TokenUsage) TableName() string {
	return "token_usage"
}

// TokenUsageDaily 每日用量汇总
type TokenUsageDaily struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64     `gorm:"not null;uniqueIndex:idx_token_usage_daily_user_date" json:"user_id"`
	UsageDate        time.Time `gorm:"type:date;not null;uniqueIndex:idx_token_usage_daily_user_date;index" json:"usage_date"`
	PromptTokens     int64     `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int64     `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens      int64     `gorm:"not null;default:0" json:"total_tokens"`
	RequestCount     int       `gorm:"not null;default:0" json:"request_count"`
	CreatedAt        time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (TokenUsageDaily) TableName() string {
	return "token_usage_daily"
}

// RequestLog LLM 请求日志
type RequestLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"not null;index:idx_request_logs_user_created" json:"user_id"`
	APIKeyID     int64     `gorm:"not null" json:"api_key_id"`
	RequestType  string    `gorm:"size:30;not null" json:"request_type"`
	Model        *string   `gorm:"size:100" json:"model"`
	StatusCode   int       `gorm:"not null" json:"status_code"`
	ErrorMessage *string   `gorm:"type:text" json:"error_message"`
	ClientIP     *string   `gorm:"size:45" json:"client_ip"`
	UserAgent    *string   `gorm:"size:500" json:"user_agent"`
	DurationMs   *int      `json:"duration_ms"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime;index:idx_request_logs_user_created;index" json:"created_at"`
}

// TableName 指定表名
func (RequestLog) TableName() string {
	return "request_logs"
}
