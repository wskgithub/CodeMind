package dto

import "time"

// LoginRequest represents login request.
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// ChangePasswordRequest represents change password request.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}

// UpdateProfileRequest represents update profile request.
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,min=1,max=100"`
	Email       *string `json:"email" binding:"omitempty,email,max=255"`
	Phone       *string `json:"phone" binding:"omitempty,max=20"`
}

// CreateUserRequest represents create user request.
type CreateUserRequest struct {
	Username     string `json:"username" binding:"required,min=2,max=50"`
	Password     string `json:"password" binding:"required,min=8,max=128"`
	DisplayName  string `json:"display_name" binding:"required,min=1,max=100"`
	Email        string `json:"email" binding:"omitempty,email,max=255"`
	Phone        string `json:"phone" binding:"omitempty,max=20"`
	Role         string `json:"role" binding:"required,oneof=super_admin dept_manager user"`
	DepartmentID *int64 `json:"department_id"`
}

// UpdateUserRequest represents update user request.
type UpdateUserRequest struct {
	DisplayName  *string `json:"display_name" binding:"omitempty,min=1,max=100"`
	Email        *string `json:"email" binding:"omitempty,email,max=255"`
	Phone        *string `json:"phone" binding:"omitempty,max=20"`
	Role         *string `json:"role" binding:"omitempty,oneof=super_admin dept_manager user"`
	DepartmentID *int64  `json:"department_id"`
	Status       *int16  `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateStatusRequest represents status toggle request.
type UpdateStatusRequest struct {
	Status int16 `json:"status" binding:"oneof=0 1"`
}

// ResetPasswordRequest represents reset password request.
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}

// UnlockUserRequest represents unlock user request.
type UnlockUserRequest struct {
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// UserListQuery represents user list query parameters.
type UserListQuery struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword      string `form:"keyword"`
	DepartmentID *int64 `form:"department_id"`
	Role         string `form:"role" binding:"omitempty,oneof=super_admin dept_manager user"`
	Status       *int16 `form:"status" binding:"omitempty,oneof=0 1"`
}

// GetPage returns page number with default handling.
func (q *UserListQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetPageSize returns page size with default handling.
func (q *UserListQuery) GetPageSize() int {
	if q.PageSize <= 0 {
		return 20
	}
	if q.PageSize > 100 {
		return 100
	}
	return q.PageSize
}

// CreateDepartmentRequest represents create department request.
type CreateDepartmentRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description"`
	ParentID    *int64  `json:"parent_id"`
	ManagerID   *int64  `json:"manager_id"`
}

// UpdateDepartmentRequest represents update department request.
type UpdateDepartmentRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description"`
	ParentID    *int64  `json:"parent_id"`
	ManagerID   *int64  `json:"manager_id"`
}

// CreateAPIKeyRequest represents create API key request.
type CreateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required,min=1,max=100"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// StatsQuery represents statistics query parameters.
type StatsQuery struct {
	Period       string `form:"period" binding:"required,oneof=daily weekly monthly"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
	UserID       *int64 `form:"user_id"`
	DepartmentID *int64 `form:"department_id"`
}

// RankingQuery represents ranking query parameters.
type RankingQuery struct {
	Type   string `form:"type" binding:"required,oneof=user department"`
	Period string `form:"period" binding:"required,oneof=daily weekly monthly"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GetLimit returns ranking limit with default handling.
func (q *RankingQuery) GetLimit() int {
	if q.Limit <= 0 {
		return 10
	}
	return q.Limit
}

// KeyUsageQuery represents API key usage query parameters.
type KeyUsageQuery struct {
	Period    string `form:"period" binding:"omitempty,oneof=daily weekly monthly"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// UpsertRateLimitRequest represents create/update rate limit request.
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

// LimitListQuery represents limit list query parameters.
type LimitListQuery struct {
	TargetType string `form:"target_type" binding:"omitempty,oneof=global department user"`
	TargetID   *int64 `form:"target_id"`
}

// CreateLLMBackendRequest represents create LLM backend request.
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

// UpdateLLMBackendRequest represents update LLM backend request.
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

// UpdateConfigsRequest represents update system config request.
type UpdateConfigsRequest struct {
	Configs []ConfigItem `json:"configs" binding:"required,dive"`
}

// ConfigItem represents a single config item.
type ConfigItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// CreateAnnouncementRequest represents create announcement request.
type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required"`
	Pinned  bool   `json:"pinned"`
	Status  int16  `json:"status" binding:"oneof=0 1"`
}

// UpdateAnnouncementRequest represents update announcement request.
type UpdateAnnouncementRequest struct {
	Title   *string `json:"title" binding:"omitempty,min=1,max=200"`
	Content *string `json:"content"`
	Pinned  *bool   `json:"pinned"`
	Status  *int16  `json:"status" binding:"omitempty,oneof=0 1"`
}

// AuditLogQuery represents audit log query parameters.
type AuditLogQuery struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Action     string `form:"action"`
	OperatorID *int64 `form:"operator_id"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
}

// GetPage returns page number.
func (q *AuditLogQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetPageSize returns page size with default handling.
func (q *AuditLogQuery) GetPageSize() int {
	if q.PageSize <= 0 {
		return 20
	}
	if q.PageSize > 100 {
		return 100
	}
	return q.PageSize
}

// TrainingDataQuery represents training data query parameters.
type TrainingDataQuery struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Model       string `form:"model"`
	RequestType string `form:"request_type" binding:"omitempty,oneof=chat_completion completion embedding responses anthropic_messages"`
	UserID      *int64 `form:"user_id"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	IsExcluded  *bool  `form:"is_excluded"`
}

