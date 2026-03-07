package repository

import (
	"time"

	"codemind/internal/model"

	"gorm.io/gorm"
)

// TrainingDataRepository 训练数据数据访问层
type TrainingDataRepository struct {
	db *gorm.DB
}

// NewTrainingDataRepository 创建训练数据 Repository
func NewTrainingDataRepository(db *gorm.DB) *TrainingDataRepository {
	return &TrainingDataRepository{db: db}
}

// TrainingDataFilter 训练数据查询筛选条件
type TrainingDataFilter struct {
	UserID      *int64
	Model       string
	RequestType string
	StartDate   *time.Time
	EndDate     *time.Time
	IsExcluded  *bool
	Page        int
	PageSize    int
}

// Create 写入单条训练数据记录
func (r *TrainingDataRepository) Create(data *model.LLMTrainingData) error {
	return r.db.Create(data).Error
}

// GetByID 根据 ID 查询单条记录（含完整请求/响应体）
func (r *TrainingDataRepository) GetByID(id int64) (*model.LLMTrainingData, error) {
	var data model.LLMTrainingData
	err := r.db.Where("id = ?", id).First(&data).Error
	return &data, err
}

// List 分页查询训练数据列表（不含大字段 request_body / response_body）
func (r *TrainingDataRepository) List(filter TrainingDataFilter) ([]model.LLMTrainingDataListItem, int64, error) {
	query := r.db.Table("llm_training_data").
		Select(`llm_training_data.id, llm_training_data.user_id, users.username,
			llm_training_data.request_type, llm_training_data.model, llm_training_data.is_stream,
			llm_training_data.prompt_tokens, llm_training_data.completion_tokens, llm_training_data.total_tokens,
			llm_training_data.duration_ms, llm_training_data.status_code,
			llm_training_data.is_excluded, llm_training_data.created_at`).
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

// UpdateExcluded 更新记录的排除状态
func (r *TrainingDataRepository) UpdateExcluded(id int64, excluded bool) error {
	return r.db.Model(&model.LLMTrainingData{}).
		Where("id = ?", id).
		Update("is_excluded", excluded).Error
}

// BatchIterator 返回按批次读取的游标查询（用于大数据量导出）
func (r *TrainingDataRepository) BatchIterator(filter TrainingDataFilter, batchSize int, fn func(batch []model.LLMTrainingData) error) error {
	query := r.db.Model(&model.LLMTrainingData{})
	query = r.applyFilter(query, filter)

	// 强制排除已标记的记录
	query = query.Where("is_excluded = FALSE")

	return query.Order("id ASC").FindInBatches(&[]model.LLMTrainingData{}, batchSize, func(tx *gorm.DB, batch int) error {
		var records []model.LLMTrainingData
		if err := tx.Scan(&records).Error; err != nil {
			return err
		}
		return fn(records)
	}).Error
}

// GetStats 获取训练数据统计信息
func (r *TrainingDataRepository) GetStats() (*model.TrainingDataStats, error) {
	stats := &model.TrainingDataStats{}

	// 总记录数
	r.db.Model(&model.LLMTrainingData{}).Count(&stats.TotalCount)

	// 今日新增
	today := time.Now().Format("2006-01-02")
	r.db.Model(&model.LLMTrainingData{}).
		Where("created_at >= ?::date", today).
		Count(&stats.TodayCount)

	// 已排除数
	r.db.Model(&model.LLMTrainingData{}).
		Where("is_excluded = TRUE").
		Count(&stats.ExcludedCount)

	// 模型分布
	r.db.Model(&model.LLMTrainingData{}).
		Select("model, COUNT(*) as count").
		Group("model").
		Order("count DESC").
		Limit(20).
		Scan(&stats.ModelDistribution)

	return stats, nil
}

// applyFilter 应用筛选条件到查询
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
