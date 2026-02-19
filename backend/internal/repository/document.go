package repository

import (
	"codemind/internal/model"
	"time"

	"gorm.io/gorm"
)

// DocumentRepository 文档仓库接口
type DocumentRepository interface {
	List() ([]model.DocumentListItem, error)
	ListAll() ([]model.Document, error)
	GetBySlug(slug string) (*model.Document, error)
	GetByID(id int64) (*model.Document, error)
	Create(doc *model.Document) error
	Update(doc *model.Document) error
	Delete(id int64) error
	BatchCreate(docs []model.Document) error
}

// documentRepository 文档仓库实现
type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建文档仓库实例
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

// List 获取已发布的文档列表（简要信息）
func (r *documentRepository) List() ([]model.DocumentListItem, error) {
	var items []model.DocumentListItem
	result := r.db.Model(&model.Document{}).
		Where("is_published = ? AND deleted_at IS NULL", true).
		Order("sort_order ASC, id ASC").
		Find(&items)
	return items, result.Error
}

// ListAll 获取所有文档（包括未发布，管理用）
func (r *documentRepository) ListAll() ([]model.Document, error) {
	var docs []model.Document
	result := r.db.Where("deleted_at IS NULL").
		Order("sort_order ASC, id ASC").
		Find(&docs)
	return docs, result.Error
}

// GetBySlug 根据 slug 获取文档
func (r *documentRepository) GetBySlug(slug string) (*model.Document, error) {
	var doc model.Document
	result := r.db.Where("slug = ? AND is_published = ? AND deleted_at IS NULL", slug, true).
		First(&doc)
	if result.Error != nil {
		return nil, result.Error
	}
	return &doc, nil
}

// GetByID 根据 ID 获取文档（管理用，可获取未发布文档）
func (r *documentRepository) GetByID(id int64) (*model.Document, error) {
	var doc model.Document
	result := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&doc)
	if result.Error != nil {
		return nil, result.Error
	}
	return &doc, nil
}

// Create 创建文档
func (r *documentRepository) Create(doc *model.Document) error {
	return r.db.Create(doc).Error
}

// Update 更新文档
func (r *documentRepository) Update(doc *model.Document) error {
	return r.db.Save(doc).Error
}

// Delete 软删除文档
func (r *documentRepository) Delete(id int64) error {
	now := time.Now()
	return r.db.Model(&model.Document{}).
		Where("id = ?", id).
		Update("deleted_at", &now).Error
}

// BatchCreate 批量创建文档（初始化用）
func (r *documentRepository) BatchCreate(docs []model.Document) error {
	return r.db.Create(&docs).Error
}
