package service

import (
	"encoding/json"
	"fmt"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/crypto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/validator"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// UserService handles user management operations
type UserService struct {
	userRepo *repository.UserRepository
	deptRepo *repository.DepartmentRepository
	auditRepo *repository.AuditRepository
	logger   *zap.Logger
}

// NewUserService creates a new UserService
func NewUserService(
	userRepo *repository.UserRepository,
	deptRepo *repository.DepartmentRepository,
	auditRepo *repository.AuditRepository,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		deptRepo:  deptRepo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// Create creates a new user
func (s *UserService) Create(req *dto.CreateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) (*dto.UserDetail, error) {
	if ok, msg := validator.ValidateUsername(req.Username); !ok {
		return nil, errcode.ErrInvalidParams.WithMessage(msg)
	}

	if ok, msg := validator.ValidatePassword(req.Password); !ok {
		return nil, errcode.ErrInvalidParams.WithMessage(msg)
	}

	exists, err := s.userRepo.ExistsUsername(req.Username)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if exists {
		return nil, errcode.ErrUsernameExists
	}

	// Hard delete soft-deleted user with same username to release the username
	existsIncludingDeleted, err := s.userRepo.ExistsUsernameIncludingDeleted(req.Username)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if existsIncludingDeleted {
		if err := s.userRepo.HardDeleteSoftDeletedUser(req.Username); err != nil {
			s.logger.Error("failed to hard delete soft-deleted user", zap.Error(err), zap.String("username", req.Username))
			return nil, errcode.ErrDatabase
		}
	}

	if req.Email != "" {
		emailExists, err := s.userRepo.ExistsEmail(req.Email)
		if err != nil {
			return nil, errcode.ErrDatabase
		}
		if emailExists {
			return nil, errcode.ErrEmailExists
		}

		emailExistsIncludingDeleted, err := s.userRepo.ExistsEmailIncludingDeleted(req.Email)
		if err != nil {
			return nil, errcode.ErrDatabase
		}
		if emailExistsIncludingDeleted {
			if err := s.userRepo.HardDeleteSoftDeletedUserByEmail(req.Email); err != nil {
				s.logger.Error("failed to hard delete soft-deleted user", zap.Error(err), zap.String("email", req.Email))
				return nil, errcode.ErrDatabase
			}
		}
	}

	// Dept managers can only create users in their own department
	if operatorRole == model.RoleDeptManager {
		if req.DepartmentID == nil || (operatorDeptID != nil && *req.DepartmentID != *operatorDeptID) {
			return nil, errcode.ErrForbiddenUser
		}
		if req.Role != model.RoleUser {
			return nil, errcode.ErrForbiddenUser
		}
	}

	if req.DepartmentID != nil {
		_, err := s.deptRepo.FindByID(*req.DepartmentID)
		if err != nil {
			return nil, errcode.ErrDeptNotFound
		}
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
		Role:         req.Role,
		DepartmentID: req.DepartmentID,
		Status:       model.StatusEnabled,
	}
	if req.Email != "" {
		user.Email = &req.Email
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionCreateUser, model.AuditTargetUser, &user.ID,
		map[string]interface{}{"username": req.Username, "role": req.Role}, clientIP)

	return s.GetDetail(user.ID)
}

// GetDetail retrieves user details by ID
func (s *UserService) GetDetail(id int64) (*dto.UserDetail, error) {
	user, err := s.userRepo.FindByID(id)
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

// Update updates user information
func (s *UserService) Update(id int64, req *dto.UpdateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// Dept managers can only edit users in their own department
	if operatorRole == model.RoleDeptManager {
		if user.DepartmentID == nil || operatorDeptID == nil || *user.DepartmentID != *operatorDeptID {
			return errcode.ErrForbiddenUser
		}
	}

	fields := make(map[string]interface{})
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Email != nil {
		if *req.Email != "" {
			exists, _ := s.userRepo.ExistsEmail(*req.Email, id)
			if exists {
				return errcode.ErrEmailExists
			}
		}
		fields["email"] = req.Email
	}
	if req.Phone != nil {
		fields["phone"] = req.Phone
	}
	if req.Role != nil && operatorRole == model.RoleSuperAdmin {
		fields["role"] = *req.Role
	}
	if req.DepartmentID != nil {
		if _, err := s.deptRepo.FindByID(*req.DepartmentID); err != nil {
			return errcode.ErrDeptNotFound
		}
		fields["department_id"] = *req.DepartmentID
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if len(fields) == 0 {
		return nil
	}

	if err := s.userRepo.UpdateFields(id, fields); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionUpdateUser, model.AuditTargetUser, &id, fields, clientIP)

	return nil
}

// Delete soft-deletes a user (super admin only)
func (s *UserService) Delete(id int64, operatorID int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	if id == operatorID {
		return errcode.ErrInvalidParams.WithMessage("cannot delete your own account")
	}

	if err := s.userRepo.Delete(id); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionDeleteUser, model.AuditTargetUser, &id,
		map[string]string{"username": user.Username}, clientIP)

	return nil
}

// UpdateStatus toggles user status
func (s *UserService) UpdateStatus(id int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// Dept managers can only operate on users in their own department
	if operatorRole == model.RoleDeptManager {
		if user.DepartmentID == nil || operatorDeptID == nil || *user.DepartmentID != *operatorDeptID {
			return errcode.ErrForbiddenUser
		}
	}

	if err := s.userRepo.UpdateFields(id, map[string]interface{}{"status": status}); err != nil {
		return errcode.ErrDatabase
	}

	action := model.AuditActionEnableUser
	if status == model.StatusDisabled {
		action = model.AuditActionDisableUser
	}
	s.recordAudit(operatorID, action, model.AuditTargetUser, &id, nil, clientIP)

	return nil
}

// ResetPassword resets a user's password
func (s *UserService) ResetPassword(id int64, newPassword string, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// Dept managers can only reset passwords for users in their own department
	if operatorRole == model.RoleDeptManager {
		if user.DepartmentID == nil || operatorDeptID == nil || *user.DepartmentID != *operatorDeptID {
			return errcode.ErrForbiddenUser
		}
	}

	if ok, msg := validator.ValidatePassword(newPassword); !ok {
		return errcode.ErrInvalidParams.WithMessage(msg)
	}

	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errcode.ErrInternal
	}

	if err := s.userRepo.UpdateFields(id, map[string]interface{}{"password_hash": hash}); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionResetPassword, model.AuditTargetUser, &id,
		map[string]string{"username": user.Username}, clientIP)

	return nil
}

