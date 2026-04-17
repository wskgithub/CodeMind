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

// TrainingDataHandler handles training data management.
type TrainingDataHandler struct {
	trainingService *service.TrainingDataService
	logger          *zap.Logger
}

// NewTrainingDataHandler creates a new training data handler.
func NewTrainingDataHandler(trainingService *service.TrainingDataService, logger *zap.Logger) *TrainingDataHandler {
	return &TrainingDataHandler{
		trainingService: trainingService,
		logger:          logger,
	}
}

// List returns training data list.
// GET /api/v1/training-data
func (h *TrainingDataHandler) List(c *gin.Context) {
	var query dto.TrainingDataQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format")
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

// GetDetail returns training data details.
// GET /api/v1/training-data/:id
func (h *TrainingDataHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid ID")
		return
	}

	data, err := h.trainingService.GetByID(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, data)
}

// UpdateExcluded updates training data excluded status.
// PUT /api/v1/training-data/:id/exclude
func (h *TrainingDataHandler) UpdateExcluded(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid ID")
		return
	}

	var req dto.TrainingDataExcludeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format")
		return
	}

	if err := h.trainingService.UpdateExcluded(id, req.Excluded); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetStats returns training data statistics.
// GET /api/v1/training-data/stats
func (h *TrainingDataHandler) GetStats(c *gin.Context) {
	stats, err := h.trainingService.GetStats()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, stats)
}

// Export exports training data as JSONL format.
// POST /api/v1/training-data/export
func (h *TrainingDataHandler) Export(c *gin.Context) {
	var query dto.TrainingDataExportQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format")
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
		h.logger.Error("failed to export training data", zap.Error(err))
		return
	}

	h.logger.Info("training data export completed", zap.Int("exported", exported))
}
