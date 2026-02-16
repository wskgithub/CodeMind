package handler

import (
	"codemind/internal/middleware"
	"codemind/internal/model"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"
	"codemind/internal/service"

	"github.com/gin-gonic/gin"
)

// LimitHandler 限额管理控制器
type LimitHandler struct {
	limitService *service.LimitService
}

// NewLimitHandler 创建限额 Handler
func NewLimitHandler(limitService *service.LimitService) *LimitHandler {
	return &LimitHandler{limitService: limitService}
}

// List 获取限额配置列表
// GET /api/v1/limits
func (h *LimitHandler) List(c *gin.Context) {
	var query dto.LimitListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误")
		return
	}

	limits, err := h.limitService.List(&query)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, limits)
}

// Upsert 创建或更新限额配置
// PUT /api/v1/limits
func (h *LimitHandler) Upsert(c *gin.Context) {
	var req dto.UpsertRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)

	// 部门经理只能设置用户级限额
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

// GetMyLimits 获取当前用户的限额信息
// GET /api/v1/limits/my
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

// Delete 删除限额配置
// DELETE /api/v1/limits/:id
func (h *LimitHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的限额 ID")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.limitService.Delete(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}
