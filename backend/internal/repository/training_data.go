package repository

import (
	"time"

	"codemind/internal/model"
	"codemind/internal/pkg/timezone"

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

// BatchCreate 批量写入训练数据记录
// 使用单条 INSERT ... VALUES (...), (...) 语句减少数据库往返次数
func (r *TrainingDataRepository) BatchCreate(records []*model.LLMTrainingData) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.CreateInBatches(records, len(records)).Error
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

// UpdateExcluded 更新记录的排除状态
func (r *TrainingDataRepository) UpdateExcluded(id int64, excluded bool) error {
	return r.db.Model(&model.LLMTrainingData{}).
		Where("id = ?", id).
		Update("is_excluded", excluded).Error
}

// BatchIterator 按批次读取训练数据（用于大数据量导出）
// FindInBatches 自动将结果填充到 records 切片，无需在回调中再次 Scan
func (r *TrainingDataRepository) BatchIterator(filter TrainingDataFilter, batchSize int, fn func(batch []model.LLMTrainingData) error) error {
	query := r.db.Model(&model.LLMTrainingData{})
	query = r.applyFilter(query, filter)
	query = query.Where("is_excluded = FALSE")

	var records []model.LLMTrainingData
	return query.Order("id ASC").FindInBatches(&records, batchSize, func(_ *gorm.DB, _ int) error {
		return fn(records)
	}).Error
}

// GetStats 获取训练数据统计信息
// 将原先 4 次独立查询合并为 2 次，减少数据库往返
func (r *TrainingDataRepository) GetStats() (*model.TrainingDataStats, error) {
	stats := &model.TrainingDataStats{}

	// 一次查询获取总数、今日新增、已排除数
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

	// 模型分布（独立查询，因为需要 GROUP BY）
	err = r.db.Model(&model.LLMTrainingData{}).
		Select("model, COUNT(*) as count").
		Group("model").
		Order("count DESC").
		Limit(20).
		Scan(&stats.ModelDistribution).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ──────────────────────────────────
// 归档相关方法
// ──────────────────────────────────

// CountAll 获取训练数据总记录数
func (r *TrainingDataRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.LLMTrainingData{}).Count(&count).Error
	return count, err
}

// GetArchiveBoundaryID 获取归档边界 ID
// 返回按 ID 升序排列第 n 条记录的 ID，用于确定归档快照的安全边界
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

// GetIDRange 获取指定 ID 范围内的记录 ID 边界（min/max）
func (r *TrainingDataRepository) GetIDRange(maxID int64) (minID, actualMaxID int64, err error) {
	err = r.db.Model(&model.LLMTrainingData{}).
		Select("MIN(id), MAX(id)").
		Where("id <= ?", maxID).
		Row().Scan(&minID, &actualMaxID)
	return
}

// StreamByIDRange 按 ID 范围分批读取完整记录（用于归档导出）
func (r *TrainingDataRepository) StreamByIDRange(maxID int64, batchSize int, fn func(batch []model.LLMTrainingData) error) error {
	var records []model.LLMTrainingData
	return r.db.Where("id <= ?", maxID).
		Order("id ASC").
		FindInBatches(&records, batchSize, func(_ *gorm.DB, _ int) error {
			return fn(records)
		}).Error
}

// DeleteByIDRange 按 ID 范围分批删除已归档的记录
// 每批删除 batchSize 条，避免长事务和锁竞争
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
