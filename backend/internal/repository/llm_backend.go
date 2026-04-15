package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// LLMBackendRepository LLM 后端服务节点数据访问层
type LLMBackendRepository struct {
	db *gorm.DB
}

// NewLLMBackendRepository 创建仓储实例
func NewLLMBackendRepository(db *gorm.DB) *LLMBackendRepository {
	return &LLMBackendRepository{db: db}
}

// Create 创建后端节点
func (r *LLMBackendRepository) Create(backend *model.LLMBackend) error {
	return r.db.Create(backend).Error
}

// FindByID 根据 ID 查找
func (r *LLMBackendRepository) FindByID(id int64) (*model.LLMBackend, error) {
	var b model.LLMBackend
	err := r.db.First(&b, id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// FindByName 根据名称查找
func (r *LLMBackendRepository) FindByName(name string) (*model.LLMBackend, error) {
	var b model.LLMBackend
	err := r.db.Where("name = ?", name).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ListAll 查询所有后端节点
func (r *LLMBackendRepository) ListAll() ([]model.LLMBackend, error) {
	var backends []model.LLMBackend
	err := r.db.Order("weight DESC, id ASC").Find(&backends).Error
	return backends, err
}

// ListEnabled 查询所有启用的后端节点
func (r *LLMBackendRepository) ListEnabled() ([]model.LLMBackend, error) {
	var backends []model.LLMBackend
	err := r.db.Where("status = ?", model.LLMBackendEnabled).
		Order("weight DESC, id ASC").Find(&backends).Error
	return backends, err
}

// Update 更新后端节点
func (r *LLMBackendRepository) Update(backend *model.LLMBackend) error {
	return r.db.Save(backend).Error
}

// Delete 删除后端节点
func (r *LLMBackendRepository) Delete(id int64) error {
	return r.db.Delete(&model.LLMBackend{}, id).Error
}

// CountAll 查询所有后端节点数量
func (r *LLMBackendRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.LLMBackend{}).Count(&count).Error
	return count, err
}

// CountEnabled 查询启用的后端节点数量
func (r *LLMBackendRepository) CountEnabled() (int64, error) {
	var count int64
	err := r.db.Model(&model.LLMBackend{}).Where("status = ?", model.LLMBackendEnabled).Count(&count).Error
	return count, err
}
