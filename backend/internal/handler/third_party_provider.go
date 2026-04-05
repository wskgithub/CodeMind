package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ThirdPartyProviderHandler 第三方模型服务控制器
type ThirdPartyProviderHandler struct {
	service ThirdPartyProviderService
}

// NewThirdPartyProviderHandler 创建 Handler 实例
func NewThirdPartyProviderHandler(service ThirdPartyProviderService) *ThirdPartyProviderHandler {
	return &ThirdPartyProviderHandler{service: service}
}

// ──────────────────────────────────
// 模板管理（管理员接口）
// ──────────────────────────────────

// ListTemplatesAdmin 管理员获取所有模板
// GET /api/v1/system/provider-templates
func (h *ThirdPartyProviderHandler) ListTemplatesAdmin(c *gin.Context) {
	templates, err := h.service.ListTemplates()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, templates)
}

// CreateTemplate 创建模板
// POST /api/v1/system/provider-templates
func (h *ThirdPartyProviderHandler) CreateTemplate(c *gin.Context) {
	var req dto.CreateProviderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	template, err := h.service.CreateTemplate(
		req.Name, req.OpenAIBaseURL, req.AnthropicBaseURL, req.Format,
		req.Models, req.Description, req.Icon,
		req.SortOrder, operatorID,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, template)
}

// UpdateTemplate 更新模板
// PUT /api/v1/system/provider-templates/:id
func (h *ThirdPartyProviderHandler) UpdateTemplate(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "ID 参数格式错误")
		return
	}

	var req dto.UpdateProviderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	if err := h.service.UpdateTemplate(
		id, req.Name, req.OpenAIBaseURL, req.AnthropicBaseURL,
		req.Models, req.Format,
		req.Description, req.Icon,
		req.SortOrder, req.Status,
	); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteTemplate 删除模板
// DELETE /api/v1/system/provider-templates/:id
func (h *ThirdPartyProviderHandler) DeleteTemplate(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "ID 参数格式错误")
		return
	}

	if err := h.service.DeleteTemplate(id); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// ──────────────────────────────────
// 用户第三方服务管理
// ──────────────────────────────────

// ListProviders 获取当前用户的第三方服务列表
// GET /api/v1/models/third-party
func (h *ThirdPartyProviderHandler) ListProviders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	providers, err := h.service.ListProviders(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, providers)
}

// CreateProvider 用户添加第三方服务
// POST /api/v1/models/third-party
func (h *ThirdPartyProviderHandler) CreateProvider(c *gin.Context) {
	var req dto.CreateThirdPartyProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	provider, err := h.service.CreateProvider(
		userID, req.Name, req.OpenAIBaseURL, req.AnthropicBaseURL,
		req.APIKey, req.Format, req.Models, req.TemplateID,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, provider)
}

// UpdateProvider 更新第三方服务
// PUT /api/v1/models/third-party/:id
func (h *ThirdPartyProviderHandler) UpdateProvider(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "ID 参数格式错误")
		return
	}

	var req dto.UpdateThirdPartyProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.service.UpdateProvider(
		id, userID, req.Name, req.OpenAIBaseURL, req.AnthropicBaseURL,
		req.APIKey, req.Models, req.Format, req.Status,
	); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// UpdateProviderStatus 切换第三方服务状态
// PUT /api/v1/models/third-party/:id/status
func (h *ThirdPartyProviderHandler) UpdateProviderStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "ID 参数格式错误")
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.service.UpdateProviderStatus(id, userID, req.Status); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteProvider 删除第三方服务
// DELETE /api/v1/models/third-party/:id
func (h *ThirdPartyProviderHandler) DeleteProvider(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "ID 参数格式错误")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.service.DeleteProvider(id, userID); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// ListTemplatesForUser 获取可用模板列表（用户选择）
// GET /api/v1/models/templates
func (h *ThirdPartyProviderHandler) ListTemplatesForUser(c *gin.Context) {
	templates, err := h.service.ListActiveTemplates()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, templates)
}

// ListPlatformModels 获取 CodeMind 平台模型列表
// GET /api/v1/models/platform
func (h *ThirdPartyProviderHandler) ListPlatformModels(c *gin.Context) {
	models, err := h.service.ListPlatformModels()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, models)
}

