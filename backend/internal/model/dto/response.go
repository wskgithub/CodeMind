package dto

import "time"

// ──────────────────────────────────
// 认证相关响应
// ──────────────────────────────────

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserBrief    `json:"user"`
}

// ──────────────────────────────────
// 用户相关响应
// ──────────────────────────────────

// UserBrief 用户简要信息（登录响应、列表项中使用）
type UserBrief struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Role        string     `json:"role"`
	Department  *DeptBrief `json:"department,omitempty"`
}

// UserDetail 用户详细信息
type UserDetail struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"display_name"`
	Email        *string    `json:"email"`
	Phone        *string    `json:"phone"`
	AvatarURL    *string    `json:"avatar_url"`
	Role         string     `json:"role"`
	DepartmentID *int64     `json:"department_id"`
	Department   *DeptBrief `json:"department,omitempty"`
	Status       int16      `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ──────────────────────────────────
// 部门相关响应
// ──────────────────────────────────

// DeptBrief 部门简要信息
type DeptBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// DeptTree 部门树形结构
type DeptTree struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Manager     *UserBrief `json:"manager"`
	UserCount   int        `json:"user_count"`
	Status      int16      `json:"status"`
	Children    []DeptTree `json:"children"`
}

// ──────────────────────────────────
// API Key 相关响应
// ──────────────────────────────────

// APIKeyResponse API Key 列表项
type APIKeyResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Status     int16      `json:"status"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// APIKeyCreateResponse 创建 API Key 的响应（包含完整 Key，仅此一次）
type APIKeyCreateResponse struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`        // 完整 Key，仅创建时返回
	KeyPrefix string     `json:"key_prefix"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// ──────────────────────────────────
// 统计相关响应
// ──────────────────────────────────

// StatsOverview 用量总览
type StatsOverview struct {
	Today          PeriodStats `json:"today"`
	ThisMonth      PeriodStats `json:"this_month"`
	TotalUsers     int64       `json:"total_users"`
	TotalDepts     int64       `json:"total_departments"`
	TotalAPIKeys   int64       `json:"total_api_keys"`
	SystemStatus   string      `json:"system_status"`
}

// PeriodStats 某个时间段的统计
type PeriodStats struct {
	TotalTokens   int64 `json:"total_tokens"`
	TotalRequests int64 `json:"total_requests"`
	ActiveUsers   int64 `json:"active_users"`
}

// UsageItem 用量统计项
type UsageItem struct {
	Date             string `json:"date"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
	RequestCount     int64  `json:"request_count"`
}

// UsageResponse 用量统计响应
type UsageResponse struct {
	Period string      `json:"period"`
	Items  []UsageItem `json:"items"`
}

// RankingItem 排行榜项
type RankingItem struct {
	Rank        int    `json:"rank"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	TotalTokens int64  `json:"total_tokens"`
	RequestCount int64 `json:"request_count"`
}

// ──────────────────────────────────
// 限额相关响应
// ──────────────────────────────────

// MyLimitResponse 当前用户限额信息
type MyLimitResponse struct {
	Limits      map[string]LimitDetail `json:"limits"`
	Concurrency ConcurrencyInfo        `json:"concurrency"`
}

// LimitDetail 限额详情
type LimitDetail struct {
	MaxTokens       int64 `json:"max_tokens"`
	UsedTokens      int64 `json:"used_tokens"`
	RemainingTokens int64 `json:"remaining_tokens"`
	UsagePercent    int   `json:"usage_percent"`
}

// ConcurrencyInfo 并发信息
type ConcurrencyInfo struct {
	Max     int `json:"max"`
	Current int `json:"current"`
}

// ──────────────────────────────────
// MCP 服务管理响应
// ──────────────────────────────────

// MCPServiceResponse MCP 服务信息响应
type MCPServiceResponse struct {
	ID            int64       `json:"id"`
	Name          string      `json:"name"`
	DisplayName   string      `json:"display_name"`
	Description   string      `json:"description"`
	EndpointURL   string      `json:"endpoint_url"`
	TransportType string      `json:"transport_type"`
	Status        string      `json:"status"`
	AuthType      string      `json:"auth_type"`
	ToolsCount    int         `json:"tools_count"`    // 工具数量
	Connected     bool        `json:"connected"`      // 是否已连接
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
}

// MCPToolInfo MCP 工具简要信息
type MCPToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ServiceName string `json:"service_name"` // 所属服务
}

// MCPAccessRuleResponse MCP 访问规则响应
type MCPAccessRuleResponse struct {
	ID          int64  `json:"id"`
	ServiceID   int64  `json:"service_id"`
	ServiceName string `json:"service_name"`
	TargetType  string `json:"target_type"`
	TargetID    int64  `json:"target_id"`
	TargetName  string `json:"target_name"` // 用户名/部门名/角色名
	Allowed     bool   `json:"allowed"`
}
