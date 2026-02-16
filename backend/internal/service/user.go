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

// UserService 用户管理业务逻辑
type UserService struct {
	userRepo *repository.UserRepository
	deptRepo *repository.DepartmentRepository
	auditRepo *repository.AuditRepository
	logger   *zap.Logger
}

// NewUserService 创建用户服务
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

// Create 创建用户
func (s *UserService) Create(req *dto.CreateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) (*dto.UserDetail, error) {
	// 验证用户名格式
	if ok, msg := validator.ValidateUsername(req.Username); !ok {
		return nil, errcode.ErrInvalidParams.WithMessage(msg)
	}

	// 验证密码强度
	if ok, msg := validator.ValidatePassword(req.Password); !ok {
		return nil, errcode.ErrInvalidParams.WithMessage(msg)
	}

	// 检查用户名是否已存在
	exists, err := s.userRepo.ExistsUsername(req.Username)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if exists {
		return nil, errcode.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		emailExists, err := s.userRepo.ExistsEmail(req.Email)
		if err != nil {
			return nil, errcode.ErrDatabase
		}
		if emailExists {
			return nil, errcode.ErrEmailExists
		}
	}

	// 部门经理只能创建本部门用户
	if operatorRole == model.RoleDeptManager {
		if req.DepartmentID == nil || (operatorDeptID != nil && *req.DepartmentID != *operatorDeptID) {
			return nil, errcode.ErrForbiddenUser
		}
		// 部门经理不能创建管理员
		if req.Role != model.RoleUser {
			return nil, errcode.ErrForbiddenUser
		}
	}

	// 验证部门是否存在
	if req.DepartmentID != nil {
		_, err := s.deptRepo.FindByID(*req.DepartmentID)
		if err != nil {
			return nil, errcode.ErrDeptNotFound
		}
	}

	// 加密密码
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err))
		return nil, errcode.ErrInternal
	}

	// 构建用户对象
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

	// 创建用户
	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("创建用户失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionCreateUser, model.AuditTargetUser, &user.ID,
		map[string]interface{}{"username": req.Username, "role": req.Role}, clientIP)

	// 查询完整用户信息返回
	return s.GetDetail(user.ID)
}

// GetDetail 获取用户详情
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

// Update 更新用户信息
func (s *UserService) Update(id int64, req *dto.UpdateUserRequest, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// 权限检查：部门经理只能编辑本部门用户
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

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionUpdateUser, model.AuditTargetUser, &id, fields, clientIP)

	return nil
}

// Delete 软删除用户（仅超级管理员）
func (s *UserService) Delete(id int64, operatorID int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// 不允许删除自己
	if id == operatorID {
		return errcode.ErrInvalidParams.WithMessage("不能删除自己的账号")
	}

	if err := s.userRepo.Delete(id); err != nil {
		return errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionDeleteUser, model.AuditTargetUser, &id,
		map[string]string{"username": user.Username}, clientIP)

	return nil
}

// UpdateStatus 切换用户状态
func (s *UserService) UpdateStatus(id int64, status int16, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// 部门经理只能操作本部门用户
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

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword(id int64, newPassword string, operatorID int64, operatorRole string, operatorDeptID *int64, clientIP string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errcode.ErrUserNotFound
	}

	// 部门经理只能重置本部门用户密码
	if operatorRole == model.RoleDeptManager {
		if user.DepartmentID == nil || operatorDeptID == nil || *user.DepartmentID != *operatorDeptID {
			return errcode.ErrForbiddenUser
		}
	}

	// 验证新密码强度
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

// List 分页查询用户列表
func (s *UserService) List(query *dto.UserListQuery, operatorRole string, operatorDeptID *int64) ([]dto.UserDetail, int64, error) {
	filters := map[string]interface{}{
		"keyword":       query.Keyword,
		"department_id": query.DepartmentID,
		"role":          query.Role,
		"status":        query.Status,
	}

	// 部门经理只能查看本部门用户
	if operatorRole == model.RoleDeptManager && operatorDeptID != nil {
		filters["department_id"] = operatorDeptID
	}

	users, total, err := s.userRepo.List(query.GetPage(), query.GetPageSize(), filters)
	if err != nil {
		return nil, 0, errcode.ErrDatabase
	}

	// 转换为 DTO
	var details []dto.UserDetail
	for _, u := range users {
		d := dto.UserDetail{
			ID:           u.ID,
			Username:     u.Username,
			DisplayName:  u.DisplayName,
			Email:        u.Email,
			Phone:        u.Phone,
			AvatarURL:    u.AvatarURL,
			Role:         u.Role,
			DepartmentID: u.DepartmentID,
			Status:       u.Status,
			LastLoginAt:  u.LastLoginAt,
			CreatedAt:    u.CreatedAt,
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

// recordAudit 记录审计日志
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
		s.logger.Error("记录审计日志失败",
			zap.Error(err),
			zap.String("action", action),
			zap.Int64("operator_id", operatorID),
		)
	}
}

// ImportUsers 批量导入用户（从 CSV）
func (s *UserService) ImportUsers(users []dto.CreateUserRequest, operatorID int64, clientIP string) (int, []string, error) {
	var successCount int
	var errors []string

	for i, req := range users {
		_, err := s.Create(&req, operatorID, model.RoleSuperAdmin, nil, clientIP)
		if err != nil {
			errors = append(errors, fmt.Sprintf("第 %d 行: %s", i+2, err.Error()))
			continue
		}
		successCount++
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionImportUsers, model.AuditTargetUser, nil,
		map[string]interface{}{"total": len(users), "success": successCount, "failed": len(errors)}, clientIP)

	return successCount, errors, nil
}
