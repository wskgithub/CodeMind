package repository

import (
	"time"

	"codemind/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UsageRepository handles token usage data access.
type UsageRepository struct {
	db *gorm.DB
}

// NewUsageRepository creates a new UsageRepository.
func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

// CreateUsage records a single request's usage.
func (r *UsageRepository) CreateUsage(usage *model.TokenUsage) error {
	return r.db.Create(usage).Error
}

// UpsertDaily upserts daily usage summary.
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
			"prompt_tokens":               gorm.Expr("token_usage_daily.prompt_tokens + ?", promptTokens),
			"completion_tokens":           gorm.Expr("token_usage_daily.completion_tokens + ?", completionTokens),
			"total_tokens":                gorm.Expr("token_usage_daily.total_tokens + ?", totalTokens),
			"cache_creation_input_tokens": gorm.Expr("token_usage_daily.cache_creation_input_tokens + ?", cacheCreationTokens),
			"cache_read_input_tokens":     gorm.Expr("token_usage_daily.cache_read_input_tokens + ?", cacheReadTokens),
			"request_count":               gorm.Expr("token_usage_daily.request_count + 1"),
			"updated_at":                  gorm.Expr("NOW()"),
		}),
	}).Create(&daily).Error
}

// UpsertDailyKey upserts key-level daily usage summary.
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
			"prompt_tokens":               gorm.Expr("token_usage_daily_key.prompt_tokens + ?", promptTokens),
			"completion_tokens":           gorm.Expr("token_usage_daily_key.completion_tokens + ?", completionTokens),
			"total_tokens":                gorm.Expr("token_usage_daily_key.total_tokens + ?", totalTokens),
			"cache_creation_input_tokens": gorm.Expr("token_usage_daily_key.cache_creation_input_tokens + ?", cacheCreationTokens),
			"cache_read_input_tokens":     gorm.Expr("token_usage_daily_key.cache_read_input_tokens + ?", cacheReadTokens),
			"request_count":               gorm.Expr("token_usage_daily_key.request_count + 1"),
			"updated_at":                  gorm.Expr("NOW()"),
		}),
	}).Create(&daily).Error
}

// GetDailyStats retrieves daily statistics.
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

// GetWeeklyStats retrieves weekly statistics.
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

// GetMonthlyStats retrieves monthly statistics.
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

// GetTodayTotalTokens returns today's total token usage (Asia/Shanghai timezone).
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

// GetTodayRequestCount returns today's total request count (Asia/Shanghai timezone).
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

