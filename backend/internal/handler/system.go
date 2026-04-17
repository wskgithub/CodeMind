package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// SystemHandler handles system management endpoints.
type SystemHandler struct {
	systemService SystemService
}

// NewSystemHandler creates a new system handler.
func NewSystemHandler(systemService SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

// GetConfigs returns system configurations.
// GET /api/v1/system/configs
func (h *SystemHandler) GetConfigs(c *gin.Context) {
	configs, err := h.systemService.GetConfigs()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, configs)
}

// UpdateConfigs updates system configurations.
// PUT /api/v1/system/configs
func (h *SystemHandler) UpdateConfigs(c *gin.Context) {
	var req dto.UpdateConfigsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format")
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.systemService.UpdateConfigs(&req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetPlatformServiceURL returns platform service URL.
// GET /api/v1/settings/platform
func (h *SystemHandler) GetPlatformServiceURL(c *gin.Context) {
	url := h.systemService.GetPlatformServiceURL()
	response.Success(c, gin.H{
		"service_url":          url,
		"openai_base_url":     url + "/api/openai/v1",
		"anthropic_base_url":  url + "/api/anthropic",
	})
}

// ListAnnouncements returns announcement list.
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

// CreateAnnouncement creates an announcement.
// POST /api/v1/system/announcements
func (h *SystemHandler) CreateAnnouncement(c *gin.Context) {
	var req dto.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
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

// UpdateAnnouncement updates an announcement.
// PUT /api/v1/system/announcements/:id
func (h *SystemHandler) UpdateAnnouncement(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid announcement ID")
		return
	}

	var req dto.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format")
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.systemService.UpdateAnnouncement(id, &req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteAnnouncement deletes an announcement.
// DELETE /api/v1/system/announcements/:id
func (h *SystemHandler) DeleteAnnouncement(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid announcement ID")
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.systemService.DeleteAnnouncement(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// ListAuditLogs returns audit logs.
// GET /api/v1/system/audit-logs
func (h *SystemHandler) ListAuditLogs(c *gin.Context) {
	var query dto.AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format")
		return
	}

	logs, total, err := h.systemService.ListAuditLogs(&query)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPage(c, logs, total, query.GetPage(), query.GetPageSize())
}
