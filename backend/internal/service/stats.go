package service

import (
	"sort"
	"sync"
	"time"

	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/timezone"
	"codemind/internal/repository"

	"go.uber.org/zap"
)

// StatsService handles usage statistics.
type StatsService struct {
	usageRepo *repository.UsageRepository
	userRepo  *repository.UserRepository
	deptRepo  *repository.DepartmentRepository
	keyRepo   *repository.APIKeyRepository
	logger    *zap.Logger
}

// NewStatsService creates a new stats service.
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

// GetOverview returns usage overview with parallel queries.
func (s *StatsService) GetOverview(userID *int64, role string, _ *int64) (*dto.StatsOverview, error) {
	overview := &dto.StatsOverview{SystemStatus: "healthy"}

	var filterUserID *int64
	if role == model.RoleUser {
		filterUserID = userID
	}

	var (
		todayTokens, todayRequests, todayActive int64
		monthTokens, monthRequests, monthActive int64
		tpTodayTokens, tpTodayRequests          int64
		tpMonthTokens, tpMonthRequests          int64
		totalUsers, totalDepts, totalKeys       int64
	)

	var wg sync.WaitGroup
	wg.Add(10) //nolint:mnd // intentional constant.

	go func() { defer wg.Done(); todayTokens, _ = s.usageRepo.GetTodayTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); todayRequests, _ = s.usageRepo.GetTodayRequestCount(filterUserID) }()
	go func() { defer wg.Done(); todayActive, _ = s.usageRepo.GetTodayActiveUsers() }()
	go func() { defer wg.Done(); monthTokens, _ = s.usageRepo.GetMonthTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); monthRequests, _ = s.usageRepo.GetMonthRequestCount(filterUserID) }()
	go func() { defer wg.Done(); monthActive, _ = s.usageRepo.GetMonthActiveUsers() }()

	go func() { defer wg.Done(); tpTodayTokens, _ = s.usageRepo.GetThirdPartyTodayTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); tpTodayRequests, _ = s.usageRepo.GetThirdPartyTodayRequestCount(filterUserID) }()
	go func() { defer wg.Done(); tpMonthTokens, _ = s.usageRepo.GetThirdPartyMonthTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); tpMonthRequests, _ = s.usageRepo.GetThirdPartyMonthRequestCount(filterUserID) }()

	if role != model.RoleUser {
		wg.Add(3) //nolint:mnd // intentional constant.
		go func() { defer wg.Done(); totalUsers, _ = s.userRepo.CountAll() }()
		go func() { defer wg.Done(); totalDepts, _ = s.deptRepo.CountAll() }()
		go func() { defer wg.Done(); totalKeys, _ = s.keyRepo.CountAll() }()
	}

	wg.Wait()

	overview.Today = dto.PeriodStats{
		TotalTokens:             todayTokens,
		TotalRequests:           todayRequests,
		ActiveUsers:             todayActive,
		ThirdPartyTotalTokens:   tpTodayTokens,
		ThirdPartyTotalRequests: tpTodayRequests,
	}
	overview.ThisMonth = dto.PeriodStats{
		TotalTokens:             monthTokens,
		TotalRequests:           monthRequests,
		ActiveUsers:             monthActive,
		ThirdPartyTotalTokens:   tpMonthTokens,
		ThirdPartyTotalRequests: tpMonthRequests,
	}
	overview.TotalUsers = totalUsers
	overview.TotalDepts = totalDepts
	overview.TotalAPIKeys = totalKeys

	return overview, nil
}

