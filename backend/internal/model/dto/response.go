package dto

import "time"

// LoginResponse represents login response.
type LoginResponse struct {
	ExpiresAt time.Time `json:"expires_at"`
	Token     string    `json:"token"`
	User      UserBrief `json:"user"`
}

// LoginErrorResponse represents login error response with lock info.
type LoginErrorResponse struct {
	Message       string `json:"message"`
	Code          int    `json:"code"`
	RemainingTime int64  `json:"remaining_time"`
	FailCount     int    `json:"fail_count"`
	MaxFailCount  int    `json:"max_fail_count"`
	Locked        bool   `json:"locked"`
}

// UserBrief represents brief user info.
type UserBrief struct {
	Department  *DeptBrief `json:"department,omitempty"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Role        string     `json:"role"`
	ID          int64      `json:"id"`
}

// UserDetail represents detailed user info.
type UserDetail struct {
	CreatedAt      time.Time  `json:"created_at"`
	Department     *DeptBrief `json:"department,omitempty"`
	LockedUntil    *time.Time `json:"locked_until"`
	Email          *string    `json:"email"`
	Phone          *string    `json:"phone"`
	AvatarURL      *string    `json:"avatar_url"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	DepartmentID   *int64     `json:"department_id"`
	Role           string     `json:"role"`
	DisplayName    string     `json:"display_name"`
	Username       string     `json:"username"`
	ID             int64      `json:"id"`
	LoginFailCount int        `json:"login_fail_count"`
	Status         int16      `json:"status"`
}

// LoginLockStatusResponse represents login lock status.
type LoginLockStatusResponse struct {
	LockedUntil    *time.Time `json:"locked_until"`
	LoginFailCount int        `json:"login_fail_count"`
	RemainingTime  int64      `json:"remaining_time"`
	Locked         bool       `json:"locked"`
}

// DeptBrief represents brief department info.
type DeptBrief struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

// DeptTree represents department tree structure.
type DeptTree struct {
	Description *string    `json:"description"`
	Manager     *UserBrief `json:"manager"`
	Name        string     `json:"name"`
	Children    []DeptTree `json:"children"`
	ID          int64      `json:"id"`
	UserCount   int        `json:"user_count"`
	Status      int16      `json:"status"`
}

// APIKeyResponse represents API key list item.
type APIKeyResponse struct {
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	ID         int64      `json:"id"`
	Status     int16      `json:"status"`
}

// APIKeyCreateResponse represents create API key response with full key.
type APIKeyCreateResponse struct {
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	KeyPrefix string     `json:"key_prefix"`
	ID        int64      `json:"id"`
}

// APIKeyCopyResponse represents copy API key response.
type APIKeyCopyResponse struct {
	Key string `json:"key"`
}

// StatsOverview represents usage overview.
type StatsOverview struct {
	SystemStatus string      `json:"system_status"`
	Today        PeriodStats `json:"today"`
	ThisMonth    PeriodStats `json:"this_month"`
	TotalUsers   int64       `json:"total_users"`
	TotalDepts   int64       `json:"total_departments"`
	TotalAPIKeys int64       `json:"total_api_keys"`
}

// PeriodStats represents statistics for a time period.
type PeriodStats struct {
	TotalTokens             int64 `json:"total_tokens"`
	TotalRequests           int64 `json:"total_requests"`
	ActiveUsers             int64 `json:"active_users"`
	ThirdPartyTotalTokens   int64 `json:"third_party_total_tokens"`
	ThirdPartyTotalRequests int64 `json:"third_party_total_requests"`
}

// UsageItem represents usage statistics item.
type UsageItem struct {
	Date                               string `json:"date"`
	PromptTokens                       int64  `json:"prompt_tokens"`
	CompletionTokens                   int64  `json:"completion_tokens"`
	TotalTokens                        int64  `json:"total_tokens"`
	RequestCount                       int64  `json:"request_count"`
	CacheCreationInputTokens           int64  `json:"cache_creation_input_tokens"`
	CacheReadInputTokens               int64  `json:"cache_read_input_tokens"`
	ThirdPartyPromptTokens             int64  `json:"third_party_prompt_tokens"`
	ThirdPartyCompletionTokens         int64  `json:"third_party_completion_tokens"`
	ThirdPartyTotalTokens              int64  `json:"third_party_total_tokens"`
	ThirdPartyRequestCount             int64  `json:"third_party_request_count"`
	ThirdPartyCacheCreationInputTokens int64  `json:"third_party_cache_creation_input_tokens"`
	ThirdPartyCacheReadInputTokens     int64  `json:"third_party_cache_read_input_tokens"`
}

