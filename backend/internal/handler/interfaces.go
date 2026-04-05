package handler

import (
	"context"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	jwtPkg "codemind/internal/pkg/jwt"
)

// DocumentService 文档服务接口
type DocumentService interface {
	List() ([]model.DocumentListItem, error)
	GetBySlug(slug string) (*model.Document, error)
	ListAll() ([]model.Document, error)
	GetByID(id int64) (*model.Document, error)
	Create(req *dto.CreateDocumentRequest) (*model.Document, error)
	Update(id int64, req *dto.UpdateDocumentRequest) (*model.Document, error)
	Delete(id int64) error
	Initialize() (int, error)
}

// AuthService 认证服务接口
type AuthService interface {
	Login(req *dto.LoginRequest, clientIP string) (*dto.LoginResponse, error)
	Logout(claims *jwtPkg.Claims) error
	GetProfile(userID int64) (*dto.UserDetail, error)
	UpdateProfile(userID int64, req *dto.UpdateProfileRequest) error
	ChangePassword(userID int64, req *dto.ChangePasswordRequest, claims *jwtPkg.Claims, clientIP string) error
	GetLoginLockStatusByUsername(username string) (*dto.LoginLockStatusResponse, error)
}

// APIKeyService API Key 服务接口
type APIKeyService interface {
	List(userID int64) ([]dto.APIKeyResponse, error)
	Create(req *dto.CreateAPIKeyRequest, userID int64, clientIP string) (*dto.APIKeyCreateResponse, error)
	UpdateStatus(keyID int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error
	Delete(keyID int64, operatorID int64, operatorRole string, clientIP string) error
}

// UserService 用户服务接口
type UserService interface {
	List(query *dto.UserListQuery, role string, deptID *int64) ([]dto.UserDetail, int64, error)
	Create(req *dto.CreateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) (*dto.UserDetail, error)
	GetDetail(id int64) (*dto.UserDetail, error)
	Update(id int64, req *dto.UpdateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error
	Delete(id int64, operatorID int64, clientIP string) error
	UpdateStatus(id int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error
	ResetPassword(id int64, newPassword string, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error
	UnlockUser(id int64, operatorID int64, operatorRole string, operatorDeptID *int64, reason string, clientIP string) error
}

// DepartmentService 部门服务接口
type DepartmentService interface {
	ListTree() ([]dto.DeptTree, error)
	Create(req *dto.CreateDepartmentRequest, operatorID int64, clientIP string) (*model.Department, error)
	GetByID(id int64) (*model.Department, error)
	Update(id int64, req *dto.UpdateDepartmentRequest, operatorID int64, clientIP string) error
	Delete(id int64, operatorID int64, clientIP string) error
}

// LimitService 限额服务接口
type LimitService interface {
	List(query *dto.LimitListQuery) ([]model.RateLimit, error)
	Upsert(req *dto.UpsertRateLimitRequest, operatorID int64, clientIP string) error
	Delete(id int64, operatorID int64, clientIP string) error
	GetMyLimits(userID int64, deptID *int64) (*dto.MyLimitResponse, error)
	GetLimitProgress(userID int64, deptID *int64) (*dto.LimitProgressResponse, error)
}

// ThirdPartyProviderService 第三方模型服务接口
type ThirdPartyProviderService interface {
	// 模板管理
	ListTemplates() ([]model.ThirdPartyProviderTemplate, error)
	ListActiveTemplates() ([]model.ThirdPartyProviderTemplate, error)
	CreateTemplate(name, openAIBaseURL, anthropicBaseURL, format string, models []string, description, icon *string, sortOrder int, operatorID int64) (*model.ThirdPartyProviderTemplate, error)
	UpdateTemplate(id int64, name, openAIBaseURL, anthropicBaseURL *string, models []string, format *string, description, icon *string, sortOrder *int, status *int16) error
	DeleteTemplate(id int64) error

	// 用户服务管理
	ListProviders(userID int64) ([]model.UserThirdPartyProvider, error)
	CreateProvider(userID int64, name, openAIBaseURL, anthropicBaseURL, apiKey, format string, models []string, templateID *int64) (*model.UserThirdPartyProvider, error)
	UpdateProvider(id, userID int64, name, openAIBaseURL, anthropicBaseURL, apiKey *string, models []string, format *string, status *int16) error
	UpdateProviderStatus(id, userID int64, status int16) error
	DeleteProvider(id, userID int64) error

	// 路由和用量
	ResolveThirdPartyModel(ctx context.Context, userID int64, modelName string, requestFormat string) *model.ThirdPartyRouteInfo
	DecryptAPIKey(encrypted string) (string, error)
	RecordThirdPartyUsage(userID, providerID, apiKeyID int64, modelName, requestType string, promptTokens, completionTokens, totalTokens int, durationMs *int)
	ListPlatformModels() ([]dto.PlatformModelInfo, error)
}

// SystemService 系统服务接口
type SystemService interface {
	GetConfigs() ([]model.SystemConfig, error)
	UpdateConfigs(req *dto.UpdateConfigsRequest, operatorID int64, clientIP string) error
	GetPlatformServiceURL() string
	ListAnnouncements(isAdmin bool) ([]model.Announcement, error)
	CreateAnnouncement(req *dto.CreateAnnouncementRequest, authorID int64, clientIP string) (*model.Announcement, error)
	UpdateAnnouncement(id int64, req *dto.UpdateAnnouncementRequest, operatorID int64, clientIP string) error
	DeleteAnnouncement(id int64, operatorID int64, clientIP string) error
	ListAuditLogs(query *dto.AuditLogQuery) ([]model.AuditLog, int64, error)
}
