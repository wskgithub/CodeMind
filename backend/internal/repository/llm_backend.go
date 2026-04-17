package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// LLMBackendRepository handles LLM backend node data access.
type LLMBackendRepository struct {
	db *gorm.DB
}

// NewLLMBackendRepository creates a new repository.
func NewLLMBackendRepository(db *gorm.DB) *LLMBackendRepository {
	return &LLMBackendRepository{db: db}
}

// Create creates a backend node.
func (r *LLMBackendRepository) Create(backend *model.LLMBackend) error {
	return r.db.Create(backend).Error
}

// FindByID returns backend by ID.
func (r *LLMBackendRepository) FindByID(id int64) (*model.LLMBackend, error) {
	var b model.LLMBackend
	err := r.db.First(&b, id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// FindByName returns backend by name.
func (r *LLMBackendRepository) FindByName(name string) (*model.LLMBackend, error) {
	var b model.LLMBackend
	err := r.db.Where("name = ?", name).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ListAll returns all backend nodes.
func (r *LLMBackendRepository) ListAll() ([]model.LLMBackend, error) {
	var backends []model.LLMBackend
	err := r.db.Order("weight DESC, id ASC").Find(&backends).Error
	return backends, err
}

// ListEnabled returns all enabled backend nodes.
func (r *LLMBackendRepository) ListEnabled() ([]model.LLMBackend, error) {
	var backends []model.LLMBackend
	err := r.db.Where("status = ?", model.LLMBackendEnabled).
		Order("weight DESC, id ASC").Find(&backends).Error
	return backends, err
}

// Update updates a backend node.
func (r *LLMBackendRepository) Update(backend *model.LLMBackend) error {
	return r.db.Save(backend).Error
}

// Delete deletes a backend node.
func (r *LLMBackendRepository) Delete(id int64) error {
	return r.db.Delete(&model.LLMBackend{}, id).Error
}

// CountAll returns total backend node count.
func (r *LLMBackendRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.LLMBackend{}).Count(&count).Error
	return count, err
}

// CountEnabled returns enabled backend node count.
func (r *LLMBackendRepository) CountEnabled() (int64, error) {
	var count int64
	err := r.db.Model(&model.LLMBackend{}).Where("status = ?", model.LLMBackendEnabled).Count(&count).Error
	return count, err
}
