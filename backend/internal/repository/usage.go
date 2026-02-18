package repository

import (
	"fmt"
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
func (r *UsageRepository) UpsertDaily(userID int64, date time.Time, promptTokens, completionTokens, totalTokens int) error {
	daily := model.TokenUsageDaily{
		UserID:           userID,
		UsageDate:        date,
		PromptTokens:     int64(promptTokens),
		CompletionTokens: int64(completionTokens),
		TotalTokens:      int64(totalTokens),
		RequestCount:     1,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "usage_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"prompt_tokens":     gorm.Expr("token_usage_daily.prompt_tokens + ?", promptTokens),
			"completion_tokens": gorm.Expr("token_usage_daily.completion_tokens + ?", completionTokens),
			"total_tokens":      gorm.Expr("token_usage_daily.total_tokens + ?", totalTokens),
			"request_count":     gorm.Expr("token_usage_daily.request_count + 1"),
			"updated_at":        gorm.Expr("NOW()"),
		}),
	}).Create(&daily).Error
}

// GetDailyStats 查询每日统计数据
func (r *UsageRepository) GetDailyStats(userID *int64, deptID *int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	query := r.db.Table("token_usage_daily").
		Select("usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(request_count) as request_count").
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
		Select("DATE_TRUNC('week', usage_date)::date as usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(request_count) as request_count").
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
		Select("DATE_TRUNC('month', usage_date)::date as usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, SUM(request_count) as request_count").
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
func (r *UsageRepository) GetPeriodUsage(userID int64, period string, periodKey string) (int64, error) {
	var total int64

	var dateFilter string
	switch period {
	case "daily":
		dateFilter = fmt.Sprintf("usage_date = '%s'", periodKey)
	case "monthly":
		dateFilter = fmt.Sprintf("DATE_TRUNC('month', usage_date) = DATE_TRUNC('month', '%s'::date)", periodKey)
	default:
		dateFilter = fmt.Sprintf("usage_date = '%s'", periodKey)
	}

	err := r.db.Table("token_usage_daily").
		Select("COALESCE(SUM(total_tokens), 0)").
		Where("user_id = ?", userID).
		Where(dateFilter).
		Scan(&total).Error
	return total, err
}

// GetKeyUsageStats 获取指定 Key 的用量统计
// 使用 Asia/Shanghai 时区提取日期，确保与每日汇总表一致
func (r *UsageRepository) GetKeyUsageStats(keyID int64, startDate, endDate time.Time) ([]DailyStatRow, error) {
	var rows []DailyStatRow

	err := r.db.Table("token_usage").
		Select("(created_at AT TIME ZONE 'Asia/Shanghai')::date as usage_date, SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, COUNT(*) as request_count").
		Where("api_key_id = ? AND created_at BETWEEN ? AND ?", keyID, startDate, endDate).
		Group("(created_at AT TIME ZONE 'Asia/Shanghai')::date").
		Order("usage_date ASC").
		Scan(&rows).Error
	return rows, err
}

// DailyStatRow 统计查询结果行
type DailyStatRow struct {
	UsageDate        time.Time `gorm:"column:usage_date"`
	PromptTokens     int64     `gorm:"column:prompt_tokens"`
	CompletionTokens int64     `gorm:"column:completion_tokens"`
	TotalTokens      int64     `gorm:"column:total_tokens"`
	RequestCount     int64     `gorm:"column:request_count"`
}

// RankingRow 排行榜查询结果行
type RankingRow struct {
	ID           int64  `gorm:"column:id"`
	Name         string `gorm:"column:name"`
	TotalTokens  int64  `gorm:"column:total_tokens"`
	RequestCount int64  `gorm:"column:request_count"`
}

// CreateRequestLog 记录请求日志
func (r *UsageRepository) CreateRequestLog(log *model.RequestLog) error {
	return r.db.Create(log).Error
}
