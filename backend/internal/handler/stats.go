package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

// ExportCSV 导出租用量报表为 CSV
// GET /api/v1/stats/export/csv
func (h *StatsHandler) ExportCSV(c *gin.Context) {
	var query dto.StatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	deptID := middleware.GetDepartmentID(c)

	// 获取导出数据
	data, err := h.statsService.ExportUsageStats(&query, role, userID, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 设置响应头
	filename := fmt.Sprintf("usage_report_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("X-Content-Type-Options", "nosniff")

	// 写入 UTF-8 BOM (让 Excel 正确识别中文)
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	// 创建 CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入表头
	headers := []string{"日期", "用户名", "部门", "Prompt Tokens", "Completion Tokens", "总 Tokens", "请求次数"}
	if err := writer.Write(headers); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// 写入数据行
	for _, item := range data {
		record := []string{
			item.Date,
			item.Username,
			item.Department,
			strconv.FormatInt(item.PromptTokens, 10),
			strconv.FormatInt(item.CompletionTokens, 10),
			strconv.FormatInt(item.TotalTokens, 10),
			strconv.FormatInt(item.RequestCount, 10),
		}
		if err := writer.Write(record); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
}
