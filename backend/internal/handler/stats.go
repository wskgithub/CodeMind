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

// StatsHandler handles usage statistics endpoints.
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler creates a new stats handler.
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// Overview returns the statistics overview (GET /api/v1/stats/overview).
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

// Usage returns usage statistics (GET /api/v1/stats/usage).
func (h *StatsHandler) Usage(c *gin.Context) {
	var query dto.StatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format: "+err.Error())
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

// Ranking returns usage rankings (GET /api/v1/stats/ranking).
func (h *StatsHandler) Ranking(c *gin.Context) {
	var query dto.RankingQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format: "+err.Error())
		return
	}

	deptID := middleware.GetDepartmentID(c)
	role := middleware.GetUserRole(c)

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

// KeyUsageSummary returns API Key usage summary (GET /api/v1/stats/key-usage).
func (h *StatsHandler) KeyUsageSummary(c *gin.Context) {
	var query dto.KeyUsageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	deptID := middleware.GetDepartmentID(c)

	data, err := h.statsService.GetKeyUsageSummary(&query, role, userID, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, data)
}

// ExportCSV exports usage statistics as CSV (GET /api/v1/stats/export/csv).
func (h *StatsHandler) ExportCSV(c *gin.Context) {
	var query dto.StatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	deptID := middleware.GetDepartmentID(c)

	data, err := h.statsService.ExportUsageStats(&query, role, userID, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	filename := fmt.Sprintf("usage_report_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("X-Content-Type-Options", "nosniff")

	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := []string{"Date", "Username", "Department", "Prompt Tokens", "Completion Tokens", "Total Tokens", "Request Count"}
	if err := writer.Write(headers); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

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
