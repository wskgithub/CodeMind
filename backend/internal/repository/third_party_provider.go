package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// ThirdPartyProviderRepository 第三方模型服务数据访问层
type ThirdPartyProviderRepository struct {
	db *gorm.DB
}

// NewThirdPartyProviderRepository 创建 Repository 实例
func NewThirdPartyProviderRepository(db *gorm.DB) *ThirdPartyProviderRepository {
	return &ThirdPartyProviderRepository{db: db}
}

// ──────────────────────────────────
// 模板管理（管理员）
// ──────────────────────────────────

// CreateTemplate 创建模板
func (r *ThirdPartyProviderRepository) CreateTemplate(template *model.ThirdPartyProviderTemplate) error {
	return r.db.Create(template).Error
}

// GetTemplateByID 根据 ID 获取模板
func (r *ThirdPartyProviderRepository) GetTemplateByID(id int64) (*model.ThirdPartyProviderTemplate, error) {
	var template model.ThirdPartyProviderTemplate
	err := r.db.Where("id = ?", id).First(&template).Error
	return &template, err
}

// ListTemplates 获取所有模板（管理员视图，含禁用项）
func (r *ThirdPartyProviderRepository) ListTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	var list []model.ThirdPartyProviderTemplate
	err := r.db.Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// ListActiveTemplates 获取所有启用的模板（用户选择视图）
func (r *ThirdPartyProviderRepository) ListActiveTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	var list []model.ThirdPartyProviderTemplate
	err := r.db.Where("status = ?", model.StatusEnabled).Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// UpdateTemplate 更新模板
func (r *ThirdPartyProviderRepository) UpdateTemplate(template *model.ThirdPartyProviderTemplate) error {
	return r.db.Save(template).Error
}

// DeleteTemplate 软删除模板
func (r *ThirdPartyProviderRepository) DeleteTemplate(id int64) error {
	return r.db.Delete(&model.ThirdPartyProviderTemplate{}, id).Error
}

// ExistsTemplateName 检查模板名称是否已存在（排除指定 ID）
func (r *ThirdPartyProviderRepository) ExistsTemplateName(name string, excludeID int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.ThirdPartyProviderTemplate{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// ──────────────────────────────────
// 用户第三方服务
// ──────────────────────────────────

// CreateProvider 创建用户第三方服务
func (r *ThirdPartyProviderRepository) CreateProvider(provider *model.UserThirdPartyProvider) error {
	return r.db.Create(provider).Error
}

// GetProviderByID 根据 ID 获取用户第三方服务
func (r *ThirdPartyProviderRepository) GetProviderByID(id int64) (*model.UserThirdPartyProvider, error) {
	var provider model.UserThirdPartyProvider
	err := r.db.Where("id = ?", id).First(&provider).Error
	return &provider, err
}

// ListProvidersByUserID 获取用户的所有第三方服务
func (r *ThirdPartyProviderRepository) ListProvidersByUserID(userID int64) ([]model.UserThirdPartyProvider, error) {
	var list []model.UserThirdPartyProvider
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	return list, err
}

// ListActiveProvidersByUserID 获取用户所有启用的第三方服务
func (r *ThirdPartyProviderRepository) ListActiveProvidersByUserID(userID int64) ([]model.UserThirdPartyProvider, error) {
	var list []model.UserThirdPartyProvider
	err := r.db.Where("user_id = ? AND status = ?", userID, model.StatusEnabled).Find(&list).Error
	return list, err
}

// UpdateProvider 更新用户第三方服务
func (r *ThirdPartyProviderRepository) UpdateProvider(provider *model.UserThirdPartyProvider) error {
	return r.db.Save(provider).Error
}

// UpdateProviderStatus 更新服务状态
func (r *ThirdPartyProviderRepository) UpdateProviderStatus(id int64, status int16) error {
	return r.db.Model(&model.UserThirdPartyProvider{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteProvider 软删除用户第三方服务
func (r *ThirdPartyProviderRepository) DeleteProvider(id int64) error {
	return r.db.Delete(&model.UserThirdPartyProvider{}, id).Error
}

// ExistsProviderName 检查用户是否已有同名服务（排除指定 ID）
func (r *ThirdPartyProviderRepository) ExistsProviderName(userID int64, name string, excludeID int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.UserThirdPartyProvider{}).Where("user_id = ? AND name = ?", userID, name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// ──────────────────────────────────
// 第三方用量记录
// ──────────────────────────────────

// CreateThirdPartyUsage 写入第三方用量记录
func (r *ThirdPartyProviderRepository) CreateThirdPartyUsage(usage *model.ThirdPartyTokenUsage) error {
	return r.db.Create(usage).Error
}

// GetThirdPartyUsageSummary 获取用户第三方服务用量汇总
func (r *ThirdPartyProviderRepository) GetThirdPartyUsageSummary(userID int64, providerID *int64) (int64, int64, error) {
	query := r.db.Model(&model.ThirdPartyTokenUsage{}).Where("user_id = ?", userID)
	if providerID != nil {
		query = query.Where("provider_id = ?", *providerID)
	}

	var totalTokens int64
	var requestCount int64
	err := query.Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokens).Error
	if err != nil {
		return 0, 0, err
	}
	err = query.Count(&requestCount).Error
	return totalTokens, requestCount, err
}
