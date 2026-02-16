package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SystemRepository 系统配置数据访问层
type SystemRepository struct {
	db *gorm.DB
}

// NewSystemRepository 创建系统配置 Repository
func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

// GetByKey 根据配置键获取值
func (r *SystemRepository) GetByKey(key string) (*model.SystemConfig, error) {
	var cfg model.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ListAll 获取所有系统配置
func (r *SystemRepository) ListAll() ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	err := r.db.Order("config_key ASC").Find(&configs).Error
	return configs, err
}

// Upsert 创建或更新配置
func (r *SystemRepository) Upsert(config *model.SystemConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"config_value", "description", "updated_at"}),
	}).Create(config).Error
}

// BatchUpsert 批量创建或更新配置
func (r *SystemRepository) BatchUpsert(configs []model.SystemConfig) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range configs {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "config_key"}},
				DoUpdates: clause.AssignmentColumns([]string{"config_value", "updated_at"}),
			}).Create(&configs[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete 删除配置
func (r *SystemRepository) Delete(key string) error {
	return r.db.Where("config_key = ?", key).Delete(&model.SystemConfig{}).Error
}
