package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// MCPRepository handles MCP service data access.
type MCPRepository struct {
	db *gorm.DB
}

// NewMCPRepository creates a new MCP repository.
func NewMCPRepository(db *gorm.DB) *MCPRepository {
	return &MCPRepository{db: db}
}

// CreateService creates an MCP service.
func (r *MCPRepository) CreateService(svc *model.MCPService) error {
	return r.db.Create(svc).Error
}

// GetServiceByID returns service by ID.
func (r *MCPRepository) GetServiceByID(id int64) (*model.MCPService, error) {
	var svc model.MCPService
	err := r.db.First(&svc, id).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// GetServiceByName returns service by name.
func (r *MCPRepository) GetServiceByName(name string) (*model.MCPService, error) {
	var svc model.MCPService
	err := r.db.Where("name = ?", name).First(&svc).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// ListServices returns service list.
func (r *MCPRepository) ListServices(status string) ([]model.MCPService, error) {
	var services []model.MCPService
	query := r.db.Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&services).Error
	return services, err
}

// UpdateService updates a service.
func (r *MCPRepository) UpdateService(svc *model.MCPService) error {
	return r.db.Save(svc).Error
}

// DeleteService soft deletes a service.
func (r *MCPRepository) DeleteService(id int64) error {
	return r.db.Delete(&model.MCPService{}, id).Error
}

// UpdateToolsSchema updates tools schema cache.
func (r *MCPRepository) UpdateToolsSchema(id int64, schema []byte) error {
	return r.db.Model(&model.MCPService{}).Where("id = ?", id).Update("tools_schema", schema).Error
}

// ListEnabledServices returns all enabled services.
func (r *MCPRepository) ListEnabledServices() ([]model.MCPService, error) {
	var services []model.MCPService
	err := r.db.Where("status = ?", model.MCPServiceEnabled).Find(&services).Error
	return services, err
}

// CreateAccessRule creates an access rule.
func (r *MCPRepository) CreateAccessRule(rule *model.MCPAccessRule) error {
	return r.db.Create(rule).Error
}

// GetAccessRule returns an access rule.
func (r *MCPRepository) GetAccessRule(serviceID int64, targetType string, targetID int64) (*model.MCPAccessRule, error) {
	var rule model.MCPAccessRule
	err := r.db.Where("service_id = ? AND target_type = ? AND target_id = ?", serviceID, targetType, targetID).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpsertAccessRule creates or updates an access rule.
func (r *MCPRepository) UpsertAccessRule(rule *model.MCPAccessRule) error {
	existing, err := r.GetAccessRule(rule.ServiceID, rule.TargetType, rule.TargetID)
	if err != nil {
		return r.db.Create(rule).Error
	}
	existing.Allowed = rule.Allowed
	return r.db.Save(existing).Error
}

// ListAccessRules returns access rules list.
func (r *MCPRepository) ListAccessRules(serviceID int64) ([]model.MCPAccessRule, error) {
	var rules []model.MCPAccessRule
	query := r.db.Preload("Service")
	if serviceID > 0 {
		query = query.Where("service_id = ?", serviceID)
	}
	err := query.Order("created_at DESC").Find(&rules).Error
	return rules, err
}

// DeleteAccessRule deletes an access rule.
func (r *MCPRepository) DeleteAccessRule(id int64) error {
	return r.db.Delete(&model.MCPAccessRule{}, id).Error
}

// CheckAccess checks if user has access to specified service.
// Priority: user rule > department rule > role rule > default allow.
func (r *MCPRepository) CheckAccess(serviceID, userID int64, deptID *int64, role string) bool {
	var userRule model.MCPAccessRule
	if r.db.Where("service_id = ? AND target_type = ? AND target_id = ?",
		serviceID, model.MCPTargetUser, userID).First(&userRule).Error == nil {
		return userRule.Allowed
	}

	if deptID != nil {
		var deptRule model.MCPAccessRule
		if r.db.Where("service_id = ? AND target_type = ? AND target_id = ?",
			serviceID, model.MCPTargetDepartment, *deptID).First(&deptRule).Error == nil {
			return deptRule.Allowed
		}
	}

	var roleRule model.MCPAccessRule
	if r.db.Where("service_id = ? AND target_type = ? AND target_id = 0",
		serviceID, model.MCPTargetRole).First(&roleRule).Error == nil {
		return roleRule.Allowed
	}

	return true
}

// DeleteAccessRulesByService deletes all access rules for a service.
func (r *MCPRepository) DeleteAccessRulesByService(serviceID int64) error {
	return r.db.Where("service_id = ?", serviceID).Delete(&model.MCPAccessRule{}).Error
}
