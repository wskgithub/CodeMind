package repository

import (
	"time"

	"codemind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UsageRepository Token 用量数据访问层
type UsageRepository struct {
	db *gorm.DB
}

// NewUsageRepository 创建用量 Repository
func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

// CreateUsage 记录单次请求用量
func (r *UsageRepository) CreateUsage(usage *model.TokenUsage) error {
	return r.db.Create(usage).Error
}

// UpsertDaily 更新或插入每日汇总（使用 UPSERT）
func (r *UsageRepository) UpsertDaily(userID int64, date time.Time, promptTokens, completionTokens, totalTokens int, cacheCreationTokens, cacheReadTokens int) error {
	daily := model.TokenUsageDaily{
		UserID:                   userID,
		UsageDate:                date,
		PromptTokens:             int64(promptTokens),
		CompletionTokens:         int64(completionTokens),
		TotalTokens:              int64(totalTokens),
		CacheCreationInputTokens: int64(cacheCreationTokens),
		CacheReadInputTokens:     int64(cacheReadTokens),
		RequestCount:             1,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "usage_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"prompt_tokens":                 gorm.Expr("token_usage_daily.prompt_tokens + ?", promptTokens),
			"completion_tokens":             gorm.Expr("token_usage_daily.completion_tokens + ?", completionTokens),
			"total_tokens":                  gorm.Expr("token_usage_daily.total_tokens + ?", totalTokens),
			"cache_creation_input_tokens":   gorm.Expr("token_usage_daily.cache_creation_input_tokens + ?", cacheCreationTokens),
			"cache_read_input_tokens":       gorm.Expr("token_usage_daily.cache_read_input_tokens + ?", cacheReadTokens),
			"request_count":                 gorm.Expr("token_usage_daily.request_count + 1"),
			"updated_at":                    gorm.Expr("NOW()"),
		}),
	}).Create(&daily).Error
}

// UpsertDailyKey 更新或插入 Key 级每日汇总（使用 UPSERT）
func (r *UsageRepository) UpsertDailyKey(keyID, userID int64, date time.Time, promptTokens, completionTokens, totalTokens int, cacheCreationTokens, cacheReadTokens int) error {
	daily := model.TokenUsageDailyKey{
		APIKeyID:                 keyID,
		UserID:                   userID,
		UsageDate:                date,
		PromptTokens:             int64(promptTokens),
		CompletionTokens:         int64(completionTokens),
		TotalTokens:              int64(totalTokens),
		CacheCreationInputTokens: int64(cacheCreationTokens),
		CacheReadInputTokens:     int64(cacheReadTokens),
		RequestCount:             1,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "api_key_id"}, {Name: "usage_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"prompt_tokens":                 gorm.Expr("token_usage_daily_key.prompt_tokens + ?", promptTokens),
			"completion_tokens":             gorm.Expr("token_usage_daily_key.completion_tokens + ?", completionTokens),
			"total_tokens":                  gorm.Expr("token_usage_daily_key.total_tokens + ?", totalTokens),
			"cache_creation_input_tokens":   gorm.Expr("token_usage_daily_key.cache_creation_input_tokens + ?", cacheCreationTokens),
			"cache_read_input_tokens":       gorm.Expr("token_usage_daily_key.cache_read_input_tokens + ?", cacheReadTokens),
			"request_count":                 gorm.Expr("token_usage_daily_key.request_count + 1"),
			"updated_at":                    gorm.Expr("NOW()"),
		}),
	}).Create(&daily).Error
}

// GetDailyStats 查询每日统计数据
func (r *UsageRepository) GetDailyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	query := r.db.Table("token_usage_daily").
		Select("usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(request_count) as request_count, SUM(cache_creation_input_tokens) as cache_creation_input_tokens, SUM(cache_read_input_tokens) as cache_read_input_tokens").
		Where("usage_date BETWEEN ? AND ?", startDate, endDate)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if deptID != nil {
		query = query.Where("user_id IN (SELECT id FROM users WHERE department_id = ? AND deleted_at IS NULL)", *deptID)
	}

	err := query.Group("usage_date").Order("usage_date ASC").Scan(&rows).Error
	return rows, err
}

