package dto

import "time"

// ──────────────────────────────────
// 认证相关请求
// ──────────────────────────────────

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,min=1,max=100"`
	Email       *string `json:"email" binding:"omitempty,email,max=255"`
	Phone       *string `json:"phone" binding:"omitempty,max=20"`
}

// ──────────────────────────────────
// 用户管理请求
// ──────────────────────────────────

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username     string `json:"username" binding:"required,min=2,max=50"`
	Password     string `json:"password" binding:"required,min=8,max=128"`
	DisplayName  string `json:"display_name" binding:"required,min=1,max=100"`
	Email        string `json:"email" binding:"omitempty,email,max=255"`
	Phone        string `json:"phone" binding:"omitempty,max=20"`
	Role         string `json:"role" binding:"required,oneof=super_admin dept_manager user"`
	DepartmentID *int64 `json:"department_id"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	DisplayName  *string `json:"display_name" binding:"omitempty,min=1,max=100"`
	Email        *string `json:"email" binding:"omitempty,email,max=255"`
	Phone        *string `json:"phone" binding:"omitempty,max=20"`
	Role         *string `json:"role" binding:"omitempty,oneof=super_admin dept_manager user"`
	DepartmentID *int64  `json:"department_id"`
	Status       *int16  `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateStatusRequest 切换状态请求
type UpdateStatusRequest struct {
	Status int16 `json:"status" binding:"oneof=0 1"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}

// UserListQuery 用户列表查询参数
type UserListQuery struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword      string `form:"keyword"`
	DepartmentID *int64 `form:"department_id"`
	Role         string `form:"role" binding:"omitempty,oneof=super_admin dept_manager user"`
	Status       *int16 `form:"status" binding:"omitempty,oneof=0 1"`
}

// GetPage 获取页码（默认值处理）
func (q *UserListQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetPageSize 获取每页数量（默认值处理）
func (q *UserListQuery) GetPageSize() int {
	if q.PageSize <= 0 {
		return 20
	}
	if q.PageSize > 100 {
		return 100
	}
	return q.PageSize
}

// ──────────────────────────────────
// 部门管理请求
// ──────────────────────────────────

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description"`
	ParentID    *int64  `json:"parent_id"`
	ManagerID   *int64  `json:"manager_id"`
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description"`
	ParentID    *int64  `json:"parent_id"`
	ManagerID   *int64  `json:"manager_id"`
}

// ──────────────────────────────────
// API Key 管理请求
// ──────────────────────────────────

// CreateAPIKeyRequest 创建 API Key 请求
type CreateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required,min=1,max=100"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// ──────────────────────────────────
// 用量统计请求
// ──────────────────────────────────

