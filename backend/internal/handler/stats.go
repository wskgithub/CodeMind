package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
)

// StatsHandler 用量统计控制器
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler 创建统计 Handler
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// Overview 获取用量总览
// GET /api/v1/stats/overview
func (h *StatsHandler) Overview(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	deptID := middleware.GetDepartmentID(c)

	overview, err := h.statsService.GetOverview(&userID, role, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, overview)
}

// Usage 获取用量统计数据
// GET /api/v1/stats/usage
func (h *StatsHandler) Usage(c *gin.Context) {
	var query dto.StatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	deptID := middleware.GetDepartmentID(c)

	data, err := h.statsService.GetUsageStats(&query, role, userID, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, data)
}

// Ranking 获取用量排行榜
// GET /api/v1/stats/ranking
func (h *StatsHandler) Ranking(c *gin.Context) {
	var query dto.RankingQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误: "+err.Error())
		return
	}

	deptID := middleware.GetDepartmentID(c)
	role := middleware.GetUserRole(c)

	// 部门经理只能看本部门排行
	var filterDeptID *int64
	if role == "dept_manager" {
		filterDeptID = deptID
	}

	items, err := h.statsService.GetRanking(&query, filterDeptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, items)
}
