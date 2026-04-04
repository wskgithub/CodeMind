package service

import (
	"context"
	"encoding/json"
	"errors"
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

// LoginLockConfig 登录锁定配置
var LoginLockConfig = struct {
	MaxFailCount    int           // 最大失败次数
	InitialLockTime time.Duration // 初始锁定时间
	MaxLockTime     time.Duration // 最大锁定时间
}{
	MaxFailCount:    5,                    // 5次失败后锁定
	InitialLockTime: 5 * time.Minute,      // 首次锁定5分钟
	MaxLockTime:     24 * time.Hour,       // 最大锁定24小时
}

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

// LoginResult 登录结果，包含可能的错误信息
type LoginResult struct {
	Success       bool
	Response      *dto.LoginResponse
	Err           error
	Locked        bool
	RemainingTime int64
	FailCount     int
}

// dummyPasswordHash 用于用户不存在时执行等耗时的 bcrypt 比较，防止时序侧信道枚举用户名
var dummyPasswordHash, _ = crypto.HashPassword("codemind-timing-safe-dummy")

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginRequest, clientIP string) (*dto.LoginResponse, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		// 用户不存在时也执行一次 bcrypt 比较，消除与「密码错误」之间的时序差异
		crypto.CheckPassword(req.Password, dummyPasswordHash)
		return nil, errcode.ErrInvalidCredentials
	}

	// 2. 检查账号是否被锁定
	if user.IsLocked() {
		return nil, errcode.ErrAccountLocked.WithMessage(
			errors.New("账号已被锁定，请稍后再试").Error(),
		)
	}

	// 3. 验证密码
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		// 密码错误，记录失败并可能锁定账号
		updatedUser, lockErr := s.handleLoginFailure(user.ID)
		if lockErr != nil {
			s.logger.Error("处理登录失败时出错", zap.Error(lockErr))
		}

		// 根据更新后的状态返回不同的错误
		if updatedUser != nil && updatedUser.IsLocked() {
			return nil, errcode.ErrAccountLocked.WithMessage(
				errors.New("账号已被锁定，请稍后再试").Error(),
			)
		}
		
		return nil, errcode.ErrInvalidCredentials
	}

	// 4. 检查用户状态
	if !user.IsActive() {
		return nil, errcode.ErrAccountDisabled
	}

	// 5. 登录成功，清除失败记录
	if user.LoginFailCount > 0 {
		if err := s.userRepo.ClearLoginFailCount(user.ID); err != nil {
			s.logger.Error("清除登录失败记录失败", zap.Error(err))
		}
	}

	// 6. 生成 JWT Token
	token, expiresAt, err := s.jwtManager.GenerateToken(
		user.ID, user.Username, user.Role, user.DepartmentID,
	)
	if err != nil {
		s.logger.Error("生成 Token 失败", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	// 7. 更新最后登录信息
	now := time.Now()
	_ = s.userRepo.UpdateFields(user.ID, map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": clientIP,
	})

	// 8. 构造响应
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

// handleLoginFailure 处理登录失败，增加失败次数并可能锁定账号
func (s *AuthService) handleLoginFailure(userID int64) (*model.User, error) {
	// 增加失败次数
	user, err := s.userRepo.IncrementLoginFailCount(userID)
	if err != nil {
		return nil, err
	}

	// 检查是否需要锁定账号
	if user.LoginFailCount >= LoginLockConfig.MaxFailCount {
		// 计算锁定时间：每次失败后锁定时间翻倍
		lockDuration := s.calculateLockDuration(user.LoginFailCount)
		lockedUntil := time.Now().Add(lockDuration)
		
		if err := s.userRepo.LockAccount(userID, lockedUntil); err != nil {
			s.logger.Error("锁定账号失败", zap.Error(err), zap.Int64("user_id", userID))
			return user, err
		}
		
		// 重新查询获取更新后的状态
		return s.userRepo.FindByID(userID)
	}

	return user, nil
}

// calculateLockDuration 根据失败次数计算锁定时间
func (s *AuthService) calculateLockDuration(failCount int) time.Duration {
	// 基础锁定时间
	baseDuration := LoginLockConfig.InitialLockTime
	
	// 超过最大失败次数后，每次翻倍
	excessCount := failCount - LoginLockConfig.MaxFailCount
	if excessCount < 0 {
		excessCount = 0
	}
	
	// 计算锁定时间：5分钟 * 2^excessCount
	duration := baseDuration
	for i := 0; i < excessCount; i++ {
		duration *= 2
		// 不超过最大锁定时间
		if duration > LoginLockConfig.MaxLockTime {
			duration = LoginLockConfig.MaxLockTime
			break
		}
	}
	
	return duration
}

// GetLoginLockStatus 获取用户的登录锁定状态
func (s *AuthService) GetLoginLockStatus(userID int64) (*dto.LoginLockStatusResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errcode.ErrUserNotFound
	}

	return &dto.LoginLockStatusResponse{
		LoginFailCount: user.LoginFailCount,
		Locked:         user.IsLocked(),
		LockedUntil:    user.LockedUntil,
		RemainingTime:  user.GetRemainingLockTime(),
	}, nil
}

// GetLoginLockStatusByUsername 根据用户名获取登录锁定状态
func (s *AuthService) GetLoginLockStatusByUsername(username string) (*dto.LoginLockStatusResponse, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errcode.ErrUserNotFound
	}

	return &dto.LoginLockStatusResponse{
		LoginFailCount: user.LoginFailCount,
		Locked:         user.IsLocked(),
		LockedUntil:    user.LockedUntil,
		RemainingTime:  user.GetRemainingLockTime(),
	}, nil
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
		ID:              user.ID,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           user.Email,
		Phone:           user.Phone,
		AvatarURL:       user.AvatarURL,
		Role:            user.Role,
		DepartmentID:    user.DepartmentID,
		Status:          user.Status,
		LastLoginAt:     user.LastLoginAt,
		LoginFailCount:  user.LoginFailCount,
		LockedUntil:     user.LockedUntil,
		CreatedAt:       user.CreatedAt,
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
// claims 为当前会话的 JWT 声明，改密成功后将该 Token 加入黑名单以强制重新登录
func (s *AuthService) ChangePassword(userID int64, req *dto.ChangePasswordRequest, claims *jwtPkg.Claims, clientIP string) error {
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

	// 将当前 Token 加入黑名单，强制用户使用新密码重新登录
	if claims != nil {
		if err := s.jwtManager.Blacklist(context.Background(), claims.ID, claims.ExpiresAt.Time); err != nil {
			s.logger.Error("改密后将 Token 加入黑名单失败", zap.Error(err))
		}
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
