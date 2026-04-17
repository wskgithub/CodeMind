package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler handles API key management endpoints
type APIKeyHandler struct {
	keyService APIKeyService
}

// NewAPIKeyHandler creates an API key handler
func NewAPIKeyHandler(keyService APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{keyService: keyService}
}

// List returns current user's API keys
func (h *APIKeyHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	keys, err := h.keyService.List(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, keys)
}

// Create creates a new API key
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	key, err := h.keyService.Create(&req, userID, c.ClientIP())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, key)
}

// UpdateStatus toggles API key status
func (h *APIKeyHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid key ID")
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)
	operatorDeptID := middleware.GetDepartmentID(c)

	if err := h.keyService.UpdateStatus(id, req.Status, operatorID, operatorRole, operatorDeptID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// Copy returns the full API key for copying
func (h *APIKeyHandler) Copy(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid key ID")
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)

	resp, err := h.keyService.Copy(id, operatorID, operatorRole, c.ClientIP())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, resp)
}

// Delete deletes an API key
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid key ID")
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)

	if err := h.keyService.Delete(id, operatorID, operatorRole, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}
