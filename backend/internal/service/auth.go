package service

import (
	"context"
	"encoding/json"
	"time"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	jwtPkg "codemind/internal/pkg/jwt"
	"codemind/internal/pkg/validator"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// AuthService 认证业务逻辑
type AuthService struct {
	userRepo   *repository.UserRepository
	auditRepo  *repository.AuditRepository
	jwtManager *jwtPkg.Manager
	logger     *zap.Logger
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
	jwtManager *jwtPkg.Manager,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		auditRepo:  auditRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginRequest, clientIP string) (*dto.LoginResponse, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errcode.ErrInvalidCredentials
	}

	// 2. 验证密码
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errcode.ErrInvalidCredentials
	}

	// 3. 检查用户状态
	if !user.IsActive() {
		return nil, errcode.ErrAccountDisabled
	}

	// 4. 生成 JWT Token
	token, expiresAt, err := s.jwtManager.GenerateToken(
		user.ID, user.Username, user.Role, user.DepartmentID,
	)
	if err != nil {
		s.logger.Error("生成 Token 失败", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	// 5. 更新最后登录信息
	now := time.Now()
	_ = s.userRepo.UpdateFields(user.ID, map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": clientIP,
	})

	// 6. 构造响应
	resp := &dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: dto.UserBrief{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Role:        user.Role,
		},
	}

	if user.Department != nil {
		resp.User.Department = &dto.DeptBrief{
			ID:   user.Department.ID,
			Name: user.Department.Name,
		}
	}

	return resp, nil
}

// Logout 用户登出（将 Token 加入黑名单）
func (s *AuthService) Logout(claims *jwtPkg.Claims) error {
	return s.jwtManager.Blacklist(
		context.Background(),
		claims.ID,
		claims.ExpiresAt.Time,
	)
}

// GetProfile 获取当前用户信息
func (s *AuthService) GetProfile(userID int64) (*dto.UserDetail, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errcode.ErrUserNotFound
	}

	detail := &dto.UserDetail{
		ID:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		Phone:        user.Phone,
		AvatarURL:    user.AvatarURL,
		Role:         user.Role,
		DepartmentID: user.DepartmentID,
		Status:       user.Status,
		LastLoginAt:  user.LastLoginAt,
		CreatedAt:    user.CreatedAt,
	}

	if user.Department != nil {
		detail.Department = &dto.DeptBrief{
			ID:   user.Department.ID,
			Name: user.Department.Name,
		}
	}

	return detail, nil
}

// UpdateProfile 更新当前用户个人信息
func (s *AuthService) UpdateProfile(userID int64, req *dto.UpdateProfileRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	fields := make(map[string]interface{})
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Email != nil {
		// 检查邮箱唯一性
		if *req.Email != "" {
			exists, _ := s.userRepo.ExistsEmail(*req.Email, user.ID)
			if exists {
				return errcode.ErrEmailExists
			}
		}
		fields["email"] = req.Email
	}
	if req.Phone != nil {
		fields["phone"] = req.Phone
	}

	if len(fields) == 0 {
		return nil
	}

	return s.userRepo.UpdateFields(userID, fields)
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID int64, req *dto.ChangePasswordRequest, clientIP string) error {
	// 验证新密码强度
	if ok, msg := validator.ValidatePassword(req.NewPassword); !ok {
		return errcode.ErrInvalidParams.WithMessage(msg)
	}

	// 查找用户
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// 验证旧密码
	if !crypto.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errcode.ErrOldPasswordWrong
	}

	// 加密新密码
	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err))
		return errcode.ErrInternal
	}

	// 更新密码
	if err := s.userRepo.UpdateFields(userID, map[string]interface{}{
		"password_hash": hash,
	}); err != nil {
		return errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(userID, model.AuditActionResetPassword, model.AuditTargetUser, &userID, nil, clientIP)

	return nil
}

// recordAudit 记录审计日志（内部方法）
func (s *AuthService) recordAudit(operatorID int64, action, targetType string, targetID *int64, detail interface{}, clientIP string) {
	var detailJSON json.RawMessage
	if detail != nil {
		data, _ := json.Marshal(detail)
		detailJSON = data
	}

	log := &model.AuditLog{
		OperatorID: operatorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detailJSON,
		ClientIP:   &clientIP,
	}

	if err := s.auditRepo.Create(log); err != nil {
		s.logger.Error("记录审计日志失败", zap.Error(err), zap.String("action", action))
	}
}