// GetPage returns page number.
func (q *TrainingDataQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetPageSize returns page size with default handling.
func (q *TrainingDataQuery) GetPageSize() int {
	if q.PageSize <= 0 {
		return 20
	}
	if q.PageSize > 100 {
		return 100
	}
	return q.PageSize
}

// TrainingDataExcludeRequest represents exclude/restore training data request.
type TrainingDataExcludeRequest struct {
	Excluded bool `json:"excluded"`
}

// TrainingDataExportQuery represents training data export query parameters.
type TrainingDataExportQuery struct {
	Model       string `form:"model"`
	RequestType string `form:"request_type" binding:"omitempty,oneof=chat_completion completion embedding responses anthropic_messages"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
}

// CreateMCPServiceRequest represents create MCP service request.
type CreateMCPServiceRequest struct {
	Name          string      `json:"name" binding:"required,min=2,max=100"`
	DisplayName   string      `json:"display_name" binding:"required,min=1,max=200"`
	Description   string      `json:"description"`
	EndpointURL   string      `json:"endpoint_url" binding:"required,url,max=500"`
	TransportType string      `json:"transport_type" binding:"required,oneof=sse streamable-http"`
	AuthType      string      `json:"auth_type" binding:"required,oneof=none bearer header"`
	AuthConfig    interface{} `json:"auth_config,omitempty"`
}

// UpdateMCPServiceRequest represents update MCP service request.
type UpdateMCPServiceRequest struct {
	DisplayName   *string     `json:"display_name" binding:"omitempty,min=1,max=200"`
	Description   *string     `json:"description"`
	EndpointURL   *string     `json:"endpoint_url" binding:"omitempty,url,max=500"`
	TransportType *string     `json:"transport_type" binding:"omitempty,oneof=sse streamable-http"`
	Status        *string     `json:"status" binding:"omitempty,oneof=enabled disabled"`
	AuthType      *string     `json:"auth_type" binding:"omitempty,oneof=none bearer header"`
	AuthConfig    interface{} `json:"auth_config,omitempty"`
}

// SetMCPAccessRuleRequest represents set MCP access rule request.
type SetMCPAccessRuleRequest struct {
	ServiceID  int64  `json:"service_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required,oneof=user department role"`
	TargetID   int64  `json:"target_id"`
	Allowed    bool   `json:"allowed"`
}

// CreateProviderTemplateRequest represents create provider template request.
type CreateProviderTemplateRequest struct {
	Name             string   `json:"name" binding:"required,min=1,max=100"`
	OpenAIBaseURL    string   `json:"openai_base_url" binding:"omitempty,max=500"`
	AnthropicBaseURL string   `json:"anthropic_base_url" binding:"omitempty,max=500"`
	Models           []string `json:"models" binding:"required,min=1"`
	Format           string   `json:"format" binding:"required,oneof=openai anthropic all"`
	Description      *string  `json:"description" binding:"omitempty,max=500"`
	Icon             *string  `json:"icon" binding:"omitempty,max=100"`
	SortOrder        int      `json:"sort_order"`
}

// UpdateProviderTemplateRequest represents update provider template request.
type UpdateProviderTemplateRequest struct {
	Name             *string  `json:"name" binding:"omitempty,min=1,max=100"`
	OpenAIBaseURL    *string  `json:"openai_base_url" binding:"omitempty,max=500"`
	AnthropicBaseURL *string  `json:"anthropic_base_url" binding:"omitempty,max=500"`
	Models           []string `json:"models"`
	Format           *string  `json:"format" binding:"omitempty,oneof=openai anthropic all"`
	Description      *string  `json:"description" binding:"omitempty,max=500"`
	Icon             *string  `json:"icon" binding:"omitempty,max=100"`
	SortOrder        *int     `json:"sort_order"`
	Status           *int16   `json:"status" binding:"omitempty,oneof=0 1"`
}

// CreateThirdPartyProviderRequest represents user create third-party service request.
type CreateThirdPartyProviderRequest struct {
	Name             string   `json:"name" binding:"required,min=1,max=100"`
	OpenAIBaseURL    string   `json:"openai_base_url" binding:"omitempty,max=500"`
	AnthropicBaseURL string   `json:"anthropic_base_url" binding:"omitempty,max=500"`
	APIKey           string   `json:"api_key" binding:"required"`
	Models           []string `json:"models" binding:"required,min=1"`
	Format           string   `json:"format" binding:"required,oneof=openai anthropic all"`
	TemplateID       *int64   `json:"template_id"`
}

// UpdateThirdPartyProviderRequest represents update third-party service request.
type UpdateThirdPartyProviderRequest struct {
	Name             *string  `json:"name" binding:"omitempty,min=1,max=100"`
	OpenAIBaseURL    *string  `json:"openai_base_url" binding:"omitempty,max=500"`
	AnthropicBaseURL *string  `json:"anthropic_base_url" binding:"omitempty,max=500"`
	APIKey           *string  `json:"api_key"`
	Models           []string `json:"models"`
	Format           *string  `json:"format" binding:"omitempty,oneof=openai anthropic all"`
	Status           *int16   `json:"status" binding:"omitempty,oneof=0 1"`
}

// CreateDocumentRequest represents create document request.
type CreateDocumentRequest struct {
	Slug        string `json:"slug" binding:"required,max=50"`
	Title       string `json:"title" binding:"required,max=200"`
	Subtitle    string `json:"subtitle" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=100"`
	Content     string `json:"content" binding:"required"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

// UpdateDocumentRequest represents update document request.
type UpdateDocumentRequest struct {
	Title       string `json:"title" binding:"required,max=200"`
	Subtitle    string `json:"subtitle" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=100"`
	Content     string `json:"content" binding:"required"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}