// UnlockUser unlocks a locked user account
func (s *UserService) UnlockUser(id int64, operatorID int64, operatorRole string, operatorDeptID *int64, reason string, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	if !user.IsLocked() && user.LoginFailCount == 0 {
		return errcode.ErrInvalidParams.WithMessage("user is not locked")
	}

	// Dept managers can only unlock users in their own department
	if operatorRole == model.RoleDeptManager {
		if user.DepartmentID == nil || operatorDeptID == nil || *user.DepartmentID != *operatorDeptID {
			return errcode.ErrForbiddenUser
		}
	}

	if err := s.userRepo.ClearLoginFailCount(id); err != nil {
		s.logger.Error("failed to unlock user", zap.Error(err), zap.Int64("user_id", id))
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionUnlockUser, model.AuditTargetUser, &id,
		map[string]interface{}{
			"username": user.Username,
			"reason":   reason,
		}, clientIP)

	return nil
}

// List returns a paginated list of users
func (s *UserService) List(query *dto.UserListQuery, operatorRole string, operatorDeptID *int64) ([]dto.UserDetail, int64, error) {
	filters := map[string]interface{}{
		"keyword":       query.Keyword,
		"department_id": query.DepartmentID,
		"role":          query.Role,
		"status":        query.Status,
	}

	// Dept managers can only view users in their own department
	if operatorRole == model.RoleDeptManager && operatorDeptID != nil {
		filters["department_id"] = operatorDeptID
	}

	users, total, err := s.userRepo.List(query.GetPage(), query.GetPageSize(), filters)
	if err != nil {
		return nil, 0, errcode.ErrDatabase
	}

	var details []dto.UserDetail
	for _, u := range users {
		d := dto.UserDetail{
			ID:              u.ID,
			Username:        u.Username,
			DisplayName:     u.DisplayName,
			Email:           u.Email,
			Phone:           u.Phone,
			AvatarURL:       u.AvatarURL,
			Role:            u.Role,
			DepartmentID:    u.DepartmentID,
			Status:          u.Status,
			LastLoginAt:     u.LastLoginAt,
			LoginFailCount:  u.LoginFailCount,
			LockedUntil:     u.LockedUntil,
			CreatedAt:       u.CreatedAt,
		}
		if u.Department != nil {
			d.Department = &dto.DeptBrief{
				ID:   u.Department.ID,
				Name: u.Department.Name,
			}
		}
		details = append(details, d)
	}

	return details, total, nil
}

func (s *UserService) recordAudit(operatorID int64, action, targetType string, targetID *int64, detail interface{}, clientIP string) {
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
		s.logger.Error("failed to record audit log",
			zap.Error(err),
			zap.String("action", action),
			zap.Int64("operator_id", operatorID),
		)
	}
}

// ImportUsers imports users from CSV
func (s *UserService) ImportUsers(users []dto.CreateUserRequest, operatorID int64, clientIP string) (int, []string, error) {
	var successCount int
	var errors []string

	for i, req := range users {
		_, err := s.Create(&req, operatorID, model.RoleSuperAdmin, nil, clientIP)
		if err != nil {
			errors = append(errors, fmt.Sprintf("row %d: %s", i+2, err.Error()))
			continue
		}
		successCount++
	}

	s.recordAudit(operatorID, model.AuditActionImportUsers, model.AuditTargetUser, nil,
		map[string]interface{}{"total": len(users), "success": successCount, "failed": len(errors)}, clientIP)

	return successCount, errors, nil
}