// GetWeeklyStats 查询每周统计
func (r *UsageRepository) GetWeeklyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	query := r.db.Table("token_usage_daily").
		Select("DATE_TRUNC('week', usage_date)::date as usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(request_count) as request_count, SUM(cache_creation_input_tokens) as cache_creation_input_tokens, SUM(cache_read_input_tokens) as cache_read_input_tokens").
		Where("usage_date BETWEEN ? AND ?", startDate, endDate)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if deptID != nil {
		query = query.Where("user_id IN (SELECT id FROM users WHERE department_id = ? AND deleted_at IS NULL)", *deptID)
	}

	err := query.Group("DATE_TRUNC('week', usage_date)").Order("usage_date ASC").Scan(&rows).Error
	return rows, err
}

// GetMonthlyStats 查询每月统计
func (r *UsageRepository) GetMonthlyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	query := r.db.Table("token_usage_daily").
		Select("DATE_TRUNC('month', usage_date)::date as usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(request_count) as request_count, SUM(cache_creation_input_tokens) as cache_creation_input_tokens, SUM(cache_read_input_tokens) as cache_read_input_tokens").
		Where("usage_date BETWEEN ? AND ?", startDate, endDate)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if deptID != nil {
		query = query.Where("user_id IN (SELECT id FROM users WHERE department_id = ? AND deleted_at IS NULL)", *deptID)
	}

	err := query.Group("DATE_TRUNC('month', usage_date)").Order("usage_date ASC").Scan(&rows).Error
	return rows, err
}

// GetTodayTotalTokens 获取今日总用量（Asia/Shanghai 时区）
func (r *UsageRepository) GetTodayTotalTokens(userID *int64) (int64, error) {
	var total int64
	query := r.db.Table("token_usage_daily").
		Select("COALESCE(SUM(total_tokens), 0)").
		Where("usage_date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')::date")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Scan(&total).Error
	return total, err
}

// GetTodayRequestCount 获取今日请求总数（Asia/Shanghai 时区）
func (r *UsageRepository) GetTodayRequestCount(userID *int64) (int64, error) {
	var count int64
	query := r.db.Table("token_usage_daily").
		Select("COALESCE(SUM(request_count), 0)").
		Where("usage_date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')::date")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Scan(&count).Error
	return count, err
}

