package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
)

// AnnouncementRepository 公告数据访问层
type AnnouncementRepository struct {
	db *gorm.DB
}

// NewAnnouncementRepository 创建公告 Repository
func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

// Create 创建公告
func (r *AnnouncementRepository) Create(ann *model.Announcement) error {
	return r.db.Create(ann).Error
}

// FindByID 根据 ID 查找公告
func (r *AnnouncementRepository) FindByID(id int64) (*model.Announcement, error) {
	var ann model.Announcement
	err := r.db.Preload("Author").First(&ann, id).Error
	if err != nil {
		return nil, err
	}
	return &ann, nil
}

// Update 更新公告
func (r *AnnouncementRepository) Update(ann *model.Announcement) error {
	return r.db.Save(ann).Error
}

// UpdateFields 更新指定字段
func (r *AnnouncementRepository) UpdateFields(id int64, fields map[string]interface{}) error {
	return r.db.Model(&model.Announcement{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除公告
func (r *AnnouncementRepository) Delete(id int64) error {
	return r.db.Delete(&model.Announcement{}, id).Error
}

// ListPublished 查询已发布的公告（前台展示）
func (r *AnnouncementRepository) ListPublished() ([]model.Announcement, error) {
	var anns []model.Announcement
	err := r.db.Preload("Author").
		Where("status = 1").
		Order("pinned DESC, created_at DESC").
		Find(&anns).Error
	return anns, err
}

// ListAll 查询所有公告（管理后台）
func (r *AnnouncementRepository) ListAll() ([]model.Announcement, error) {
	var anns []model.Announcement
	err := r.db.Preload("Author").
		Order("pinned DESC, created_at DESC").
		Find(&anns).Error
	return anns, err
}