// GetUsageStats returns usage statistics with merged platform and third-party data.
func (s *StatsService) GetUsageStats(query *dto.StatsQuery, operatorRole string, operatorID int64, operatorDeptID *int64) (*dto.UsageResponse, error) {
	var userID *int64
	var deptID *int64

	switch operatorRole {
	case model.RoleSuperAdmin:
		userID = query.UserID
		deptID = query.DepartmentID
	case model.RoleDeptManager:
		if query.UserID != nil {
			userID = query.UserID
		}
		deptID = operatorDeptID
	default:
		userID = &operatorID
	}

	if query.Period != model.PeriodDaily && query.Period != model.PeriodWeekly && query.Period != model.PeriodMonthly {
		return nil, errcode.ErrInvalidParams.WithMessage("invalid period")
	}

	startDate, endDate := s.parseDateRange(query.StartDate, query.EndDate, query.Period)

	var platformRows, tpRows []repository.DailyStatRow
	var platformErr, tpErr error

	var wg sync.WaitGroup
	wg.Add(2) //nolint:mnd // intentional constant.

	go func() {
		defer wg.Done()
		switch query.Period {
		case model.PeriodDaily:
			platformRows, platformErr = s.usageRepo.GetDailyStats(userID, deptID, startDate, endDate)
		case model.PeriodWeekly:
			platformRows, platformErr = s.usageRepo.GetWeeklyStats(userID, deptID, startDate, endDate)
		case model.PeriodMonthly:
			platformRows, platformErr = s.usageRepo.GetMonthlyStats(userID, deptID, startDate, endDate)
		}
	}()

	go func() {
		defer wg.Done()
		switch query.Period {
		case model.PeriodDaily:
			tpRows, tpErr = s.usageRepo.GetThirdPartyDailyStats(userID, deptID, startDate, endDate)
		case model.PeriodWeekly:
			tpRows, tpErr = s.usageRepo.GetThirdPartyWeeklyStats(userID, deptID, startDate, endDate)
		case model.PeriodMonthly:
			tpRows, tpErr = s.usageRepo.GetThirdPartyMonthlyStats(userID, deptID, startDate, endDate)
		}
	}()

	wg.Wait()

	if platformErr != nil {
		s.logger.Error("failed to query platform stats", zap.Error(platformErr))
		return nil, errcode.ErrDatabase
	}
	if tpErr != nil {
		s.logger.Warn("failed to query third-party stats, skipping", zap.Error(tpErr))
	}

	items := s.mergeUsageStats(platformRows, tpRows)

	return &dto.UsageResponse{
		Period: query.Period,
		Items:  items,
	}, nil
}

// mergeUsageStats merges platform and third-party usage by date.
func (s *StatsService) mergeUsageStats(platform, thirdParty []repository.DailyStatRow) []dto.UsageItem {
	dateMap := make(map[string]*dto.UsageItem, len(platform))

	for _, row := range platform {
		dateStr := row.UsageDate.Format("2006-01-02")
		dateMap[dateStr] = &dto.UsageItem{
			Date:                     dateStr,
			PromptTokens:             row.PromptTokens,
			CompletionTokens:         row.CompletionTokens,
			TotalTokens:              row.TotalTokens,
			RequestCount:             row.RequestCount,
			CacheCreationInputTokens: row.CacheCreationInputTokens,
			CacheReadInputTokens:     row.CacheReadInputTokens,
		}
	}

	for _, row := range thirdParty {
		dateStr := row.UsageDate.Format("2006-01-02")
		item, ok := dateMap[dateStr]
		if !ok {
			item = &dto.UsageItem{Date: dateStr}
			dateMap[dateStr] = item
		}
		item.ThirdPartyPromptTokens = row.PromptTokens
		item.ThirdPartyCompletionTokens = row.CompletionTokens
		item.ThirdPartyTotalTokens = row.TotalTokens
		item.ThirdPartyRequestCount = row.RequestCount
		item.ThirdPartyCacheCreationInputTokens = row.CacheCreationInputTokens
		item.ThirdPartyCacheReadInputTokens = row.CacheReadInputTokens
	}

	items := make([]dto.UsageItem, 0, len(dateMap))
	for _, item := range dateMap {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Date < items[j].Date
	})

	return items
}

// GetRanking returns usage ranking.
func (s *StatsService) GetRanking(query *dto.RankingQuery, operatorDeptID *int64) ([]dto.RankingItem, error) {
	now := timezone.Now()
	startDate, endDate := s.getPeriodRange(now, query.Period)
	limit := query.GetLimit()

	var rows []repository.RankingRow
	var err error

	switch query.Type {
	case model.RoleUser:
		rows, err = s.usageRepo.GetUserRanking(operatorDeptID, startDate, endDate, limit)
	case "department":
		rows, err = s.usageRepo.GetDeptRanking(startDate, endDate, limit)
	default:
		return nil, errcode.ErrInvalidParams.WithMessage("invalid ranking type")
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

// parseDateRange parses date range with defaults when not specified.
func (s *StatsService) parseDateRange(startStr, endStr, period string) (time.Time, time.Time) {
	now := timezone.Now()
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
		case model.PeriodDaily:
			startDate = now.AddDate(0, 0, -30)
		case model.PeriodWeekly:
			startDate = now.AddDate(0, -3, 0)
		case model.PeriodMonthly:
			startDate = now.AddDate(-1, 0, 0)
		default:
			startDate = now.AddDate(0, 0, -30)
		}
	}

	return startDate, endDate
}

