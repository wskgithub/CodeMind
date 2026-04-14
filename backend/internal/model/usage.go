package model

import (
	"encoding/json"
	"time"
)

// TokenUsage 单次请求 Token 用量记录
type TokenUsage struct {
	ID                       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID                   int64     `gorm:"not null;index:idx_token_usage_user_created" json:"user_id"`
	APIKeyID                 int64     `gorm:"index" json:"api_key_id"`
	Model                    string    `gorm:"size:100;not null;index" json:"model"`
	PromptTokens             int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens         int       `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens              int       `gorm:"not null;default:0" json:"total_tokens"`
	CacheCreationInputTokens int       `gorm:"not null;default:0" json:"cache_creation_input_tokens"` // 缓存创建 Token 数
	CacheReadInputTokens     int       `gorm:"not null;default:0" json:"cache_read_input_tokens"`     // 缓存命中 Token 数
	RequestType              string    `gorm:"size:30;not null" json:"request_type"`                  // chat_completion | completion
	DurationMs               *int      `json:"duration_ms"`
	CreatedAt                time.Time `gorm:"not null;autoCreateTime;index:idx_token_usage_user_created;index" json:"created_at"`
}

// TableName 指定表名
func (TokenUsage) TableName() string {
	return "token_usage"
}

// TokenUsageDaily 每日用量汇总
type TokenUsageDaily struct {
	ID                       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID                   int64     `gorm:"not null;uniqueIndex:idx_token_usage_daily_user_date" json:"user_id"`
	UsageDate                time.Time `gorm:"type:date;not null;uniqueIndex:idx_token_usage_daily_user_date;index" json:"usage_date"`
	PromptTokens             int64     `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens         int64     `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens              int64     `gorm:"not null;default:0" json:"total_tokens"`
	CacheCreationInputTokens int64     `gorm:"not null;default:0" json:"cache_creation_input_tokens"` // 缓存创建 Token 数
	CacheReadInputTokens     int64     `gorm:"not null;default:0" json:"cache_read_input_tokens"`     // 缓存命中 Token 数
	RequestCount             int       `gorm:"not null;default:0" json:"request_count"`
	CreatedAt                time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (TokenUsageDaily) TableName() string {
	return "token_usage_daily"
}

// TokenUsageDailyKey Key 级每日用量汇总
type TokenUsageDailyKey struct {
	ID                       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	APIKeyID                 int64     `gorm:"not null;uniqueIndex:idx_token_usage_daily_key_key_date" json:"api_key_id"`
	UserID                   int64     `gorm:"not null;index:idx_token_usage_daily_key_user_date" json:"user_id"`
	UsageDate                time.Time `gorm:"type:date;not null;uniqueIndex:idx_token_usage_daily_key_key_date" json:"usage_date"`
	PromptTokens             int64     `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens         int64     `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens              int64     `gorm:"not null;default:0" json:"total_tokens"`
	CacheCreationInputTokens int64     `gorm:"not null;default:0" json:"cache_creation_input_tokens"` // 缓存创建 Token 数
	CacheReadInputTokens     int64     `gorm:"not null;default:0" json:"cache_read_input_tokens"`     // 缓存命中 Token 数
	RequestCount             int       `gorm:"not null;default:0" json:"request_count"`
	CreatedAt                time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (TokenUsageDailyKey) TableName() string {
	return "token_usage_daily_key"
}

// RequestLog LLM 请求日志
type RequestLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"not null;index:idx_request_logs_user_created" json:"user_id"`
	APIKeyID     int64     `json:"api_key_id"`
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

// LLMTrainingData LLM 训练数据记录
// 存储完整的请求/响应 JSON，用于后续模型训练
type LLMTrainingData struct {
	ID                   int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID               int64           `gorm:"not null;index:idx_training_data_user_created" json:"user_id"`
	APIKeyID             int64           `json:"api_key_id"`
	RequestType          string          `gorm:"size:30;not null;index:idx_training_data_type" json:"request_type"`
	Model                string          `gorm:"size:100;not null;index:idx_training_data_model" json:"model"`
	IsStream             bool            `gorm:"not null;default:false" json:"is_stream"`
	RequestBody          json.RawMessage `gorm:"type:jsonb;not null" json:"request_body"`
	ResponseBody         json.RawMessage `gorm:"type:jsonb" json:"response_body"`
	PromptTokens         int             `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens     int             `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens          int             `gorm:"not null;default:0" json:"total_tokens"`
	DurationMs           *int            `json:"duration_ms"`
	StatusCode           int             `gorm:"not null;default:200" json:"status_code"`
	ClientIP             *string         `gorm:"size:45" json:"client_ip"`
	IsExcluded           bool            `gorm:"not null;default:false" json:"is_excluded"`
	Source               string          `gorm:"size:20;not null;default:platform;index:idx_training_data_source" json:"source"`    // platform | third_party
	ThirdPartyProviderID *int64          `json:"third_party_provider_id,omitempty"`
	// 增强字段
	IsSanitized    bool    `gorm:"not null;default:false" json:"is_sanitized"`                            // 是否已脱敏
	ConversationID *string `gorm:"size:64;index:idx_training_data_conversation" json:"conversation_id"`  // 会话ID
	ContentHash    *string `gorm:"size:64;index:idx_training_data_content_hash" json:"content_hash"`     // 内容哈希（去重）
	QualityScore   *int    `gorm:"type:smallint;index:idx_training_data_quality" json:"quality_score"`    // 质量评分 0-100
	CreatedAt      time.Time `gorm:"not null;autoCreateTime;index:idx_training_data_user_created;index:idx_training_data_created" json:"created_at"`
}

// TableName 指定表名
func (LLMTrainingData) TableName() string {
	return "llm_training_data"
}

// LLMTrainingDataListItem 训练数据列表项（不含大字段，用于列表查询）
type LLMTrainingDataListItem struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	Username         string    `json:"username"`
	RequestType      string    `json:"request_type"`
	Model            string    `json:"model"`
	IsStream         bool      `json:"is_stream"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	DurationMs       *int      `json:"duration_ms"`
	StatusCode       int       `json:"status_code"`
	IsExcluded       bool      `json:"is_excluded"`
	// 增强字段
	IsSanitized    bool    `json:"is_sanitized"`
	ConversationID *string `json:"conversation_id,omitempty"`
	ContentHash    *string `json:"content_hash,omitempty"`
	QualityScore   *int    `json:"quality_score,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// TrainingDataStats 训练数据统计信息
type TrainingDataStats struct {
	TotalCount     int64                     `json:"total_count"`
	TodayCount     int64                     `json:"today_count"`
	ExcludedCount  int64                     `json:"excluded_count"`
	ModelDistribution []ModelDistributionItem `json:"model_distribution"`
}

// ModelDistributionItem 模型分布统计项
type ModelDistributionItem struct {
	Model string `json:"model"`
	Count int64  `json:"count"`
}
