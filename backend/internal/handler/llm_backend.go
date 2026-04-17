package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
)

// LLMBackendHandler handles LLM backend node management.
type LLMBackendHandler struct {
	backendService *service.LLMBackendService
}

// NewLLMBackendHandler creates a new handler.
func NewLLMBackendHandler(backendService *service.LLMBackendService) *LLMBackendHandler {
	return &LLMBackendHandler{backendService: backendService}
}

// List handles GET /api/v1/system/llm-backends requests.
func (h *LLMBackendHandler) List(c *gin.Context) {
	backends, err := h.backendService.List()
	if err != nil {
		handleServiceError(c, err)
		return
	}

	items := make([]dto.LLMBackendResponse, len(backends))
	for i, b := range backends {
		items[i] = dto.LLMBackendResponse{
			ID:                   b.ID,
			Name:                 b.Name,
			DisplayName:          b.DisplayName,
			BaseURL:              b.BaseURL,
			HasAPIKey:            b.APIKey != "",
			Format:               b.Format,
			Weight:               b.Weight,
			MaxConcurrency:       b.MaxConcurrency,
			Status:               b.Status,
			Healthy:              true,
			HealthCheckURL:       b.HealthCheckURL,
			TimeoutSeconds:       b.TimeoutSeconds,
			StreamTimeoutSeconds: b.StreamTimeoutSeconds,
			ModelPatterns:        b.ModelPatterns,
			CreatedAt:            b.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:            b.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Success(c, items)
}

// Create handles POST /api/v1/system/llm-backends requests.
func (h *LLMBackendHandler) Create(c *gin.Context) {
	var req dto.CreateLLMBackendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.backendService.Create(&req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// Update handles PUT /api/v1/system/llm-backends/:id requests.
func (h *LLMBackendHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	var req dto.UpdateLLMBackendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.backendService.Update(id, &req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// Delete handles DELETE /api/v1/system/llm-backends/:id requests.
func (h *LLMBackendHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid node ID")
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.backendService.Delete(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}
