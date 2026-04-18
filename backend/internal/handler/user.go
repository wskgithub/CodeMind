package handler

import (
	"strconv"

	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user management endpoints.
type UserHandler struct {
	userService UserService
}

// NewUserHandler creates a user handler.
func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List returns paginated user list.
func (h *UserHandler) List(c *gin.Context) {
	var query dto.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query parameters")
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

// Create creates a new user.
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
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

// GetDetail returns user details.
func (h *UserHandler) GetDetail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	user, err := h.userService.GetDetail(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, user)
}

// Update updates user information.
func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
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

// Delete soft-deletes a user.
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.userService.Delete(id, operatorID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateStatus toggles user status.
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
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

	if err := h.userService.UpdateStatus(id, req.Status, operatorID, operatorRole, operatorDeptID, c.ClientIP()); err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

// ResetPassword resets a user's password.
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "new password required")
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

// UnlockUser unlocks a locked user account.
func (h *UserHandler) UnlockUser(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	var req dto.UnlockUserRequest
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

func parseID(c *gin.Context) (int64, error) {
	idStr := c.Param("id")
	return strconv.ParseInt(idStr, 10, 64)
}

func handleServiceError(c *gin.Context, err error) {
	if e, ok := err.(*errcode.ErrCode); ok {
		response.Error(c, e)
		return
	}
	response.InternalError(c)
}