// UsageResponse represents usage statistics response.
type UsageResponse struct {
	Period string      `json:"period"`
	Items  []UsageItem `json:"items"`
}

// UsageExportItem represents usage export data item.
type UsageExportItem struct {
	Date             string `json:"date"`
	Username         string `json:"username"`
	Department       string `json:"department"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
	RequestCount     int64  `json:"request_count"`
}

// RankingItem represents ranking item.
type RankingItem struct {
	Name         string `json:"name"`
	Rank         int    `json:"rank"`
	ID           int64  `json:"id"`
	TotalTokens  int64  `json:"total_tokens"`
	RequestCount int64  `json:"request_count"`
}

// KeyUsageItem represents API key usage statistics item.
type KeyUsageItem struct {
	Name             string `json:"name"`
	ID               int64  `json:"id"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
	RequestCount     int64  `json:"request_count"`
}

// MyLimitResponse represents current user's limit info (legacy API).
type MyLimitResponse struct {
	Limits      map[string]LimitDetail `json:"limits"`
	Concurrency ConcurrencyInfo        `json:"concurrency"`
}

// LimitDetail represents limit details.
type LimitDetail struct {
	MaxTokens       int64 `json:"max_tokens"`
	UsedTokens      int64 `json:"used_tokens"`
	RemainingTokens int64 `json:"remaining_tokens"`
	UsagePercent    int   `json:"usage_percent"`
}

// ConcurrencyInfo represents concurrency info.
type ConcurrencyInfo struct {
	Max     int `json:"max"`
	Current int `json:"current"`
}

// LimitProgressResponse represents limit progress response with reset times.
type LimitProgressResponse struct {
	Limits      []LimitProgressItem `json:"limits"`
	Concurrency ConcurrencyInfo     `json:"concurrency"`
	AnyExceeded bool                `json:"any_exceeded"`
}

// LimitProgressItem represents progress info for a single limit rule.
type LimitProgressItem struct {
	CycleStartAt    *int64   `json:"cycle_start_at"`
	ResetAt         *int64   `json:"reset_at"`
	ResetInHours    *float64 `json:"reset_in_hours"`
	Period          string   `json:"period"`
	RuleID          int64    `json:"rule_id"`
	PeriodHours     int      `json:"period_hours"`
	MaxTokens       int64    `json:"max_tokens"`
	UsedTokens      int64    `json:"used_tokens"`
	RemainingTokens int64    `json:"remaining_tokens"`
	UsagePercent    int      `json:"usage_percent"`
	Exceeded        bool     `json:"exceeded"`
}

// LLMBackendResponse represents LLM backend info.
type LLMBackendResponse struct {
	HealthCheckURL       string `json:"health_check_url"`
	Name                 string `json:"name"`
	DisplayName          string `json:"display_name"`
	BaseURL              string `json:"base_url"`
	UpdatedAt            string `json:"updated_at"`
	Format               string `json:"format"`
	CreatedAt            string `json:"created_at"`
	ModelPatterns        string `json:"model_patterns"`
	TimeoutSeconds       int    `json:"timeout_seconds"`
	ActiveConnections    int    `json:"active_connections"`
	ID                   int64  `json:"id"`
	StreamTimeoutSeconds int    `json:"stream_timeout_seconds"`
	MaxConcurrency       int    `json:"max_concurrency"`
	Weight               int    `json:"weight"`
	Status               int16  `json:"status"`
	Healthy              bool   `json:"healthy"`
	HasAPIKey            bool   `json:"has_api_key"`
}

// PlatformModelInfo represents platform model info for users.
type PlatformModelInfo struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Format        string `json:"format"`
	ModelPatterns string `json:"model_patterns"`
	Status        int16  `json:"status"`
}

// MCPServiceResponse represents MCP service info.
type MCPServiceResponse struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description"`
	EndpointURL   string `json:"endpoint_url"`
	TransportType string `json:"transport_type"`
	Status        string `json:"status"`
	AuthType      string `json:"auth_type"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	ID            int64  `json:"id"`
	ToolsCount    int    `json:"tools_count"`
	Connected     bool   `json:"connected"`
}

// MCPToolInfo represents MCP tool brief info.
type MCPToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ServiceName string `json:"service_name"`
}

// MCPAccessRuleResponse represents MCP access rule response.
type MCPAccessRuleResponse struct {
	ServiceName string `json:"service_name"`
	TargetType  string `json:"target_type"`
	TargetName  string `json:"target_name"`
	ID          int64  `json:"id"`
	ServiceID   int64  `json:"service_id"`
	TargetID    int64  `json:"target_id"`
	Allowed     bool   `json:"allowed"`
}
