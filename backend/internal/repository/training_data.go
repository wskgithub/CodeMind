package repository

import (
	"time"

	"codemind/internal/model"
	"codemind/internal/pkg/timezone"

	"gorm.io/gorm"
)

// TrainingDataRepository handles training data access.
type TrainingDataRepository struct {
	db *gorm.DB
}

// NewTrainingDataRepository creates a new TrainingDataRepository.
func NewTrainingDataRepository(db *gorm.DB) *TrainingDataRepository {
	return &TrainingDataRepository{db: db}
}

// TrainingDataFilter defines query filter conditions for training data.
type TrainingDataFilter struct {
	UserID      *int64
	StartDate   *time.Time
	EndDate     *time.Time
	IsExcluded  *bool
	Model       string
	RequestType string
	Page        int
	PageSize    int
}

// Create writes a single training data record.
func (r *TrainingDataRepository) Create(data *model.LLMTrainingData) error {
	return r.db.Create(data).Error
}

// BatchCreate batch inserts training data records.
func (r *TrainingDataRepository) BatchCreate(records []*model.LLMTrainingData) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.CreateInBatches(records, len(records)).Error
}

// GetByID retrieves a single record by ID (with full request/response body).
func (r *TrainingDataRepository) GetByID(id int64) (*model.LLMTrainingData, error) {
	var data model.LLMTrainingData
	err := r.db.Where("id = ?", id).First(&data).Error
	return &data, err
}

// List retrieves paginated training data list (without large body fields).
func (r *TrainingDataRepository) List(filter TrainingDataFilter) ([]model.LLMTrainingDataListItem, int64, error) {
	query := r.db.Table("llm_training_data").
		Select(`llm_training_data.id, llm_training_data.user_id, users.username,
			llm_training_data.request_type, llm_training_data.model, llm_training_data.is_stream,
			llm_training_data.prompt_tokens, llm_training_data.completion_tokens, llm_training_data.total_tokens,
			llm_training_data.duration_ms, llm_training_data.status_code,
			llm_training_data.is_excluded, llm_training_data.created_at,
			llm_training_data.is_sanitized, llm_training_data.conversation_id,
			llm_training_data.content_hash, llm_training_data.quality_score`).
		Joins("LEFT JOIN users ON users.id = llm_training_data.user_id")

	query = r.applyFilter(query, filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.LLMTrainingDataListItem
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Order("llm_training_data.created_at DESC").
		Offset(offset).Limit(filter.PageSize).
		Scan(&items).Error

	return items, total, err
}

// UpdateExcluded updates the exclusion status of a record.
func (r *TrainingDataRepository) UpdateExcluded(id int64, excluded bool) error {
	return r.db.Model(&model.LLMTrainingData{}).
		Where("id = ?", id).
		Update("is_excluded", excluded).Error
}

// BatchIterator reads training data in batches (for large data exports).
func (r *TrainingDataRepository) BatchIterator(filter TrainingDataFilter, batchSize int, fn func(batch []model.LLMTrainingData) error) error {
	query := r.db.Model(&model.LLMTrainingData{})
	query = r.applyFilter(query, filter)
	query = query.Where("is_excluded = FALSE")

	var records []model.LLMTrainingData
	return query.Order("id ASC").FindInBatches(&records, batchSize, func(_ *gorm.DB, _ int) error {
		return fn(records)
	}).Error
}

// GetStats retrieves training data statistics.
func (r *TrainingDataRepository) GetStats() (*model.TrainingDataStats, error) {
	stats := &model.TrainingDataStats{}

	today := timezone.TodayStr()
	var counters struct {
		TotalCount    int64 `gorm:"column:total_count"`
		TodayCount    int64 `gorm:"column:today_count"`
		ExcludedCount int64 `gorm:"column:excluded_count"`
	}
	err := r.db.Model(&model.LLMTrainingData{}).
		Select(`COUNT(*) as total_count,
			COUNT(*) FILTER (WHERE created_at >= ?::date) as today_count,
			COUNT(*) FILTER (WHERE is_excluded = TRUE) as excluded_count`, today).
		Scan(&counters).Error
	if err != nil {
		return nil, err
	}
	stats.TotalCount = counters.TotalCount
	stats.TodayCount = counters.TodayCount
	stats.ExcludedCount = counters.ExcludedCount

	err = r.db.Model(&model.LLMTrainingData{}).
		Select("model, COUNT(*) as count").
		Group("model").
		Order("count DESC").
		Limit(20). //nolint:mnd // top N models
		Scan(&stats.ModelDistribution).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// CountAll returns total training data record count.
func (r *TrainingDataRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.LLMTrainingData{}).Count(&count).Error
	return count, err
}

// GetArchiveBoundaryID returns the archive boundary ID (the n-th record's ID by ascending order).
func (r *TrainingDataRepository) GetArchiveBoundaryID(n int) (int64, error) {
	var id int64
	err := r.db.Model(&model.LLMTrainingData{}).
		Select("id").
		Order("id ASC").
		Offset(n - 1).
		Limit(1).
		Scan(&id).Error
	return id, err
}

// GetIDRange returns min/max IDs within the specified ID range.
func (r *TrainingDataRepository) GetIDRange(maxID int64) (minID, actualMaxID int64, err error) {
	err = r.db.Model(&model.LLMTrainingData{}).
		Select("MIN(id), MAX(id)").
		Where("id <= ?", maxID).
		Row().Scan(&minID, &actualMaxID)
	return
}

// StreamByIDRange reads complete records in batches by ID range (for archive export).
func (r *TrainingDataRepository) StreamByIDRange(maxID int64, batchSize int, fn func(batch []model.LLMTrainingData) error) error {
	var records []model.LLMTrainingData
	return r.db.Where("id <= ?", maxID).
		Order("id ASC").
		FindInBatches(&records, batchSize, func(_ *gorm.DB, _ int) error {
			return fn(records)
		}).Error
}

// DeleteByIDRange batch deletes archived records by ID range.
func (r *TrainingDataRepository) DeleteByIDRange(maxID int64, batchSize int) (int64, error) {
	var totalDeleted int64
	for {
		result := r.db.Where("id <= ?", maxID).
			Limit(batchSize).
			Delete(&model.LLMTrainingData{})
		if result.Error != nil {
			return totalDeleted, result.Error
		}
		totalDeleted += result.RowsAffected
		if result.RowsAffected < int64(batchSize) {
			break
		}
	}
	return totalDeleted, nil
}

// applyFilter applies filter conditions to the query.
func (r *TrainingDataRepository) applyFilter(query *gorm.DB, filter TrainingDataFilter) *gorm.DB {
	if filter.UserID != nil {
		query = query.Where("llm_training_data.user_id = ?", *filter.UserID)
	}
	if filter.Model != "" {
		query = query.Where("llm_training_data.model = ?", filter.Model)
	}
	if filter.RequestType != "" {
		query = query.Where("llm_training_data.request_type = ?", filter.RequestType)
	}
	if filter.StartDate != nil {
		query = query.Where("llm_training_data.created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("llm_training_data.created_at <= ?", *filter.EndDate)
	}
	if filter.IsExcluded != nil {
		query = query.Where("llm_training_data.is_excluded = ?", *filter.IsExcluded)
	}
	return query
}
