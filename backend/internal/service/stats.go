package service

import (
	"sort"
	"sync"
	"time"

	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/timezone"
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
// 所有独立查询并行执行，将总延迟从 N 次串行往返降低到 1 次
func (s *StatsService) GetOverview(userID *int64, role string, deptID *int64) (*dto.StatsOverview, error) {
	overview := &dto.StatsOverview{SystemStatus: "healthy"}

	var filterUserID *int64
	if role == "user" {
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
	wg.Add(10)

	// 平台用量
	go func() { defer wg.Done(); todayTokens, _ = s.usageRepo.GetTodayTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); todayRequests, _ = s.usageRepo.GetTodayRequestCount(filterUserID) }()
	go func() { defer wg.Done(); todayActive, _ = s.usageRepo.GetTodayActiveUsers() }()
	go func() { defer wg.Done(); monthTokens, _ = s.usageRepo.GetMonthTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); monthRequests, _ = s.usageRepo.GetMonthRequestCount(filterUserID) }()
	go func() { defer wg.Done(); monthActive, _ = s.usageRepo.GetMonthActiveUsers() }()

	// 第三方用量
	go func() { defer wg.Done(); tpTodayTokens, _ = s.usageRepo.GetThirdPartyTodayTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); tpTodayRequests, _ = s.usageRepo.GetThirdPartyTodayRequestCount(filterUserID) }()
	go func() { defer wg.Done(); tpMonthTokens, _ = s.usageRepo.GetThirdPartyMonthTotalTokens(filterUserID) }()
	go func() { defer wg.Done(); tpMonthRequests, _ = s.usageRepo.GetThirdPartyMonthRequestCount(filterUserID) }()

	if role != "user" {
		wg.Add(3)
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

// GetUsageStats 获取用量统计数据（平台 + 第三方并行查询后合并）
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

	if query.Period != "daily" && query.Period != "weekly" && query.Period != "monthly" {
		return nil, errcode.ErrInvalidParams.WithMessage("无效的统计周期")
	}

	// 解析日期范围
	startDate, endDate := s.parseDateRange(query.StartDate, query.EndDate, query.Period)

	// 并行查询平台和第三方用量
	var platformRows, tpRows []repository.DailyStatRow
	var platformErr, tpErr error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		switch query.Period {
		case "daily":
			platformRows, platformErr = s.usageRepo.GetDailyStats(userID, deptID, startDate, endDate)
		case "weekly":
			platformRows, platformErr = s.usageRepo.GetWeeklyStats(userID, deptID, startDate, endDate)
		case "monthly":
			platformRows, platformErr = s.usageRepo.GetMonthlyStats(userID, deptID, startDate, endDate)
		}
	}()

	go func() {
		defer wg.Done()
		switch query.Period {
		case "daily":
			tpRows, tpErr = s.usageRepo.GetThirdPartyDailyStats(userID, deptID, startDate, endDate)
		case "weekly":
			tpRows, tpErr = s.usageRepo.GetThirdPartyWeeklyStats(userID, deptID, startDate, endDate)
		case "monthly":
			tpRows, tpErr = s.usageRepo.GetThirdPartyMonthlyStats(userID, deptID, startDate, endDate)
		}
	}()

	wg.Wait()

	if platformErr != nil {
		s.logger.Error("查询平台统计数据失败", zap.Error(platformErr))
		return nil, errcode.ErrDatabase
	}
	if tpErr != nil {
		s.logger.Warn("查询第三方统计数据失败，跳过第三方数据", zap.Error(tpErr))
	}

	items := s.mergeUsageStats(platformRows, tpRows)

	return &dto.UsageResponse{
		Period: query.Period,
		Items:  items,
	}, nil
}

// mergeUsageStats 将平台和第三方用量按日期合并为统一的 UsageItem 列表
// 缓存数据分别存储：平台缓存存在 CacheCreationInputTokens/CacheReadInputTokens
// 第三方缓存存在 ThirdPartyCacheCreationInputTokens/ThirdPartyCacheReadInputTokens
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
		// 第三方缓存数据单独存储
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

// GetRanking 获取用量排行榜
func (s *StatsService) GetRanking(query *dto.RankingQuery, operatorDeptID *int64) ([]dto.RankingItem, error) {
	now := timezone.Now()
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
// 注意：time.Parse 返回 UTC 时间，这对 DATE 类型比较无影响；
// 未指定日期时使用 Asia/Shanghai 时区的当前时间作为默认范围
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

// GetKeyUsageSummary 获取 Key 用量汇总（平台 + 第三方）
func (s *StatsService) GetKeyUsageSummary(query *dto.KeyUsageQuery, operatorRole string, operatorID int64, operatorDeptID *int64) ([]dto.KeyUsageItem, error) {
	// 确定查询范围
	var userID *int64
	var deptID *int64

	switch operatorRole {
	case "super_admin":
		// 管理员可查看全部，不传 userID 限制
	case "dept_manager":
		deptID = operatorDeptID
	default:
		userID = &operatorID
	}

	// 解析日期范围（period 仅用于默认范围，此处传空字符串使用默认 30 天）
	startDate, endDate := s.parseDateRange(query.StartDate, query.EndDate, "")

	// 并行查询平台用量和第三方用量
	var platformRows, tpRows []repository.KeyUsageRow
	var platformErr, tpErr error

	var wg sync.WaitGroup
	wg.Add(2)

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
		s.logger.Error("查询平台 Key 用量汇总失败", zap.Error(platformErr))
		return nil, errcode.ErrDatabase
	}
	if tpErr != nil {
		s.logger.Warn("查询第三方 Key 用量汇总失败，跳过第三方数据", zap.Error(tpErr))
	}

	// 合并平台和第三方用量
	items := s.mergeKeyUsageStats(platformRows, tpRows)

	return items, nil
}

// mergeKeyUsageStats 合并平台和第三方 Key 用量
func (s *StatsService) mergeKeyUsageStats(platform, thirdParty []repository.KeyUsageRow) []dto.KeyUsageItem {
	keyMap := make(map[int64]*dto.KeyUsageItem)

	// 添加平台用量
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

	// 合并第三方用量
	for _, row := range thirdParty {
		if item, ok := keyMap[row.ID]; ok {
			// 已存在，累加用量
			item.PromptTokens += row.PromptTokens
			item.CompletionTokens += row.CompletionTokens
			item.TotalTokens += row.TotalTokens
			item.RequestCount += row.RequestCount
		} else {
			// 新建条目
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

	// 转换为切片并按总用量排序
	items := make([]dto.KeyUsageItem, 0, len(keyMap))
	for _, item := range keyMap {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].TotalTokens > items[j].TotalTokens
	})

	return items
}

// ExportUsageStats 导出租用量统计数据
// 返回详细的每日用户用量数据，用于 CSV 导出
func (s *StatsService) ExportUsageStats(query *dto.StatsQuery, operatorRole string, operatorID int64, operatorDeptID *int64) ([]dto.UsageExportItem, error) {
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
		// 普通用户只能导出自己的数据
		userID = &operatorID
	}

	// 解析日期范围
	startDate, endDate := s.parseDateRange(query.StartDate, query.EndDate, query.Period)

	// 查询详细数据
	rows, err := s.usageRepo.GetDetailedUsageStats(userID, deptID, startDate, endDate)
	if err != nil {
		s.logger.Error("查询导出数据失败", zap.Error(err))
		return nil, errcode.ErrDatabase
	}

	// 转换为响应格式
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
