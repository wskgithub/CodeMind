package handler

import (
	"strconv"

	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理控制器
type UserHandler struct {
	userService UserService
}

// NewUserHandler 创建用户 Handler
func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List 获取用户列表
// GET /api/v1/users
func (h *UserHandler) List(c *gin.Context) {
	var query dto.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数格式错误")
		return
	}

	role := middleware.GetUserRole(c)
	deptID := middleware.GetDepartmentID(c)

	users, total, err := h.userService.List(&query, role, deptID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.SuccessWithPage(c, users, total, query.GetPage(), query.GetPageSize())
}

// Create 创建用户
// POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)
	operatorDeptID := middleware.GetDepartmentID(c)

	user, err := h.userService.Create(&req, operatorID, operatorRole, operatorDeptID, c.ClientIP())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, user)
}

// GetDetail 获取用户详情
// GET /api/v1/users/:id
func (h *UserHandler) GetDetail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	user, err := h.userService.GetDetail(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, user)
}

// Update 更新用户信息
// PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)
	operatorDeptID := middleware.GetDepartmentID(c)

	if err := h.userService.Update(id, &req, operatorID, operatorRole, operatorDeptID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// Delete 删除用户
// DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.userService.Delete(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateStatus 切换用户状态
// PUT /api/v1/users/:id/status
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
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

	if err := h.userService.UpdateStatus(id, req.Status, operatorID, operatorRole, operatorDeptID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// ResetPassword 重置用户密码
// PUT /api/v1/users/:id/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入新密码")
		return
	}

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)
	operatorDeptID := middleware.GetDepartmentID(c)

	if err := h.userService.ResetPassword(id, req.NewPassword, operatorID, operatorRole, operatorDeptID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// UnlockUser 解锁用户账号
// PUT /api/v1/users/:id/unlock
func (h *UserHandler) UnlockUser(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	var req dto.UnlockUserRequest
	// 请求体是可选的
	_ = c.ShouldBindJSON(&req)

	operatorID := middleware.GetUserID(c)
	operatorRole := middleware.GetUserRole(c)
	operatorDeptID := middleware.GetDepartmentID(c)

	if err := h.userService.UnlockUser(id, operatorID, operatorRole, operatorDeptID, req.Reason, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// ──────────────────────────────────
// 辅助函数
// ──────────────────────────────────

// parseID 从 URL 参数解析 ID
func parseID(c *gin.Context) (int64, error) {
	idStr := c.Param("id")
	return strconv.ParseInt(idStr, 10, 64)
}

// handleServiceError 处理 Service 层返回的错误
func handleServiceError(c *gin.Context, err error) {
	if e, ok := err.(*errcode.ErrCode); ok {
		response.Error(c, e)
		return
	}
	response.InternalError(c)
}
