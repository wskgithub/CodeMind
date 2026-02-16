package service

import (
	"time"

	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// StatsService 用量统计业务逻辑
type StatsService struct {
	usageRepo *repository.UsageRepository
	userRepo  *repository.UserRepository
	deptRepo  *repository.DepartmentRepository
	keyRepo   *repository.APIKeyRepository
	logger    *zap.Logger
}

// NewStatsService 创建统计服务
func NewStatsService(
	usageRepo *repository.UsageRepository,
	userRepo *repository.UserRepository,
	deptRepo *repository.DepartmentRepository,
	keyRepo *repository.APIKeyRepository,
	logger *zap.Logger,
) *StatsService {
	return &StatsService{
		usageRepo: usageRepo,
		userRepo:  userRepo,
		deptRepo:  deptRepo,
		keyRepo:   keyRepo,
		logger:    logger,
	}
}

// GetOverview 获取用量总览
func (s *StatsService) GetOverview(userID *int64, role string, deptID *int64) (*dto.StatsOverview, error) {
	overview := &dto.StatsOverview{SystemStatus: "healthy"}

	// 根据角色决定统计范围
	var filterUserID *int64
	if role == "user" {
		filterUserID = userID
	}

	// 今日统计
	todayTokens, _ := s.usageRepo.GetTodayTotalTokens(filterUserID)
	todayRequests, _ := s.usageRepo.GetTodayRequestCount(filterUserID)
	todayActive, _ := s.usageRepo.GetTodayActiveUsers()

	overview.Today = dto.PeriodStats{
		TotalTokens:   todayTokens,
		TotalRequests: todayRequests,
		ActiveUsers:   todayActive,
	}

	// 本月统计
	monthTokens, _ := s.usageRepo.GetMonthTotalTokens(filterUserID)
	monthRequests, _ := s.usageRepo.GetMonthRequestCount(filterUserID)
	monthActive, _ := s.usageRepo.GetMonthActiveUsers()

	overview.ThisMonth = dto.PeriodStats{
		TotalTokens:   monthTokens,
		TotalRequests: monthRequests,
		ActiveUsers:   monthActive,
	}

	// 系统概况（管理员可见）
	if role != "user" {
		totalUsers, _ := s.userRepo.CountAll()
		totalDepts, _ := s.deptRepo.CountAll()
		totalKeys, _ := s.keyRepo.CountAll()
		overview.TotalUsers = totalUsers
		overview.TotalDepts = totalDepts
		overview.TotalAPIKeys = totalKeys
	}

	return overview, nil
}

// GetUsageStats 获取用量统计数据
func (s *StatsService) GetUsageStats(query *dto.StatsQuery, operatorRole string, operatorID int64, operatorDeptID *int64) (*dto.UsageResponse, error) {
	// 确定查询范围
	var userID *int64
	var deptID *int64

	switch operatorRole {
	case "super_admin":
		userID = query.UserID
		deptID = query.DepartmentID
	case "dept_manager":
		if query.UserID != nil {
			userID = query.UserID
		}
		deptID = operatorDeptID
	default:
		userID = &operatorID
	}

	// 解析日期范围
	startDate, endDate := s.parseDateRange(query.StartDate, query.EndDate, query.Period)

	// 根据周期查询
	var rows []repository.DailyStatRow
	var err error

	switch query.Period {
	case "daily":
		rows, err = s.usageRepo.GetDailyStats(userID, deptID, startDate, endDate)
	case "weekly":
		rows, err = s.usageRepo.GetWeeklyStats(userID, deptID, startDate, endDate)
	case "monthly":
		rows, err = s.usageRepo.GetMonthlyStats(userID, deptID, startDate, endDate)
	default:
		return nil, errcode.ErrInvalidParams.WithMessage("无效的统计周期")
	}

	if err != nil {
		s.logger.Error("查询统计数据失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	// 转换为响应格式
	items := make([]dto.UsageItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.UsageItem{
			Date:             row.UsageDate.Format("2006-01-02"),
			PromptTokens:     row.PromptTokens,
			CompletionTokens: row.CompletionTokens,
			TotalTokens:      row.TotalTokens,
			RequestCount:     row.RequestCount,
		})
	}

	return &dto.UsageResponse{
		Period: query.Period,
		Items:  items,
	}, nil
}

// GetRanking 获取用量排行榜
func (s *StatsService) GetRanking(query *dto.RankingQuery, operatorDeptID *int64) ([]dto.RankingItem, error) {
	now := time.Now()
	startDate, endDate := s.getPeriodRange(now, query.Period)
	limit := query.GetLimit()

	var rows []repository.RankingRow
	var err error

	switch query.Type {
	case "user":
		rows, err = s.usageRepo.GetUserRanking(operatorDeptID, startDate, endDate, limit)
	case "department":
		rows, err = s.usageRepo.GetDeptRanking(startDate, endDate, limit)
	default:
		return nil, errcode.ErrInvalidParams.WithMessage("无效的排行类型")
	}

	if err != nil {
		return nil, errcode.ErrDatabase
	}

	items := make([]dto.RankingItem, 0, len(rows))
	for i, row := range rows {
		items = append(items, dto.RankingItem{
			Rank:         i + 1,
			ID:           row.ID,
			Name:         row.Name,
			TotalTokens:  row.TotalTokens,
			RequestCount: row.RequestCount,
		})
	}

	return items, nil
}

// parseDateRange 解析日期范围，无参数时使用默认值
func (s *StatsService) parseDateRange(startStr, endStr, period string) (time.Time, time.Time) {
	now := time.Now()
	var startDate, endDate time.Time

	if endStr != "" {
		endDate, _ = time.Parse("2006-01-02", endStr)
	}
	if endDate.IsZero() {
		endDate = now
	}

	if startStr != "" {
		startDate, _ = time.Parse("2006-01-02", startStr)
	}
	if startDate.IsZero() {
		switch period {
		case "daily":
			startDate = now.AddDate(0, 0, -30)
		case "weekly":
			startDate = now.AddDate(0, -3, 0)
		case "monthly":
			startDate = now.AddDate(-1, 0, 0)
		default:
			startDate = now.AddDate(0, 0, -30)
		}
	}

	return startDate, endDate
}

// getPeriodRange 根据当前时间和周期获取日期范围
func (s *StatsService) getPeriodRange(now time.Time, period string) (time.Time, time.Time) {
	switch period {
	case "daily":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	case "weekly":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -(weekday - 1))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, now
	default:
		return now.AddDate(0, 0, -30), now
	}
}
