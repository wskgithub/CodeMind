package service

import (
	"encoding/json"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// DepartmentService 部门管理业务逻辑
type DepartmentService struct {
	deptRepo  *repository.DepartmentRepository
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	logger    *zap.Logger
}

// NewDepartmentService 创建部门服务
func NewDepartmentService(
	deptRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
	logger *zap.Logger,
) *DepartmentService {
	return &DepartmentService{
		deptRepo:  deptRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// Create 创建部门
func (s *DepartmentService) Create(req *dto.CreateDepartmentRequest, operatorID int64, clientIP string) (*model.Department, error) {
	// 检查部门名称是否已存在
	exists, err := s.deptRepo.ExistsName(req.Name)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if exists {
		return nil, errcode.ErrInvalidParams.WithMessage("部门名称已存在")
	}

	// 验证上级部门
	if req.ParentID != nil {
		if _, err := s.deptRepo.FindByID(*req.ParentID); err != nil {
			return nil, errcode.ErrDeptNotFound.WithMessage("上级部门不存在")
		}
	}

	// 验证经理用户
	if req.ManagerID != nil {
		if _, err := s.userRepo.FindByID(*req.ManagerID); err != nil {
			return nil, errcode.ErrUserNotFound.WithMessage("指定的部门经理不存在")
		}
	}

	dept := &model.Department{
		Name:      req.Name,
		ParentID:  req.ParentID,
		ManagerID: req.ManagerID,
		Status:    model.StatusEnabled,
	}
	if req.Description != nil {
		dept.Description = req.Description
	}

	if err := s.deptRepo.Create(dept); err != nil {
		s.logger.Error("创建部门失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionCreateDept, model.AuditTargetDepartment, &dept.ID,
		map[string]interface{}{"name": req.Name}, clientIP)

	return dept, nil
}

// GetByID 获取部门详情
func (s *DepartmentService) GetByID(id int64) (*model.Department, error) {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return nil, errcode.ErrDeptNotFound
	}
	return dept, nil
}

// Update 更新部门信息
func (s *DepartmentService) Update(id int64, req *dto.UpdateDepartmentRequest, operatorID int64, clientIP string) error {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return errcode.ErrDeptNotFound
	}

	fields := make(map[string]interface{})

	if req.Name != nil {
		// 检查名称唯一性
		exists, _ := s.deptRepo.ExistsName(*req.Name, id)
		if exists {
			return errcode.ErrInvalidParams.WithMessage("部门名称已存在")
		}
		fields["name"] = *req.Name
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.ParentID != nil {
		// 不能将自己设为上级
		if *req.ParentID == id {
			return errcode.ErrInvalidParams.WithMessage("不能将自己设为上级部门")
		}
		fields["parent_id"] = *req.ParentID
	}
	if req.ManagerID != nil {
		if _, err := s.userRepo.FindByID(*req.ManagerID); err != nil {
			return errcode.ErrUserNotFound.WithMessage("指定的部门经理不存在")
		}
		fields["manager_id"] = *req.ManagerID
	}

	if len(fields) == 0 {
		return nil
	}

	if err := s.deptRepo.UpdateFields(id, fields); err != nil {
		return errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionUpdateDept, model.AuditTargetDepartment, &id,
		map[string]interface{}{"old_name": dept.Name, "changes": fields}, clientIP)

	return nil
}

// Delete 删除部门
func (s *DepartmentService) Delete(id int64, operatorID int64, clientIP string) error {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return errcode.ErrDeptNotFound
	}

	// 检查部门下是否有用户
	count, err := s.userRepo.CountByDepartment(id)
	if err != nil {
		return errcode.ErrDatabase
	}
	if count > 0 {
		return errcode.ErrDeptHasUsers
	}

	// 检查是否有子部门
	hasChildren, err := s.deptRepo.HasChildren(id)
	if err != nil {
		return errcode.ErrDatabase
	}
	if hasChildren {
		return errcode.ErrInvalidParams.WithMessage("部门下还有子部门，无法删除")
	}

	if err := s.deptRepo.Delete(id); err != nil {
		return errcode.ErrDatabase
	}

	// 记录审计日志
	s.recordAudit(operatorID, model.AuditActionDeleteDept, model.AuditTargetDepartment, &id,
		map[string]string{"name": dept.Name}, clientIP)

	return nil
}

// ListTree 获取部门树形结构
func (s *DepartmentService) ListTree() ([]dto.DeptTree, error) {
	depts, err := s.deptRepo.ListAll()
	if err != nil {
		return nil, errcode.ErrDatabase
	}

	// 查询每个部门的用户数
	userCounts := make(map[int64]int)
	for _, dept := range depts {
		count, _ := s.userRepo.CountByDepartment(dept.ID)
		userCounts[dept.ID] = int(count)
	}

	// 构建树形结构
	return s.buildTree(depts, nil, userCounts), nil
}

// buildTree 递归构建部门树
func (s *DepartmentService) buildTree(depts []model.Department, parentID *int64, userCounts map[int64]int) []dto.DeptTree {
	var tree []dto.DeptTree

	for _, dept := range depts {
		// 匹配当前层级的部门
		if (parentID == nil && dept.ParentID == nil) ||
			(parentID != nil && dept.ParentID != nil && *dept.ParentID == *parentID) {

			node := dto.DeptTree{
				ID:          dept.ID,
				Name:        dept.Name,
				Description: dept.Description,
				UserCount:   userCounts[dept.ID],
				Status:      dept.Status,
				Children:    s.buildTree(depts, &dept.ID, userCounts),
			}

			if dept.Manager != nil {
				node.Manager = &dto.UserBrief{
					ID:          dept.Manager.ID,
					Username:    dept.Manager.Username,
					DisplayName: dept.Manager.DisplayName,
					Role:        dept.Manager.Role,
				}
			}

			tree = append(tree, node)
		}
	}

	if tree == nil {
		tree = []dto.DeptTree{}
	}
	return tree
}

// recordAudit 记录审计日志
func (s *DepartmentService) recordAudit(operatorID int64, action, targetType string, targetID *int64, detail interface{}, clientIP string) {
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
