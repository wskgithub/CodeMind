package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统管理控制器
type SystemHandler struct {
	systemService SystemService
}

// NewSystemHandler 创建系统管理 Handler
func NewSystemHandler(systemService SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

// ──────────────────────────────────
// 系统配置
// ──────────────────────────────────

// GetConfigs 获取系统配置
// GET /api/v1/system/configs
func (h *SystemHandler) GetConfigs(c *gin.Context) {
	configs, err := h.systemService.GetConfigs()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, configs)
}

// UpdateConfigs 更新系统配置
// PUT /api/v1/system/configs
func (h *SystemHandler) UpdateConfigs(c *gin.Context) {
	var req dto.UpdateConfigsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.systemService.UpdateConfigs(&req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// ──────────────────────────────────
// 公告管理
// ──────────────────────────────────

// ListAnnouncements 获取公告列表
// GET /api/v1/system/announcements
func (h *SystemHandler) ListAnnouncements(c *gin.Context) {
	role := middleware.GetUserRole(c)
	isAdmin := role == model.RoleSuperAdmin

	anns, err := h.systemService.ListAnnouncements(isAdmin)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, anns)
}

// CreateAnnouncement 创建公告
// POST /api/v1/system/announcements
func (h *SystemHandler) CreateAnnouncement(c *gin.Context) {
	var req dto.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	authorID := middleware.GetUserID(c)
	ann, err := h.systemService.CreateAnnouncement(&req, authorID, c.ClientIP())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, ann)
}

// UpdateAnnouncement 更新公告
// PUT /api/v1/system/announcements/:id
func (h *SystemHandler) UpdateAnnouncement(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的公告 ID")
		return
	}

	var req dto.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.systemService.UpdateAnnouncement(id, &req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteAnnouncement 删除公告
// DELETE /api/v1/system/announcements/:id
func (h *SystemHandler) DeleteAnnouncement(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的公告 ID")
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.systemService.DeleteAnnouncement(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// ──────────────────────────────────
// 审计日志
// ──────────────────────────────────

// ListAuditLogs 获取审计日志
// GET /api/v1/system/audit-logs
func (h *SystemHandler) ListAuditLogs(c *gin.Context) {
	var query dto.AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误")
		return
	}

	logs, total, err := h.systemService.ListAuditLogs(&query)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPage(c, logs, total, query.GetPage(), query.GetPageSize())
}
