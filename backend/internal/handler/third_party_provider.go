package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ThirdPartyProviderHandler handles third-party provider endpoints.
type ThirdPartyProviderHandler struct {
	service ThirdPartyProviderService
}

// NewThirdPartyProviderHandler creates a new handler.
func NewThirdPartyProviderHandler(service ThirdPartyProviderService) *ThirdPartyProviderHandler {
	return &ThirdPartyProviderHandler{service: service}
}

// ListTemplatesAdmin returns provider templates for admin (GET /api/v1/system/provider-templates).
func (h *ThirdPartyProviderHandler) ListTemplatesAdmin(c *gin.Context) {
	templates, err := h.service.ListTemplates()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, templates)
}

// CreateTemplate creates a provider template (POST /api/v1/system/provider-templates).
func (h *ThirdPartyProviderHandler) CreateTemplate(c *gin.Context) {
	var req dto.CreateProviderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
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

// UpdateTemplate updates a provider template (PUT /api/v1/system/provider-templates/:id).
func (h *ThirdPartyProviderHandler) UpdateTemplate(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	var req dto.UpdateProviderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
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

// DeleteTemplate deletes a provider template (DELETE /api/v1/system/provider-templates/:id).
func (h *ThirdPartyProviderHandler) DeleteTemplate(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	if err := h.service.DeleteTemplate(id); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// ListProviders returns the user's third-party providers (GET /api/v1/models/third-party).
func (h *ThirdPartyProviderHandler) ListProviders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	providers, err := h.service.ListProviders(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, providers)
}

// CreateProvider creates a third-party provider (POST /api/v1/models/third-party).
func (h *ThirdPartyProviderHandler) CreateProvider(c *gin.Context) {
	var req dto.CreateThirdPartyProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
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

// UpdateProvider updates a third-party provider (PUT /api/v1/models/third-party/:id).
func (h *ThirdPartyProviderHandler) UpdateProvider(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	var req dto.UpdateThirdPartyProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
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

// UpdateProviderStatus updates a third-party provider's status (PUT /api/v1/models/third-party/:id/status).
func (h *ThirdPartyProviderHandler) UpdateProviderStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.service.UpdateProviderStatus(id, userID, req.Status); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteProvider deletes a third-party provider (DELETE /api/v1/models/third-party/:id).
func (h *ThirdPartyProviderHandler) DeleteProvider(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.service.DeleteProvider(id, userID); err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, nil)
}

// ListTemplatesForUser returns active provider templates for users (GET /api/v1/models/templates).
func (h *ThirdPartyProviderHandler) ListTemplatesForUser(c *gin.Context) {
	templates, err := h.service.ListActiveTemplates()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, templates)
}

// ListPlatformModels returns the platform model list (GET /api/v1/models/platform).
func (h *ThirdPartyProviderHandler) ListPlatformModels(c *gin.Context) {
	models, err := h.service.ListPlatformModels()
	if err != nil {
		handleServiceError(c, err)
		return
	}
	response.Success(c, models)
}
