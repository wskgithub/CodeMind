package handler

import (
	"net/http"
	"strconv"
	"time"

	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"
	"codemind/internal/repository"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TrainingDataHandler 训练数据管理控制器
type TrainingDataHandler struct {
	trainingService *service.TrainingDataService
	logger          *zap.Logger
}

// NewTrainingDataHandler 创建训练数据管理 Handler
func NewTrainingDataHandler(trainingService *service.TrainingDataService, logger *zap.Logger) *TrainingDataHandler {
	return &TrainingDataHandler{
		trainingService: trainingService,
		logger:          logger,
	}
}

// List 获取训练数据列表
// GET /api/v1/training-data
func (h *TrainingDataHandler) List(c *gin.Context) {
	var query dto.TrainingDataQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误")
		return
	}

	filter := repository.TrainingDataFilter{
		Model:       query.Model,
		RequestType: query.RequestType,
		UserID:      query.UserID,
		IsExcluded:  query.IsExcluded,
		Page:        query.GetPage(),
		PageSize:    query.GetPageSize(),
	}

	if query.StartDate != "" {
		if t, err := time.Parse("2006-01-02", query.StartDate); err == nil {
			filter.StartDate = &t
		}
	}
	if query.EndDate != "" {
		if t, err := time.Parse("2006-01-02", query.EndDate); err == nil {
			end := t.Add(24 * time.Hour)
			filter.EndDate = &end
		}
	}

	items, total, err := h.trainingService.List(filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPage(c, items, total, query.GetPage(), query.GetPageSize())
}

// GetDetail 获取训练数据详情
// GET /api/v1/training-data/:id
func (h *TrainingDataHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	data, err := h.trainingService.GetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, data)
}

// UpdateExcluded 更新训练数据排除状态
// PUT /api/v1/training-data/:id/exclude
func (h *TrainingDataHandler) UpdateExcluded(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	var req dto.TrainingDataExcludeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.trainingService.UpdateExcluded(id, req.Excluded); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetStats 获取训练数据统计
// GET /api/v1/training-data/stats
func (h *TrainingDataHandler) GetStats(c *gin.Context) {
	stats, err := h.trainingService.GetStats()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, stats)
}

// Export 导出训练数据为 JSONL 格式
// POST /api/v1/training-data/export
func (h *TrainingDataHandler) Export(c *gin.Context) {
	var query dto.TrainingDataExportQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误")
		return
	}

	filter := repository.TrainingDataFilter{
		Model:       query.Model,
		RequestType: query.RequestType,
	}

	if query.StartDate != "" {
		if t, err := time.Parse("2006-01-02", query.StartDate); err == nil {
			filter.StartDate = &t
		}
	}
	if query.EndDate != "" {
		if t, err := time.Parse("2006-01-02", query.EndDate); err == nil {
			end := t.Add(24 * time.Hour)
			filter.EndDate = &end
		}
	}

	filename := h.trainingService.ExportFilename()
	c.Writer.Header().Set("Content-Type", "application/x-ndjson")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Status(http.StatusOK)

	exported, err := h.trainingService.ExportJSONL(filter, c.Writer)
	if err != nil {
		h.logger.Error("导出训练数据失败", zap.Error(err))
		return
	}

	h.logger.Info("训练数据导出完成", zap.Int("exported", exported))
}
