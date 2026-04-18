package handler

import (
	"strconv"

	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MCPAdminHandler handles MCP service management.
type MCPAdminHandler struct {
	mcpService *service.MCPService
	logger     *zap.Logger
}

// NewMCPAdminHandler creates a new MCP admin handler.
func NewMCPAdminHandler(mcpService *service.MCPService, logger *zap.Logger) *MCPAdminHandler {
	return &MCPAdminHandler{
		mcpService: mcpService,
		logger:     logger,
	}
}

// ListServices returns the list of MCP services (GET /api/v1/mcp/services).
func (h *MCPAdminHandler) ListServices(c *gin.Context) {
	status := c.Query("status")
	services, err := h.mcpService.ListServices(status)
	if err != nil {
		h.logger.Error("failed to list MCP services", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Success(c, services)
}

// CreateService creates a new MCP service (POST /api/v1/mcp/services).
func (h *MCPAdminHandler) CreateService(c *gin.Context) {
	var req dto.CreateMCPServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid parameters: "+err.Error())
		return
	}

	svc, err := h.mcpService.CreateService(&req)
	if err != nil {
		h.logger.Error("failed to create MCP service", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, svc)
}

// UpdateService updates an MCP service (PUT /api/v1/mcp/services/:id).
func (h *MCPAdminHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid service ID")
		return
	}

	var req dto.UpdateMCPServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid parameters: "+err.Error())
		return
	}

	if err := h.mcpService.UpdateService(id, &req); err != nil {
		h.logger.Error("failed to update MCP service", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteService deletes an MCP service (DELETE /api/v1/mcp/services/:id).
func (h *MCPAdminHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid service ID")
		return
	}

	if err := h.mcpService.DeleteService(id); err != nil {
		h.logger.Error("failed to delete MCP service", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// SyncTools synchronizes tools for an MCP service (POST /api/v1/mcp/services/:id/sync).
func (h *MCPAdminHandler) SyncTools(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid service ID")
		return
	}

	if err := h.mcpService.SyncTools(id); err != nil {
		h.logger.Error("failed to sync MCP tools", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetServiceTools returns the tool list for an MCP service (GET /api/v1/mcp/services/:id/tools).
func (h *MCPAdminHandler) GetServiceTools(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid service ID")
		return
	}

	tools, err := h.mcpService.GetServiceTools(id)
	if err != nil {
		h.logger.Error("failed to get MCP tools", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, tools)
}

// ListAccessRules returns the list of MCP access rules (GET /api/v1/mcp/access-rules).
func (h *MCPAdminHandler) ListAccessRules(c *gin.Context) {
	serviceIDStr := c.Query("service_id")
	var serviceID int64
	if serviceIDStr != "" {
		var err error
		serviceID, err = strconv.ParseInt(serviceIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid service ID")
			return
		}
	}

	rules, err := h.mcpService.ListAccessRules(serviceID)
	if err != nil {
		h.logger.Error("failed to get MCP access rules", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, rules)
}

// SetAccessRule creates or updates an MCP access rule (POST /api/v1/mcp/access-rules).
func (h *MCPAdminHandler) SetAccessRule(c *gin.Context) {
	var req dto.SetMCPAccessRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid parameters: "+err.Error())
		return
	}

	if err := h.mcpService.SetAccessRule(&req); err != nil {
		h.logger.Error("failed to set MCP access rule", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteAccessRule deletes an MCP access rule (DELETE /api/v1/mcp/access-rules/:id).
func (h *MCPAdminHandler) DeleteAccessRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid rule ID")
		return
	}

	if err := h.mcpService.DeleteAccessRule(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
