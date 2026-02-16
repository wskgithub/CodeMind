package handler

import (
	"strconv"

	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MCPAdminHandler MCP 服务管理控制器
type MCPAdminHandler struct {
	mcpService *service.MCPService
	logger     *zap.Logger
}

// NewMCPAdminHandler 创建 MCP 管理 Handler
func NewMCPAdminHandler(mcpService *service.MCPService, logger *zap.Logger) *MCPAdminHandler {
	return &MCPAdminHandler{
		mcpService: mcpService,
		logger:     logger,
	}
}

// ListServices 获取 MCP 服务列表
// GET /api/v1/mcp/services
func (h *MCPAdminHandler) ListServices(c *gin.Context) {
	status := c.Query("status")
	services, err := h.mcpService.ListServices(status)
	if err != nil {
		h.logger.Error("获取 MCP 服务列表失败", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Success(c, services)
}

// CreateService 创建 MCP 服务
// POST /api/v1/mcp/services
func (h *MCPAdminHandler) CreateService(c *gin.Context) {
	var req dto.CreateMCPServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	svc, err := h.mcpService.CreateService(&req)
	if err != nil {
		h.logger.Error("创建 MCP 服务失败", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, svc)
}

// UpdateService 更新 MCP 服务
// PUT /api/v1/mcp/services/:id
func (h *MCPAdminHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务 ID")
		return
	}

	var req dto.UpdateMCPServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.mcpService.UpdateService(id, &req); err != nil {
		h.logger.Error("更新 MCP 服务失败", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteService 删除 MCP 服务
// DELETE /api/v1/mcp/services/:id
func (h *MCPAdminHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务 ID")
		return
	}

	if err := h.mcpService.DeleteService(id); err != nil {
		h.logger.Error("删除 MCP 服务失败", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// SyncTools 同步 MCP 服务工具列表
// POST /api/v1/mcp/services/:id/sync
func (h *MCPAdminHandler) SyncTools(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务 ID")
		return
	}

	if err := h.mcpService.SyncTools(id); err != nil {
		h.logger.Error("同步 MCP 工具列表失败", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetServiceTools 获取 MCP 服务工具列表
// GET /api/v1/mcp/services/:id/tools
func (h *MCPAdminHandler) GetServiceTools(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务 ID")
		return
	}

	tools, err := h.mcpService.GetServiceTools(id)
	if err != nil {
		h.logger.Error("获取 MCP 工具列表失败", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, tools)
}

// ListAccessRules 获取 MCP 访问规则列表
// GET /api/v1/mcp/access-rules
func (h *MCPAdminHandler) ListAccessRules(c *gin.Context) {
	serviceIDStr := c.Query("service_id")
	var serviceID int64
	if serviceIDStr != "" {
		var err error
		serviceID, err = strconv.ParseInt(serviceIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的服务 ID")
			return
		}
	}

	rules, err := h.mcpService.ListAccessRules(serviceID)
	if err != nil {
		h.logger.Error("获取 MCP 访问规则失败", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Success(c, rules)
}

// SetAccessRule 设置 MCP 访问规则
// POST /api/v1/mcp/access-rules
func (h *MCPAdminHandler) SetAccessRule(c *gin.Context) {
	var req dto.SetMCPAccessRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.mcpService.SetAccessRule(&req); err != nil {
		h.logger.Error("设置 MCP 访问规则失败", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteAccessRule 删除 MCP 访问规则
// DELETE /api/v1/mcp/access-rules/:id
func (h *MCPAdminHandler) DeleteAccessRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的规则 ID")
		return
	}

	if err := h.mcpService.DeleteAccessRule(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