// getPeriodRange returns date range based on current time and period.
func (s *StatsService) getPeriodRange(now time.Time, period string) (time.Time, time.Time) {
	switch period {
	case model.PeriodDaily:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	case model.PeriodWeekly:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -(weekday - 1))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, now.Location())
		return start, now
	case model.PeriodMonthly:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, now
	default:
		return now.AddDate(0, 0, -30), now
	}
}

// GetKeyUsageSummary returns API key usage summary.
func (s *StatsService) GetKeyUsageSummary(query *dto.KeyUsageQuery, operatorRole string, operatorID int64, operatorDeptID *int64) ([]dto.KeyUsageItem, error) {
	var userID *int64
	var deptID *int64

	switch operatorRole {
	case model.RoleSuperAdmin:
	case model.RoleDeptManager:
		deptID = operatorDeptID
	default:
		userID = &operatorID
	}

	startDate, endDate := s.parseDateRange(query.StartDate, query.EndDate, "")

	var platformRows, tpRows []repository.KeyUsageRow
	var platformErr, tpErr error

	var wg sync.WaitGroup
	wg.Add(2) //nolint:mnd // intentional constant.

	go func() {
		defer wg.Done()
		platformRows, platformErr = s.usageRepo.GetKeyUsageSummary(userID, deptID, startDate, endDate)
	}()

	go func() {
		defer wg.Done()
		tpRows, tpErr = s.usageRepo.GetThirdPartyKeyUsageSummary(userID, deptID, startDate, endDate)
	}()

	wg.Wait()

	if platformErr != nil {
		s.logger.Error("failed to query platform key usage", zap.Error(platformErr))
		return nil, errcode.ErrDatabase
	}
	if tpErr != nil {
		s.logger.Warn("failed to query third-party key usage, skipping", zap.Error(tpErr))
	}

	items := s.mergeKeyUsageStats(platformRows, tpRows)

	return items, nil
}

// mergeKeyUsageStats merges platform and third-party key usage.
func (s *StatsService) mergeKeyUsageStats(platform, thirdParty []repository.KeyUsageRow) []dto.KeyUsageItem {
	keyMap := make(map[int64]*dto.KeyUsageItem)

	for _, row := range platform {
		keyMap[row.ID] = &dto.KeyUsageItem{
			ID:               row.ID,
			Name:             row.Name,
			PromptTokens:     row.PromptTokens,
			CompletionTokens: row.CompletionTokens,
			TotalTokens:      row.TotalTokens,
			RequestCount:     row.RequestCount,
		}
	}

	for _, row := range thirdParty {
		if item, ok := keyMap[row.ID]; ok {
			item.PromptTokens += row.PromptTokens
			item.CompletionTokens += row.CompletionTokens
			item.TotalTokens += row.TotalTokens
			item.RequestCount += row.RequestCount
		} else {
			keyMap[row.ID] = &dto.KeyUsageItem{
				ID:               row.ID,
				Name:             row.Name,
				PromptTokens:     row.PromptTokens,
				CompletionTokens: row.CompletionTokens,
				TotalTokens:      row.TotalTokens,
				RequestCount:     row.RequestCount,
			}
		}
	}

	items := make([]dto.KeyUsageItem, 0, len(keyMap))
	for _, item := range keyMap {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].TotalTokens > items[j].TotalTokens
	})

	return items
}

// ExportUsageStats exports detailed usage statistics for CSV export.
func (s *StatsService) ExportUsageStats(query *dto.StatsQuery, operatorRole string, operatorID int64, operatorDeptID *int64) ([]dto.UsageExportItem, error) {
	var userID *int64
	var deptID *int64

	switch operatorRole {
	case model.RoleSuperAdmin:
		userID = query.UserID
		deptID = query.DepartmentID
	case model.RoleDeptManager:
		if query.UserID != nil {
			userID = query.UserID
		}
		deptID = operatorDeptID
	default:
		userID = &operatorID
	}

	startDate, endDate := s.parseDateRange(query.StartDate, query.EndDate, query.Period)

	rows, err := s.usageRepo.GetDetailedUsageStats(userID, deptID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to query export data", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	items := make([]dto.UsageExportItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.UsageExportItem{
			Date:             row.UsageDate.Format("2006-01-02"),
			Username:         row.UserName,
			Department:       row.DeptName,
			PromptTokens:     row.PromptTokens,
			CompletionTokens: row.CompletionTokens,
			TotalTokens:      row.TotalTokens,
			RequestCount:     row.RequestCount,
		})
	}

	return items, nil
}