// GetTodayActiveUsers returns today's active user count (Asia/Shanghai timezone).
func (r *UsageRepository) GetTodayActiveUsers() (int64, error) {
	var count int64
	err := r.db.Table("token_usage_daily").
		Where("usage_date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')::date").
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// GetMonthTotalTokens returns this month's total token usage (Asia/Shanghai timezone).
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

// GetMonthRequestCount returns this month's total request count (Asia/Shanghai timezone).
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

// GetMonthActiveUsers returns this month's active user count (Asia/Shanghai timezone).
func (r *UsageRepository) GetMonthActiveUsers() (int64, error) {
	var count int64
	err := r.db.Table("token_usage_daily").
		Where("usage_date >= DATE_TRUNC('month', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai')").
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// GetUserRanking returns user usage rankings.
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

// GetDeptRanking returns department usage rankings.
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

// GetPeriodUsage returns total usage for a user within a specified period.
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

// GetKeyUsageStats returns usage statistics for a specific API key.
func (r *UsageRepository) GetKeyUsageStats(keyID int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	err := r.db.Table("token_usage_daily_key").
		Select("usage_date, prompt_tokens, completion_tokens, total_tokens, request_count").
		Where("api_key_id = ? AND usage_date BETWEEN ? AND ?", keyID, startDate, endDate).
		Order("usage_date ASC").
		Scan(&rows).Error
	return rows, err
}

// GetDetailedUsageStats returns detailed usage data for export.
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

// UsageExportRow represents an export data row.
type UsageExportRow struct {
	UsageDate        time.Time `gorm:"column:usage_date"`
	UserName         string    `gorm:"column:user_name"`
	DeptName         string    `gorm:"column:dept_name"`
	PromptTokens     int64     `gorm:"column:prompt_tokens"`
	CompletionTokens int64     `gorm:"column:completion_tokens"`
	TotalTokens      int64     `gorm:"column:total_tokens"`
	RequestCount     int64     `gorm:"column:request_count"`
}

// DailyStatRow represents a statistics query result row.
type DailyStatRow struct {
	UsageDate                time.Time `gorm:"column:usage_date"`
	PromptTokens             int64     `gorm:"column:prompt_tokens"`
	CompletionTokens         int64     `gorm:"column:completion_tokens"`
	TotalTokens              int64     `gorm:"column:total_tokens"`
	RequestCount             int64     `gorm:"column:request_count"`
	CacheCreationInputTokens int64     `gorm:"column:cache_creation_input_tokens"`
	CacheReadInputTokens     int64     `gorm:"column:cache_read_input_tokens"`
}

// RankingRow represents a ranking query result row.
type RankingRow struct {
	Name         string `gorm:"column:name"`
	ID           int64  `gorm:"column:id"`
	TotalTokens  int64  `gorm:"column:total_tokens"`
	RequestCount int64  `gorm:"column:request_count"`
}

// GetKeyUsageSummary returns platform usage summary for each API key.
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

// GetThirdPartyKeyUsageSummary returns third-party usage summary for each API key.
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

// KeyUsageRow represents a key usage query result row.
type KeyUsageRow struct {
	Name             string `gorm:"column:name"`
	ID               int64  `gorm:"column:id"`
	PromptTokens     int64  `gorm:"column:prompt_tokens"`
	CompletionTokens int64  `gorm:"column:completion_tokens"`
	TotalTokens      int64  `gorm:"column:total_tokens"`
	RequestCount     int64  `gorm:"column:request_count"`
}

// GetThirdPartyTodayTotalTokens returns today's third-party service total tokens (Asia/Shanghai timezone).
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

// GetThirdPartyTodayRequestCount returns today's third-party service request count (Asia/Shanghai timezone).
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

// GetThirdPartyMonthTotalTokens returns this month's third-party service total tokens (Asia/Shanghai timezone).
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

// GetThirdPartyMonthRequestCount returns this month's third-party service request count (Asia/Shanghai timezone).
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

// GetThirdPartyDailyStats retrieves daily statistics for third-party services.
func (r *UsageRepository) GetThirdPartyDailyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	return r.getThirdPartyStats("", userID, deptID, startDate, endDate)
}

// GetThirdPartyWeeklyStats retrieves weekly statistics for third-party services.
func (r *UsageRepository) GetThirdPartyWeeklyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	return r.getThirdPartyStats("week", userID, deptID, startDate, endDate)
}

// GetThirdPartyMonthlyStats retrieves monthly statistics for third-party services.
func (r *UsageRepository) GetThirdPartyMonthlyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	return r.getThirdPartyStats("month", userID, deptID, startDate, endDate)
}

// getThirdPartyStats performs aggregated third-party usage query.
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

// CreateRequestLog records a request log.
func (r *UsageRepository) CreateRequestLog(log *model.RequestLog) error {
	return r.db.Create(log).Error
}

// DeleteOldUsageRecords batch deletes token_usage records older than the retention period.
func (r *UsageRepository) DeleteOldUsageRecords(before time.Time, batchSize int) (int64, error) {
	return r.batchDeleteByCreatedAt("token_usage", before, batchSize)
}

// DeleteOldRequestLogs batch deletes request logs older than the retention period.
func (r *UsageRepository) DeleteOldRequestLogs(before time.Time, batchSize int) (int64, error) {
	return r.batchDeleteByCreatedAt("request_logs", before, batchSize)
}

// batchDeleteByCreatedAt performs batch deletion by created_at field.
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
