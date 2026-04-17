package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// ThirdPartyProviderRepository handles third-party provider data access.
type ThirdPartyProviderRepository struct {
	db *gorm.DB
}

// NewThirdPartyProviderRepository creates a new repository.
func NewThirdPartyProviderRepository(db *gorm.DB) *ThirdPartyProviderRepository {
	return &ThirdPartyProviderRepository{db: db}
}

// CreateTemplate creates a template.
func (r *ThirdPartyProviderRepository) CreateTemplate(template *model.ThirdPartyProviderTemplate) error {
	return r.db.Create(template).Error
}

// GetTemplateByID returns template by ID.
func (r *ThirdPartyProviderRepository) GetTemplateByID(id int64) (*model.ThirdPartyProviderTemplate, error) {
	var template model.ThirdPartyProviderTemplate
	err := r.db.Where("id = ?", id).First(&template).Error
	return &template, err
}

// ListTemplates returns all templates (admin view, including disabled).
func (r *ThirdPartyProviderRepository) ListTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	var list []model.ThirdPartyProviderTemplate
	err := r.db.Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// ListActiveTemplates returns all enabled templates (user view).
func (r *ThirdPartyProviderRepository) ListActiveTemplates() ([]model.ThirdPartyProviderTemplate, error) {
	var list []model.ThirdPartyProviderTemplate
	err := r.db.Where("status = ?", model.StatusEnabled).Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// UpdateTemplate updates a template.
func (r *ThirdPartyProviderRepository) UpdateTemplate(template *model.ThirdPartyProviderTemplate) error {
	return r.db.Save(template).Error
}

// DeleteTemplate soft deletes a template.
func (r *ThirdPartyProviderRepository) DeleteTemplate(id int64) error {
	return r.db.Delete(&model.ThirdPartyProviderTemplate{}, id).Error
}

// ExistsTemplateName checks if template name exists (excluding specific ID).
func (r *ThirdPartyProviderRepository) ExistsTemplateName(name string, excludeID int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.ThirdPartyProviderTemplate{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// CreateProvider creates a user third-party provider.
func (r *ThirdPartyProviderRepository) CreateProvider(provider *model.UserThirdPartyProvider) error {
	return r.db.Create(provider).Error
}

// GetProviderByID returns user third-party provider by ID.
func (r *ThirdPartyProviderRepository) GetProviderByID(id int64) (*model.UserThirdPartyProvider, error) {
	var provider model.UserThirdPartyProvider
	err := r.db.Where("id = ?", id).First(&provider).Error
	return &provider, err
}

// ListProvidersByUserID returns all third-party providers for a user.
func (r *ThirdPartyProviderRepository) ListProvidersByUserID(userID int64) ([]model.UserThirdPartyProvider, error) {
	var list []model.UserThirdPartyProvider
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	return list, err
}

// ListActiveProvidersByUserID returns all enabled providers for a user.
func (r *ThirdPartyProviderRepository) ListActiveProvidersByUserID(userID int64) ([]model.UserThirdPartyProvider, error) {
	var list []model.UserThirdPartyProvider
	err := r.db.Where("user_id = ? AND status = ?", userID, model.StatusEnabled).Find(&list).Error
	return list, err
}

// UpdateProvider updates a user third-party provider.
func (r *ThirdPartyProviderRepository) UpdateProvider(provider *model.UserThirdPartyProvider) error {
	return r.db.Save(provider).Error
}

// UpdateProviderStatus updates provider status.
func (r *ThirdPartyProviderRepository) UpdateProviderStatus(id int64, status int16) error {
	return r.db.Model(&model.UserThirdPartyProvider{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteProvider soft deletes a user third-party provider.
func (r *ThirdPartyProviderRepository) DeleteProvider(id int64) error {
	return r.db.Delete(&model.UserThirdPartyProvider{}, id).Error
}

// ExistsProviderName checks if user has provider with same name (excluding specific ID).
func (r *ThirdPartyProviderRepository) ExistsProviderName(userID int64, name string, excludeID int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.UserThirdPartyProvider{}).Where("user_id = ? AND name = ?", userID, name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// CreateThirdPartyUsage creates a third-party usage record.
func (r *ThirdPartyProviderRepository) CreateThirdPartyUsage(usage *model.ThirdPartyTokenUsage) error {
	return r.db.Create(usage).Error
}

// GetThirdPartyUsageSummary returns user's third-party usage summary.
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
