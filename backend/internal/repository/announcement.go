// Package repository provides data access layer for all entities.
package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// AnnouncementRepository handles announcement data access.
type AnnouncementRepository struct {
	db *gorm.DB
}

// NewAnnouncementRepository creates a new announcement repository.
func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

// Create creates an announcement.
func (r *AnnouncementRepository) Create(ann *model.Announcement) error {
	return r.db.Create(ann).Error
}

// FindByID returns announcement by ID.
func (r *AnnouncementRepository) FindByID(id int64) (*model.Announcement, error) {
	var ann model.Announcement
	err := r.db.Preload("Author").First(&ann, id).Error
	if err != nil {
		return nil, err
	}
	return &ann, nil
}

// Update updates an announcement.
func (r *AnnouncementRepository) Update(ann *model.Announcement) error {
	return r.db.Save(ann).Error
}

// UpdateFields updates specified fields.
func (r *AnnouncementRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Announcement{}).Where("id = ?", id).Updates(fields).Error
}

// Delete deletes an announcement.
func (r *AnnouncementRepository) Delete(id int64) error {
	return r.db.Delete(&model.Announcement{}, id).Error
}

// ListPublished returns published announcements (for frontend).
func (r *AnnouncementRepository) ListPublished() ([]model.Announcement, error) {
	var anns []model.Announcement
	err := r.db.Preload("Author").
		Where("status = 1").
		Order("pinned DESC, created_at DESC").
		Find(&anns).Error
	return anns, err
}

// ListAll returns all announcements (for admin).
func (r *AnnouncementRepository) ListAll() ([]model.Announcement, error) {
	var anns []model.Announcement
	err := r.db.Preload("Author").
		Order("pinned DESC, created_at DESC").
		Find(&anns).Error
	return anns, err
}
