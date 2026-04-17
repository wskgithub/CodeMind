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

// LoginLockConfig defines account lockout parameters
var LoginLockConfig = struct {
	MaxFailCount    int
	InitialLockTime time.Duration
	MaxLockTime     time.Duration
}{
	MaxFailCount:    5,
	InitialLockTime: 5 * time.Minute,
	MaxLockTime:     24 * time.Hour,
}

// AuthService handles authentication operations
type AuthService struct {
	userRepo   *repository.UserRepository
	auditRepo  *repository.AuditRepository
	jwtManager *jwtPkg.Manager
	logger     *zap.Logger
}

// NewAuthService creates a new AuthService
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

// LoginResult contains login outcome with possible error details
type LoginResult struct {
	Success       bool
	Response      *dto.LoginResponse
	Err           error
	Locked        bool
	RemainingTime int64
	FailCount     int
}

// Used for timing-safe comparison when user doesn't exist
var dummyPasswordHash, _ = crypto.HashPassword("codemind-timing-safe-dummy")

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(req *dto.LoginRequest, clientIP string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		// Perform bcrypt comparison even for non-existent users to prevent timing attacks
		crypto.CheckPassword(req.Password, dummyPasswordHash)
		return nil, errcode.ErrInvalidCredentials
	}

	if user.IsLocked() {
		return nil, errcode.ErrAccountLocked.WithMessage(
			errors.New("account is locked, please try again later").Error(),
		)
	}

	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		updatedUser, lockErr := s.handleLoginFailure(user.ID)
		if lockErr != nil {
			s.logger.Error("failed to handle login failure", zap.Error(lockErr))
		}

		if updatedUser != nil && updatedUser.IsLocked() {
			return nil, errcode.ErrAccountLocked.WithMessage(
				errors.New("account is locked, please try again later").Error(),
			)
		}
		
		return nil, errcode.ErrInvalidCredentials
	}

	if !user.IsActive() {
		return nil, errcode.ErrAccountDisabled
	}

	if user.LoginFailCount > 0 {
		if err := s.userRepo.ClearLoginFailCount(user.ID); err != nil {
			s.logger.Error("failed to clear login fail count", zap.Error(err))
		}
	}

	token, expiresAt, err := s.jwtManager.GenerateToken(
		user.ID, user.Username, user.Role, user.DepartmentID,
	)
	if err != nil {
		s.logger.Error("failed to generate token", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	now := time.Now()
	_ = s.userRepo.UpdateFields(user.ID, map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": clientIP,
	})
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

// handleLoginFailure increments fail count and may lock the account
func (s *AuthService) handleLoginFailure(userID int64) (*model.User, error) {
	user, err := s.userRepo.IncrementLoginFailCount(userID)
	if err != nil {
		return nil, err
	}

	if user.LoginFailCount >= LoginLockConfig.MaxFailCount {
		lockDuration := s.calculateLockDuration(user.LoginFailCount)
		lockedUntil := time.Now().Add(lockDuration)
		
		if err := s.userRepo.LockAccount(userID, lockedUntil); err != nil {
			s.logger.Error("failed to lock account", zap.Error(err), zap.Int64("user_id", userID))
			return user, err
		}
		
		return s.userRepo.FindByID(userID)
	}

	return user, nil
}

// calculateLockDuration calculates lock duration based on fail count (doubles each time)
func (s *AuthService) calculateLockDuration(failCount int) time.Duration {
	baseDuration := LoginLockConfig.InitialLockTime
	
	excessCount := failCount - LoginLockConfig.MaxFailCount
	if excessCount < 0 {
		excessCount = 0
	}
	
	duration := baseDuration
	for i := 0; i < excessCount; i++ {
		duration *= 2
		if duration > LoginLockConfig.MaxLockTime {
			duration = LoginLockConfig.MaxLockTime
			break
		}
	}
	
	return duration
}

// GetLoginLockStatus retrieves user's login lock status
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

// GetLoginLockStatusByUsername retrieves login lock status by username
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

// Logout invalidates the user's token
func (s *AuthService) Logout(claims *jwtPkg.Claims) error {
	return s.jwtManager.Blacklist(
		context.Background(),
		claims.ID,
		claims.ExpiresAt.Time,
	)
}

// GetProfile retrieves current user's profile
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

// UpdateProfile updates current user's profile
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

// ChangePassword changes user's password and invalidates current token
func (s *AuthService) ChangePassword(userID int64, req *dto.ChangePasswordRequest, claims *jwtPkg.Claims, clientIP string) error {
	if ok, msg := validator.ValidatePassword(req.NewPassword); !ok {
		return errcode.ErrInvalidParams.WithMessage(msg)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	if !crypto.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errcode.ErrOldPasswordWrong
	}

	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return errcode.ErrInternal
	}

	if err := s.userRepo.UpdateFields(userID, map[string]interface{}{
		"password_hash": hash,
	}); err != nil {
		return errcode.ErrDatabase
	}

	// Invalidate current token to force re-login
	if claims != nil {
		if err := s.jwtManager.Blacklist(context.Background(), claims.ID, claims.ExpiresAt.Time); err != nil {
			s.logger.Error("failed to blacklist token after password change", zap.Error(err))
		}
	}

	s.recordAudit(userID, model.AuditActionResetPassword, model.AuditTargetUser, &userID, nil, clientIP)

	return nil
}

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
		s.logger.Error("failed to record audit log", zap.Error(err), zap.String("action", action))
	}
}
