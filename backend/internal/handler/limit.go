package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// LimitHandler handles limit management endpoints.
type LimitHandler struct {
	limitService LimitService
}

// NewLimitHandler creates a new limit handler.
func NewLimitHandler(limitService LimitService) *LimitHandler {
	return &LimitHandler{limitService: limitService}
}

// List 获取限流规则列表 (GET /api/v1/limits)。
func (h *LimitHandler) List(c *gin.Context) {
	var query dto.LimitListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query format")
		return
	}

	limits, err := h.limitService.List(&query)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, limits)
}

// Upsert 创建或更新限流规则 (PUT /api/v1/limits)。
func (h *LimitHandler) Upsert(c *gin.Context) {
	var req dto.UpsertRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request format: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)

	if operatorRole == model.RoleDeptManager && req.TargetType != model.TargetTypeUser {
		response.Error(c, errcode.ErrForbidden)
		return
	}

	if err := h.limitService.Upsert(&req, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetMyLimits 获取当前用户的限流配置 (GET /api/v1/limits/my)。
func (h *LimitHandler) GetMyLimits(c *gin.Context) {
	userID := middleware.GetUserID(c)
	deptID := middleware.GetDepartmentID(c)

	data, err := h.limitService.GetMyLimits(userID, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, data)
}

// GetMyProgress 获取当前用户的限流使用进度 (GET /api/v1/limits/my/progress)。
func (h *LimitHandler) GetMyProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	deptID := middleware.GetDepartmentID(c)

	data, err := h.limitService.GetLimitProgress(userID, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, data)
}

// Delete 删除限流规则 (DELETE /api/v1/limits/:id)。
func (h *LimitHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid limit ID")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.limitService.Delete(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}