// StatsQuery 统计查询参数
type StatsQuery struct {
	Period       string `form:"period" binding:"required,oneof=daily weekly monthly"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
	UserID       *int64 `form:"user_id"`
	DepartmentID *int64 `form:"department_id"`
}

// RankingQuery 排行榜查询参数
type RankingQuery struct {
	Type   string `form:"type" binding:"required,oneof=user department"`
	Period string `form:"period" binding:"required,oneof=daily weekly monthly"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GetLimit 获取排行数量（默认值处理）
func (q *RankingQuery) GetLimit() int {
	if q.Limit <= 0 {
		return 10
	}
	return q.Limit
}

// KeyUsageQuery API Key 用量查询参数
type KeyUsageQuery struct {
	Period    string `form:"period" binding:"omitempty,oneof=daily weekly monthly"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// ──────────────────────────────────
// 限额管理请求
// ──────────────────────────────────

// UpsertRateLimitRequest 创建/更新限额配置请求
// Period 为显示标签，PeriodHours 为实际周期时长
type UpsertRateLimitRequest struct {
	TargetType     string `json:"target_type" binding:"required,oneof=global department user"`
	TargetID       int64  `json:"target_id"`
	Period         string `json:"period" binding:"required,oneof=daily weekly monthly custom"`
	PeriodHours    int    `json:"period_hours" binding:"omitempty,min=1"`
	MaxTokens      int64  `json:"max_tokens" binding:"required,min=0"`
	MaxRequests    int    `json:"max_requests" binding:"omitempty,min=0"`
	MaxConcurrency int    `json:"max_concurrency" binding:"omitempty,min=1"`
	AlertThreshold int16  `json:"alert_threshold" binding:"omitempty,min=0,max=100"`
}

// LimitListQuery 限额列表查询参数
type LimitListQuery struct {
	TargetType string `form:"target_type" binding:"omitempty,oneof=global department user"`
	TargetID   *int64 `form:"target_id"`
}

// ──────────────────────────────────
// LLM 后端管理请求
// ──────────────────────────────────

// CreateLLMBackendRequest 创建 LLM 后端请求
type CreateLLMBackendRequest struct {
	Name                 string `json:"name" binding:"required,min=2,max=100"`
	DisplayName          string `json:"display_name" binding:"omitempty,max=200"`
	BaseURL              string `json:"base_url" binding:"required,url,max=500"`
	APIKey               string `json:"api_key"`
	Format               string `json:"format" binding:"required,oneof=openai anthropic"`
	Weight               int    `json:"weight" binding:"omitempty,min=1,max=10000"`
	MaxConcurrency       int    `json:"max_concurrency" binding:"omitempty,min=1"`
	HealthCheckURL       string `json:"health_check_url" binding:"omitempty,max=500"`
	TimeoutSeconds       int    `json:"timeout_seconds" binding:"omitempty,min=1"`
	StreamTimeoutSeconds int    `json:"stream_timeout_seconds" binding:"omitempty,min=1"`
	ModelPatterns        string `json:"model_patterns" binding:"omitempty"`
}

// UpdateLLMBackendRequest 更新 LLM 后端请求
type UpdateLLMBackendRequest struct {
	DisplayName          *string `json:"display_name" binding:"omitempty,max=200"`
	BaseURL              *string `json:"base_url" binding:"omitempty,url,max=500"`
	APIKey               *string `json:"api_key"`
	Format               *string `json:"format" binding:"omitempty,oneof=openai anthropic"`
	Weight               *int    `json:"weight" binding:"omitempty,min=1,max=10000"`
	MaxConcurrency       *int    `json:"max_concurrency" binding:"omitempty,min=1"`
	Status               *int16  `json:"status" binding:"omitempty,oneof=0 1 2"`
	HealthCheckURL       *string `json:"health_check_url" binding:"omitempty,max=500"`
	TimeoutSeconds       *int    `json:"timeout_seconds" binding:"omitempty,min=1"`
	StreamTimeoutSeconds *int    `json:"stream_timeout_seconds" binding:"omitempty,min=1"`
	ModelPatterns        *string `json:"model_patterns"`
}

// ──────────────────────────────────
// 系统管理请求
// ──────────────────────────────────

// UpdateConfigsRequest 更新系统配置请求
type UpdateConfigsRequest struct {
	Configs []ConfigItem `json:"configs" binding:"required,dive"`
}

// ConfigItem 单个配置项
type ConfigItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// CreateAnnouncementRequest 创建公告请求
type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required"`
	Pinned  bool   `json:"pinned"`
	Status  int16  `json:"status" binding:"oneof=0 1"`
}

// UpdateAnnouncementRequest 更新公告请求
type UpdateAnnouncementRequest struct {
	Title   *string `json:"title" binding:"omitempty,min=1,max=200"`
	Content *string `json:"content"`
	Pinned  *bool   `json:"pinned"`
	Status  *int16  `json:"status" binding:"omitempty,oneof=0 1"`
}

// AuditLogQuery 审计日志查询参数
type AuditLogQuery struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Action     string `form:"action"`
	OperatorID *int64 `form:"operator_id"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
}

// GetPage 获取页码
func (q *AuditLogQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetPageSize 获取每页数量
func (q *AuditLogQuery) GetPageSize() int {
	if q.PageSize <= 0 {
		return 20
	}
	if q.PageSize > 100 {
		return 100
	}
	return q.PageSize
}

// ──────────────────────────────────
// MCP 服务管理请求
// ──────────────────────────────────

// CreateMCPServiceRequest 创建 MCP 服务请求
type CreateMCPServiceRequest struct {
	Name          string      `json:"name" binding:"required,min=2,max=100"`
	DisplayName   string      `json:"display_name" binding:"required,min=1,max=200"`
	Description   string      `json:"description"`
	EndpointURL   string      `json:"endpoint_url" binding:"required,url,max=500"`
	TransportType string      `json:"transport_type" binding:"required,oneof=sse streamable-http"`
	AuthType      string      `json:"auth_type" binding:"required,oneof=none bearer header"`
	AuthConfig    interface{} `json:"auth_config,omitempty"`
}

// UpdateMCPServiceRequest 更新 MCP 服务请求
type UpdateMCPServiceRequest struct {
	DisplayName   *string     `json:"display_name" binding:"omitempty,min=1,max=200"`
	Description   *string     `json:"description"`
	EndpointURL   *string     `json:"endpoint_url" binding:"omitempty,url,max=500"`
	TransportType *string     `json:"transport_type" binding:"omitempty,oneof=sse streamable-http"`
	Status        *string     `json:"status" binding:"omitempty,oneof=enabled disabled"`
	AuthType      *string     `json:"auth_type" binding:"omitempty,oneof=none bearer header"`
	AuthConfig    interface{} `json:"auth_config,omitempty"`
}

// SetMCPAccessRuleRequest 设置 MCP 访问规则请求
type SetMCPAccessRuleRequest struct {
	ServiceID  int64  `json:"service_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required,oneof=user department role"`
	TargetID   int64  `json:"target_id"`
	Allowed    bool   `json:"allowed"`
}
