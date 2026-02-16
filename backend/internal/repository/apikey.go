package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// APIKeyRepository API Key 数据访问层
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository 创建 API Key Repository
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create 创建 API Key
func (r *APIKeyRepository) Create(key *model.APIKey) error {
	return r.db.Create(key).Error
}

// FindByID 根据 ID 查找 API Key
func (r *APIKeyRepository) FindByID(id int64) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.Preload("User").First(&key, id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// FindByHash 根据哈希值查找 API Key
func (r *APIKeyRepository) FindByHash(hash string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.Preload("User").Where("key_hash = ?", hash).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// ListByUserID 查询用户的所有 API Key
func (r *APIKeyRepository) ListByUserID(userID int64) ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

// CountByUserID 统计用户的 API Key 数量
func (r *APIKeyRepository) CountByUserID(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.APIKey{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountAll 统计所有 API Key 数量
func (r *APIKeyRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.APIKey{}).Count(&count).Error
	return count, err
}

// UpdateStatus 更新 Key 状态
func (r *APIKeyRepository) UpdateStatus(id int64, status int16) error {
	return r.db.Model(&model.APIKey{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateLastUsed 更新最后使用时间
func (r *APIKeyRepository) UpdateLastUsed(id int64) error {
	return r.db.Model(&model.APIKey{}).Where("id = ?", id).
		Update("last_used_at", gorm.Expr("NOW()")).Error
}

// Delete 删除 API Key
func (r *APIKeyRepository) Delete(id int64) error {
	return r.db.Delete(&model.APIKey{}, id).Error
}
