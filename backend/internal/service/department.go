package service

import (
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"
	"encoding/json"

	"go.uber.org/zap"
)

// DepartmentService handles department management.
type DepartmentService struct {
	deptRepo  *repository.DepartmentRepository
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	logger    *zap.Logger
}

// NewDepartmentService creates a department service.
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

// Create creates a new department.
func (s *DepartmentService) Create(req *dto.CreateDepartmentRequest, operatorID int64, clientIP string) (*model.Department, error) {
	exists, err := s.deptRepo.ExistsName(req.Name)
	if err != nil {
		return nil, errcode.ErrDatabase
	}
	if exists {
		return nil, errcode.ErrInvalidParams.WithMessage("department name already exists")
	}

	if req.ParentID != nil {
		if _, err := s.deptRepo.FindByID(*req.ParentID); err != nil {
			return nil, errcode.ErrDeptNotFound.WithMessage("parent department not found")
		}
	}

	if req.ManagerID != nil {
		if _, err := s.userRepo.FindByID(*req.ManagerID); err != nil {
			return nil, errcode.ErrUserNotFound.WithMessage("specified manager not found")
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
		s.logger.Error("failed to create department", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionCreateDept, model.AuditTargetDepartment, &dept.ID,
		map[string]interface{}{"name": req.Name}, clientIP)

	return dept, nil
}

// GetByID retrieves department by ID.
func (s *DepartmentService) GetByID(id int64) (*model.Department, error) {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return nil, errcode.ErrDeptNotFound
	}
	return dept, nil
}

// Update updates department information.
func (s *DepartmentService) Update(id int64, req *dto.UpdateDepartmentRequest, operatorID int64, clientIP string) error {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return errcode.ErrDeptNotFound
	}

	fields := make(map[string]interface{})

	if req.Name != nil {
		exists, _ := s.deptRepo.ExistsName(*req.Name, id)
		if exists {
			return errcode.ErrInvalidParams.WithMessage("department name already exists")
		}
		fields["name"] = *req.Name
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.ParentID != nil {
		if *req.ParentID == id {
			return errcode.ErrInvalidParams.WithMessage("cannot set self as parent department")
		}
		fields["parent_id"] = *req.ParentID
	}
	if req.ManagerID != nil {
		if _, err := s.userRepo.FindByID(*req.ManagerID); err != nil {
			return errcode.ErrUserNotFound.WithMessage("specified manager not found")
		}
		fields["manager_id"] = *req.ManagerID
	}

	if len(fields) == 0 {
		return nil
	}

	if err := s.deptRepo.UpdateFields(id, fields); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionUpdateDept, model.AuditTargetDepartment, &id,
		map[string]interface{}{"old_name": dept.Name, "changes": fields}, clientIP)

	return nil
}

// Delete deletes a department.
func (s *DepartmentService) Delete(id int64, operatorID int64, clientIP string) error {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return errcode.ErrDeptNotFound
	}

	count, err := s.userRepo.CountByDepartment(id)
	if err != nil {
		return errcode.ErrDatabase
	}
	if count > 0 {
		return errcode.ErrDeptHasUsers
	}

	hasChildren, err := s.deptRepo.HasChildren(id)
	if err != nil {
		return errcode.ErrDatabase
	}
	if hasChildren {
		return errcode.ErrInvalidParams.WithMessage("cannot delete department with sub-departments")
	}

	if err := s.deptRepo.Delete(id); err != nil {
		return errcode.ErrDatabase
	}

	s.recordAudit(operatorID, model.AuditActionDeleteDept, model.AuditTargetDepartment, &id,
		map[string]string{"name": dept.Name}, clientIP)

	return nil
}

// ListTree returns departments as a tree structure.
func (s *DepartmentService) ListTree() ([]dto.DeptTree, error) {
	depts, err := s.deptRepo.ListAll()
	if err != nil {
		return nil, errcode.ErrDatabase
	}

	userCounts, err := s.userRepo.CountByDepartmentBatch()
	if err != nil {
		s.logger.Warn("failed to batch count department users", zap.Error(err))
		userCounts = make(map[int64]int)
	}

	return s.buildTree(depts, nil, userCounts), nil
}

func (s *DepartmentService) buildTree(depts []model.Department, parentID *int64, userCounts map[int64]int) []dto.DeptTree {
	var tree []dto.DeptTree

	for _, dept := range depts {
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
		s.logger.Error("failed to record audit log", zap.Error(err), zap.String("action", action))
	}
}
