package repository

import (
	"codemind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SystemRepository handles system config data access.
type SystemRepository struct {
	db *gorm.DB
}

// NewSystemRepository creates a new system config repository.
func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

// GetByKey returns config value by key.
func (r *SystemRepository) GetByKey(key string) (*model.SystemConfig, error) {
	var cfg model.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ListAll returns all system configs.
func (r *SystemRepository) ListAll() ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	err := r.db.Order("config_key ASC").Find(&configs).Error
	return configs, err
}

// Upsert creates or updates config.
func (r *SystemRepository) Upsert(config *model.SystemConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"config_value", "description", "updated_at"}),
	}).Create(config).Error
}

// BatchUpsert batch creates or updates configs.
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

// Delete removes config.
func (r *SystemRepository) Delete(key string) error {
	return r.db.Where("config_key = ?", key).Delete(&model.SystemConfig{}).Error
}
