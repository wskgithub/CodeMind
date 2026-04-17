package repository

import (
	"time"

	"codemind/internal/model"

	"gorm.io/gorm"
)

// DocumentRepository defines the document repository interface.
type DocumentRepository interface {
	List() ([]model.DocumentListItem, error)
	ListAll() ([]model.Document, error)
	GetBySlug(slug string) (*model.Document, error)
	GetByID(id int64) (*model.Document, error)
	Create(doc *model.Document) error
	Update(doc *model.Document) error
	Delete(id int64) error
}

type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository creates a new document repository instance.
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

// List returns published documents with minimal fields.
func (r *documentRepository) List() ([]model.DocumentListItem, error) {
	var items []model.DocumentListItem
	result := r.db.Model(&model.Document{}).
		Where("is_published = ? AND deleted_at IS NULL", true).
		Order("sort_order ASC, id ASC").
		Find(&items)
	return items, result.Error
}

// ListAll returns all documents for admin use, without body content.
func (r *documentRepository) ListAll() ([]model.Document, error) {
	var docs []model.Document
	result := r.db.Select("id, slug, title, subtitle, icon, sort_order, is_published, created_at, updated_at, deleted_at").
		Where("deleted_at IS NULL").
		Order("sort_order ASC, id ASC").
		Find(&docs)
	return docs, result.Error
}

// GetBySlug retrieves a published document by its slug.
func (r *documentRepository) GetBySlug(slug string) (*model.Document, error) {
	var doc model.Document
	result := r.db.Where("slug = ? AND is_published = ? AND deleted_at IS NULL", slug, true).
		First(&doc)
	if result.Error != nil {
		return nil, result.Error
	}
	return &doc, nil
}

// GetByID retrieves a document by ID, including unpublished ones.
func (r *documentRepository) GetByID(id int64) (*model.Document, error) {
	var doc model.Document
	result := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&doc)
	if result.Error != nil {
		return nil, result.Error
	}
	return &doc, nil
}

// Create creates a new document.
func (r *documentRepository) Create(doc *model.Document) error {
	return r.db.Create(doc).Error
}

// Update updates an existing document.
func (r *documentRepository) Update(doc *model.Document) error {
	return r.db.Save(doc).Error
}

// Delete soft-deletes a document.
func (r *documentRepository) Delete(id int64) error {
	now := time.Now()
	return r.db.Model(&model.Document{}).
		Where("id = ?", id).
		Update("deleted_at", &now).Error
}