// GetTodayActiveUsers 获取今日活跃用户数（Asia/Shanghai 时区）
func (r *UsageRepository) GetTodayActiveUsers() (int64, error) {
	var count int64
	err := r.db.Table("token_usage_daily").
		Where("usage_date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')::date").
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// GetMonthTotalTokens 获取本月总用量（Asia/Shanghai 时区）
func (r *UsageRepository) GetMonthTotalTokens(userID *int64) (int64, error) {
	var total int64
	query := r.db.Table("token_usage_daily").
		Select("COALESCE(SUM(total_tokens), 0)").
		Where("usage_date >= DATE_TRUNC('month', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Scan(&total).Error
	return total, err
}

// GetMonthRequestCount 获取本月请求总数（Asia/Shanghai 时区）
func (r *UsageRepository) GetMonthRequestCount(userID *int64) (int64, error) {
	var count int64
	query := r.db.Table("token_usage_daily").
		Select("COALESCE(SUM(request_count), 0)").
		Where("usage_date >= DATE_TRUNC('month', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Scan(&count).Error
	return count, err
}

// GetMonthActiveUsers 获取本月活跃用户数（Asia/Shanghai 时区）
func (r *UsageRepository) GetMonthActiveUsers() (int64, error) {
	var count int64
	err := r.db.Table("token_usage_daily").
		Where("usage_date >= DATE_TRUNC('month', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')").
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// GetUserRanking 获取用户用量排行
func (r *UsageRepository) GetUserRanking(deptID *int64, startDate, endDate time.Time, limit int) ([]RankingRow, error) {
	var rows []RankingRow

	query := r.db.Table("token_usage_daily d").
		Select("d.user_id as id, u.display_name as name, SUM(d.total_tokens) as total_tokens, SUM(d.request_count) as request_count").
		Joins("JOIN users u ON u.id = d.user_id AND u.deleted_at IS NULL").
		Where("d.usage_date BETWEEN ? AND ?", startDate, endDate)

	if deptID != nil {
		query = query.Where("u.department_id = ?", *deptID)
	}

	err := query.Group("d.user_id, u.display_name").
		Order("total_tokens DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetDeptRanking 获取部门用量排行
func (r *UsageRepository) GetDeptRanking(startDate, endDate time.Time, limit int) ([]RankingRow, error) {
	var rows []RankingRow

	err := r.db.Table("token_usage_daily d").
		Select("u.department_id as id, dep.name as name, SUM(d.total_tokens) as total_tokens, SUM(d.request_count) as request_count").
		Joins("JOIN users u ON u.id = d.user_id AND u.deleted_at IS NULL").
		Joins("JOIN departments dep ON dep.id = u.department_id").
		Where("d.usage_date BETWEEN ? AND ?", startDate, endDate).
		Where("u.department_id IS NOT NULL").
		Group("u.department_id, dep.name").
		Order("total_tokens DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// GetPeriodUsage 获取指定周期的用户用量总计
// 使用参数化查询防止 SQL 注入
func (r *UsageRepository) GetPeriodUsage(userID int64, period string, periodKey string) (int64, error) {
	var total int64

	query := r.db.Table("token_usage_daily").
		Select("COALESCE(SUM(total_tokens), 0)").
		Where("user_id = ?", userID)

	switch period {
	case "monthly":
		query = query.Where("DATE_TRUNC('month', usage_date) = DATE_TRUNC('month', ?::date)", periodKey)
	default:
		query = query.Where("usage_date = ?::date", periodKey)
	}

	err := query.Scan(&total).Error
	return total, err
}

// GetKeyUsageStats 获取指定 Key 的用量统计
// 直接查询 token_usage_daily_key 汇总表，避免扫描原始明细表
func (r *UsageRepository) GetKeyUsageStats(keyID int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	err := r.db.Table("token_usage_daily_key").
		Select("usage_date, prompt_tokens, completion_tokens, total_tokens, request_count").
		Where("api_key_id = ? AND usage_date BETWEEN ? AND ?", keyID, startDate, endDate).
		Order("usage_date ASC").
		Scan(&rows).Error
	return rows, err
}

// GetDetailedUsageStats 获取详细用量数据（用于导出）
// 返回每个用户每天的用量明细
func (r *UsageRepository) GetDetailedUsageStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]UsageExportRow, error) {
	var rows []UsageExportRow

	query := r.db.Table("token_usage_daily d").
		Select("d.usage_date, u.username as user_name, COALESCE(dep.name, '-') as dept_name, d.prompt_tokens, d.completion_tokens, d.total_tokens, d.request_count").
		Joins("JOIN users u ON u.id = d.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN departments dep ON dep.id = u.department_id").
		Where("d.usage_date BETWEEN ? AND ?", startDate, endDate)

	if userID != nil {
		query = query.Where("d.user_id = ?", *userID)
	}
	if deptID != nil {
		query = query.Where("u.department_id = ?", *deptID)
	}

	err := query.Order("d.usage_date DESC, u.username ASC").Scan(&rows).Error
	return rows, err
}

// UsageExportRow 导出数据查询结果行
type UsageExportRow struct {
	UsageDate        time.Time `gorm:"column:usage_date"`
	UserName         string    `gorm:"column:user_name"`
	DeptName         string    `gorm:"column:dept_name"`
	PromptTokens     int64     `gorm:"column:prompt_tokens"`
	CompletionTokens int64     `gorm:"column:completion_tokens"`
	TotalTokens      int64     `gorm:"column:total_tokens"`
	RequestCount     int64     `gorm:"column:request_count"`
}

// DailyStatRow 统计查询结果行
type DailyStatRow struct {
	UsageDate                time.Time `gorm:"column:usage_date"`
	PromptTokens             int64     `gorm:"column:prompt_tokens"`
	CompletionTokens         int64     `gorm:"column:completion_tokens"`
	TotalTokens              int64     `gorm:"column:total_tokens"`
	RequestCount             int64     `gorm:"column:request_count"`
	CacheCreationInputTokens int64     `gorm:"column:cache_creation_input_tokens"` // 缓存创建 Token 数
	CacheReadInputTokens     int64     `gorm:"column:cache_read_input_tokens"`     // 缓存命中 Token 数
}

// RankingRow 排行榜查询结果行
type RankingRow struct {
	ID           int64  `gorm:"column:id"`
	Name         string `gorm:"column:name"`
	TotalTokens  int64  `gorm:"column:total_tokens"`
	RequestCount int64  `gorm:"column:request_count"`
}

// GetKeyUsageSummary 获取每个 Key 的平台用量汇总（基于 Key 级每日汇总表）
func (r *UsageRepository) GetKeyUsageSummary(userID *int64, deptID *int64, startDate, endDate time.Time) ([]KeyUsageRow, error) {
	var rows []KeyUsageRow

	query := r.db.Table("token_usage_daily_key d").
		Select("d.api_key_id as id, k.name, SUM(d.prompt_tokens) as prompt_tokens, SUM(d.completion_tokens) as completion_tokens, SUM(d.total_tokens) as total_tokens, SUM(d.request_count) as request_count").
		Joins("JOIN api_keys k ON k.id = d.api_key_id").
		Where("d.usage_date BETWEEN ? AND ?", startDate, endDate)

	if userID != nil {
		query = query.Where("d.user_id = ?", *userID)
	}
	if deptID != nil {
		query = query.Where("d.user_id IN (SELECT id FROM users WHERE department_id = ? AND deleted_at IS NULL)", *deptID)
	}

	err := query.Group("d.api_key_id, k.name").
		Order("total_tokens DESC").
		Scan(&rows).Error
	return rows, err
}

// GetThirdPartyKeyUsageSummary 获取每个 Key 的第三方用量汇总
func (r *UsageRepository) GetThirdPartyKeyUsageSummary(userID *int64, deptID *int64, startDate, endDate time.Time) ([]KeyUsageRow, error) {
	var rows []KeyUsageRow

	query := r.db.Table("third_party_token_usage t").
		Select("t.api_key_id as id, k.name, SUM(t.prompt_tokens) as prompt_tokens, SUM(t.completion_tokens) as completion_tokens, SUM(t.total_tokens) as total_tokens, COUNT(*) as request_count").
		Joins("JOIN api_keys k ON k.id = t.api_key_id").
		Where("(t.created_at AT TIME ZONE 'Asia/Shanghai')::date BETWEEN ? AND ?", startDate, endDate)

	if userID != nil {
		query = query.Where("t.user_id = ?", *userID)
	}
	if deptID != nil {
		query = query.Where("t.user_id IN (SELECT id FROM users WHERE department_id = ? AND deleted_at IS NULL)", *deptID)
	}

	err := query.Group("t.api_key_id, k.name").
		Order("total_tokens DESC").
		Scan(&rows).Error
	return rows, err
}

// KeyUsageRow Key 用量查询结果行
type KeyUsageRow struct {
	ID               int64  `gorm:"column:id"`
	Name             string `gorm:"column:name"`
	PromptTokens     int64  `gorm:"column:prompt_tokens"`
	CompletionTokens int64  `gorm:"column:completion_tokens"`
	TotalTokens      int64  `gorm:"column:total_tokens"`
	RequestCount     int64  `gorm:"column:request_count"`
}

// ──────────────────────────────────
// 第三方模型服务用量查询
// ──────────────────────────────────

// GetThirdPartyTodayTotalTokens 获取今日第三方服务总 token 用量（Asia/Shanghai 时区）
func (r *UsageRepository) GetThirdPartyTodayTotalTokens(userID *int64) (int64, error) {
	var total int64
	query := r.db.Table("third_party_token_usage").
		Select("COALESCE(SUM(total_tokens), 0)").
		Where("created_at >= DATE_TRUNC('day', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai') AT TIME ZONE 'Asia/Shanghai'")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Scan(&total).Error
	return total, err
}

// GetThirdPartyTodayRequestCount 获取今日第三方服务请求数（Asia/Shanghai 时区）
func (r *UsageRepository) GetThirdPartyTodayRequestCount(userID *int64) (int64, error) {
	var count int64
	query := r.db.Table("third_party_token_usage").
		Where("created_at >= DATE_TRUNC('day', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai') AT TIME ZONE 'Asia/Shanghai'")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Count(&count).Error
	return count, err
}

// GetThirdPartyMonthTotalTokens 获取本月第三方服务总 token 用量（Asia/Shanghai 时区）
func (r *UsageRepository) GetThirdPartyMonthTotalTokens(userID *int64) (int64, error) {
	var total int64
	query := r.db.Table("third_party_token_usage").
		Select("COALESCE(SUM(total_tokens), 0)").
		Where("created_at >= DATE_TRUNC('month', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai') AT TIME ZONE 'Asia/Shanghai'")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Scan(&total).Error
	return total, err
}

// GetThirdPartyMonthRequestCount 获取本月第三方服务请求数（Asia/Shanghai 时区）
func (r *UsageRepository) GetThirdPartyMonthRequestCount(userID *int64) (int64, error) {
	var count int64
	query := r.db.Table("third_party_token_usage").
		Where("created_at >= DATE_TRUNC('month', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai') AT TIME ZONE 'Asia/Shanghai'")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Count(&count).Error
	return count, err
}

// GetThirdPartyDailyStats 查询第三方服务每日统计数据
func (r *UsageRepository) GetThirdPartyDailyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	return r.getThirdPartyStats("", userID, deptID, startDate, endDate)
}

// GetThirdPartyWeeklyStats 查询第三方服务每周统计
func (r *UsageRepository) GetThirdPartyWeeklyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	return r.getThirdPartyStats("week", userID, deptID, startDate, endDate)
}

// GetThirdPartyMonthlyStats 查询第三方服务每月统计
func (r *UsageRepository) GetThirdPartyMonthlyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	return r.getThirdPartyStats("month", userID, deptID, startDate, endDate)
}

// getThirdPartyStats 通用第三方用量聚合查询
// trunc 为空字符串表示按日聚合，否则按 DATE_TRUNC(trunc, ...) 聚合
func (r *UsageRepository) getThirdPartyStats(trunc string, userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	dateExpr := "(created_at AT TIME ZONE 'Asia/Shanghai')::date"
	if trunc != "" {
		dateExpr = "DATE_TRUNC('" + trunc + "', created_at AT TIME ZONE 'Asia/Shanghai')::date"
	}

	var rows []DailyStatRow
	query := r.db.Table("third_party_token_usage").
		Select(dateExpr+" as usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, COUNT(*) as request_count, SUM(cache_creation_input_tokens) as cache_creation_input_tokens, SUM(cache_read_input_tokens) as cache_read_input_tokens").
		Where("(created_at AT TIME ZONE 'Asia/Shanghai')::date BETWEEN ? AND ?", startDate, endDate)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if deptID != nil {
		query = query.Where("user_id IN (SELECT id FROM users WHERE department_id = ? AND deleted_at IS NULL)", *deptID)
	}

	err := query.Group(dateExpr).Order("usage_date ASC").Scan(&rows).Error
	return rows, err
}

// CreateRequestLog 记录请求日志
func (r *UsageRepository) CreateRequestLog(log *model.RequestLog) error {
	return r.db.Create(log).Error
}

// ──────────────────────────────────
// 数据保留清理
// ──────────────────────────────────

// DeleteOldUsageRecords 分批删除超过保留期的 token_usage 明细记录
// token_usage 数据已聚合到 token_usage_daily，明细仅用于排查
func (r *UsageRepository) DeleteOldUsageRecords(before time.Time, batchSize int) (int64, error) {
	return r.batchDeleteByCreatedAt("token_usage", before, batchSize)
}

// DeleteOldRequestLogs 分批删除超过保留期的请求日志
func (r *UsageRepository) DeleteOldRequestLogs(before time.Time, batchSize int) (int64, error) {
	return r.batchDeleteByCreatedAt("request_logs", before, batchSize)
}

// batchDeleteByCreatedAt 通用分批删除（按 created_at 字段）
func (r *UsageRepository) batchDeleteByCreatedAt(table string, before time.Time, batchSize int) (int64, error) {
	var totalDeleted int64
	for {
		result := r.db.Exec(
			"DELETE FROM "+table+" WHERE id IN (SELECT id FROM "+table+" WHERE created_at < ? LIMIT ?)",
			before, batchSize,
		)
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
