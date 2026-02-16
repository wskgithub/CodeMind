package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// MCPRepository MCP 服务数据访问层
type MCPRepository struct {
	db *gorm.DB
}

// NewMCPRepository 创建 MCP 数据仓库
func NewMCPRepository(db *gorm.DB) *MCPRepository {
	return &MCPRepository{db: db}
}

// ──────────────────────────────────
// MCP 服务 CRUD
// ──────────────────────────────────

// CreateService 创建 MCP 服务
func (r *MCPRepository) CreateService(svc *model.MCPService) error {
	return r.db.Create(svc).Error
}

// GetServiceByID 根据 ID 获取服务
func (r *MCPRepository) GetServiceByID(id int64) (*model.MCPService, error) {
	var svc model.MCPService
	err := r.db.First(&svc, id).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// GetServiceByName 根据名称获取服务
func (r *MCPRepository) GetServiceByName(name string) (*model.MCPService, error) {
	var svc model.MCPService
	err := r.db.Where("name = ?", name).First(&svc).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// ListServices 获取服务列表
func (r *MCPRepository) ListServices(status string) ([]model.MCPService, error) {
	var services []model.MCPService
	query := r.db.Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&services).Error
	return services, err
}

// UpdateService 更新服务
func (r *MCPRepository) UpdateService(svc *model.MCPService) error {
	return r.db.Save(svc).Error
}

// DeleteService 软删除服务
func (r *MCPRepository) DeleteService(id int64) error {
	return r.db.Delete(&model.MCPService{}, id).Error
}

// UpdateToolsSchema 更新工具列表缓存
func (r *MCPRepository) UpdateToolsSchema(id int64, schema []byte) error {
	return r.db.Model(&model.MCPService{}).Where("id = ?", id).Update("tools_schema", schema).Error
}

// ListEnabledServices 获取所有启用的服务
func (r *MCPRepository) ListEnabledServices() ([]model.MCPService, error) {
	var services []model.MCPService
	err := r.db.Where("status = ?", model.MCPServiceEnabled).Find(&services).Error
	return services, err
}

// ──────────────────────────────────
// MCP 访问规则 CRUD
// ──────────────────────────────────

// CreateAccessRule 创建访问规则
func (r *MCPRepository) CreateAccessRule(rule *model.MCPAccessRule) error {
	return r.db.Create(rule).Error
}

// GetAccessRule 获取访问规则
func (r *MCPRepository) GetAccessRule(serviceID int64, targetType string, targetID int64) (*model.MCPAccessRule, error) {
	var rule model.MCPAccessRule
	err := r.db.Where("service_id = ? AND target_type = ? AND target_id = ?", serviceID, targetType, targetID).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpsertAccessRule 创建或更新访问规则
func (r *MCPRepository) UpsertAccessRule(rule *model.MCPAccessRule) error {
	existing, err := r.GetAccessRule(rule.ServiceID, rule.TargetType, rule.TargetID)
	if err != nil {
		// 不存在则创建
		return r.db.Create(rule).Error
	}
	// 存在则更新
	existing.Allowed = rule.Allowed
	return r.db.Save(existing).Error
}

// ListAccessRules 获取访问规则列表
func (r *MCPRepository) ListAccessRules(serviceID int64) ([]model.MCPAccessRule, error) {
	var rules []model.MCPAccessRule
	query := r.db.Preload("Service")
	if serviceID > 0 {
		query = query.Where("service_id = ?", serviceID)
	}
	err := query.Order("created_at DESC").Find(&rules).Error
	return rules, err
}

// DeleteAccessRule 删除访问规则
func (r *MCPRepository) DeleteAccessRule(id int64) error {
	return r.db.Delete(&model.MCPAccessRule{}, id).Error
}

// CheckAccess 检查用户是否有权访问指定服务
// 检查优先级: 用户规则 > 部门规则 > 角色规则 > 默认允许
func (r *MCPRepository) CheckAccess(serviceID, userID int64, deptID *int64, role string) bool {
	// 检查用户级规则
	var userRule model.MCPAccessRule
	if r.db.Where("service_id = ? AND target_type = ? AND target_id = ?",
		serviceID, model.MCPTargetUser, userID).First(&userRule).Error == nil {
		return userRule.Allowed
	}

	// 检查部门级规则
	if deptID != nil {
		var deptRule model.MCPAccessRule
		if r.db.Where("service_id = ? AND target_type = ? AND target_id = ?",
			serviceID, model.MCPTargetDepartment, *deptID).First(&deptRule).Error == nil {
			return deptRule.Allowed
		}
	}

	// 检查角色级规则（target_id 用 0 表示所有该角色用户）
	var roleRule model.MCPAccessRule
	if r.db.Where("service_id = ? AND target_type = ? AND target_id = 0",
		serviceID, model.MCPTargetRole).First(&roleRule).Error == nil {
		return roleRule.Allowed
	}

	// 默认允许
	return true
}

// DeleteAccessRulesByService 删除服务的所有访问规则
func (r *MCPRepository) DeleteAccessRulesByService(serviceID int64) error {
	return r.db.Where("service_id = ?", serviceID).Delete(&model.MCPAccessRule{}).Error
}
