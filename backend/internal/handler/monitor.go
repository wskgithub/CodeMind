package handler

import (
	"context"
	"time"

	"codemind/internal/model/monitor"
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MonitorHandler handles monitoring endpoints.
type MonitorHandler struct {
	monitorService *service.MonitorService
	logger         *zap.Logger
}

// NewMonitorHandler creates a new monitor handler.
func NewMonitorHandler(monitorService *service.MonitorService, logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{
		monitorService: monitorService,
		logger:         logger,
	}
}

// DashboardSummary handles GET /api/v1/monitor/dashboard requests.
func (h *MonitorHandler) DashboardSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	summary, err := h.monitorService.GetDashboardSummary(ctx)
	if err != nil {
		h.logger.Error("failed to get dashboard data", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, summary)
}

// SystemMetrics handles GET /api/v1/monitor/system requests.
func (h *MonitorHandler) SystemMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	summary, err := h.monitorService.GetSystemMetricsSummary(ctx)
	if err != nil {
		h.logger.Error("failed to get system metrics", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, summary)
}

// RequestMetrics handles GET /api/v1/monitor/requests requests.
func (h *MonitorHandler) RequestMetrics(c *gin.Context) {
	duration := 5 * time.Minute //nolint:mnd // intentional constant.
	if d := c.Query("duration"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			duration = parsed
		}
	}

	metrics, err := h.monitorService.GetRequestMetrics(c.Request.Context(), duration)
	if err != nil {
		h.logger.Error("failed to get request metrics", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, metrics)
}

// LLMNodeMetrics returns LLM node metrics (GET /api/v1/monitor/llm-nodes).
func (h *MonitorHandler) LLMNodeMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	nodes, err := h.monitorService.GetLLMNodeSummaries(ctx)
	if err != nil {
		h.logger.Error("failed to get LLM node metrics", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, nodes)
}

// HealthCheck performs a health check (GET /api/v1/monitor/health).
func (h *MonitorHandler) HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status":    "healthy",
		"hostname":  h.monitorService.GetHostname(),
		"timestamp": time.Now().Unix(),
	})
}

// LLMNodeReport reports LLM node status (POST /api/v1/monitor/nodes/report).
func (h *MonitorHandler) LLMNodeReport(c *gin.Context) {
	var req monitor.NodeReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid parameters: "+err.Error())
		return
	}

	if req.NodeID == "" {
		response.BadRequest(c, "node_id cannot be empty")
		return
	}

	if req.Timestamp == 0 {
		req.Timestamp = time.Now().Unix()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) //nolint:mnd // intentional constant.
	defer cancel()

	if err := h.monitorService.ReportLLMNodeMetrics(ctx, &req); err != nil {
		h.logger.Error("failed to save LLM node metrics",
			zap.Error(err),
			zap.String("node_id", req.NodeID))
		response.InternalError(c)
		return
	}

	response.Success(c, gin.H{"message": "reported"})
}

// RecordRequestMetrics records request metrics (called by middleware).
func (h *MonitorHandler) RecordRequestMetrics(statusCode int, responseTimeMs float64) {
	h.monitorService.RecordRequest(statusCode, responseTimeMs)
}
