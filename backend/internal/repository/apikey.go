package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// APIKeyRepository handles API key data access
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates an API key repository
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(key *model.APIKey) error {
	return r.db.Create(key).Error
}

// FindByID finds an API key by ID
func (r *APIKeyRepository) FindByID(id int64) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.Preload("User").First(&key, id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// FindByHash finds an API key by hash
func (r *APIKeyRepository) FindByHash(hash string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.Preload("User").Where("key_hash = ?", hash).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// ListByUserID returns all API keys for a user
func (r *APIKeyRepository) ListByUserID(userID int64) ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

// CountByUserID counts API keys for a user
func (r *APIKeyRepository) CountByUserID(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.APIKey{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountAll counts all API keys
func (r *APIKeyRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.APIKey{}).Count(&count).Error
	return count, err
}

// UpdateStatus updates API key status
func (r *APIKeyRepository) UpdateStatus(id int64, status int16) error {
	return r.db.Model(&model.APIKey{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateLastUsed updates last used timestamp
func (r *APIKeyRepository) UpdateLastUsed(id int64) error {
	return r.db.Model(&model.APIKey{}).Where("id = ?", id).
		Update("last_used_at", gorm.Expr("NOW()")).Error
}

// Delete deletes an API key
func (r *APIKeyRepository) Delete(id int64) error {
	return r.db.Delete(&model.APIKey{}, id).Error
}
