// Package handler 监控处理器
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

// MonitorHandler 监控处理器
type MonitorHandler struct {
	monitorService *service.MonitorService
	logger         *zap.Logger
}

// NewMonitorHandler 创建监控处理器
func NewMonitorHandler(monitorService *service.MonitorService, logger *zap.Logger) *MonitorHandler {
	return &MonitorHandler{
		monitorService: monitorService,
		logger:         logger,
	}
}

// DashboardSummary 获取仪表盘汇总数据
// GET /api/v1/monitor/dashboard
func (h *MonitorHandler) DashboardSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	summary, err := h.monitorService.GetDashboardSummary(ctx)
	if err != nil {
		h.logger.Error("获取仪表盘数据失败", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, summary)
}

// SystemMetrics 获取系统资源指标
// GET /api/v1/monitor/system
func (h *MonitorHandler) SystemMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	summary, err := h.monitorService.GetSystemMetricsSummary(ctx)
	if err != nil {
		h.logger.Error("获取系统指标失败", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, summary)
}

// RequestMetrics 获取请求性能指标
// GET /api/v1/monitor/requests
func (h *MonitorHandler) RequestMetrics(c *gin.Context) {
	duration := 5 * time.Minute
	if d := c.Query("duration"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			duration = parsed
		}
	}

	metrics, err := h.monitorService.GetRequestMetrics(c.Request.Context(), duration)
	if err != nil {
		h.logger.Error("获取请求指标失败", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, metrics)
}

// LLMNodeMetrics 获取 LLM 节点指标
// GET /api/v1/monitor/llm-nodes
func (h *MonitorHandler) LLMNodeMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	nodes, err := h.monitorService.GetLLMNodeSummaries(ctx)
	if err != nil {
		h.logger.Error("获取 LLM 节点指标失败", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, nodes)
}

// HealthCheck 健康检查端点（用于外部监控）
// GET /api/v1/monitor/health
func (h *MonitorHandler) HealthCheck(c *gin.Context) {
	// 检查数据库连接
	// TODO: 可以添加更多健康检查项

	response.Success(c, gin.H{
		"status":    "healthy",
		"hostname":  h.monitorService.GetHostname(),
		"timestamp": time.Now().Unix(),
	})
}

// ==================== LLM 节点上报接口 ====================

// LLMNodeReport LLM 节点指标上报
// POST /api/v1/monitor/nodes/report
func (h *MonitorHandler) LLMNodeReport(c *gin.Context) {
	var req monitor.NodeReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.NodeID == "" {
		response.BadRequest(c, "node_id 不能为空")
		return
	}

	// 设置默认上报时间
	if req.Timestamp == 0 {
		req.Timestamp = time.Now().Unix()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.monitorService.ReportLLMNodeMetrics(ctx, &req); err != nil {
		h.logger.Error("保存 LLM 节点指标失败", 
			zap.Error(err), 
			zap.String("node_id", req.NodeID))
		response.InternalError(c)
		return
	}

	response.Success(c, gin.H{"message": "上报成功"})
}

// ==================== 内部使用接口（供中间件调用） ====================

// RecordRequestMetrics 记录请求指标（由中间件调用）
func (h *MonitorHandler) RecordRequestMetrics(statusCode int, responseTimeMs float64) {
	h.monitorService.RecordRequest(statusCode, responseTimeMs)
}
