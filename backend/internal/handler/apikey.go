package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler API Key 管理控制器
type APIKeyHandler struct {
	keyService APIKeyService
}

// NewAPIKeyHandler 创建 API Key Handler
func NewAPIKeyHandler(keyService APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{keyService: keyService}
}

// List 获取当前用户的 API Key 列表
// GET /api/v1/keys
func (h *APIKeyHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	keys, err := h.keyService.List(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, keys)
}

// Create 创建新的 API Key
// POST /api/v1/keys
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
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

// UpdateStatus 切换 Key 状态
// PUT /api/v1/keys/:id/status
func (h *APIKeyHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的 Key ID")
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
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

// Copy 复制 API Key（返回完整 Key，但不展示在界面上）
// POST /api/v1/keys/:id/copy
func (h *APIKeyHandler) Copy(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的 Key ID")
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

// Delete 删除 API Key
// DELETE /api/v1/keys/:id
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的 Key ID")
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
