package handler

import (
	"strconv"

	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证控制器
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "用户名和密码不能为空")
		return
	}

	resp, err := h.authService.Login(&req, c.ClientIP())
	if err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			// 如果是账号锁定错误，尝试获取更详细的锁定信息
			if e.Code == errcode.ErrAccountLocked.Code {
				h.handleLockError(c, req.Username)
				return
			}
			response.Error(c, e)
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, resp)
}

// handleLockError 处理账号锁定错误，返回详细的锁定信息
func (h *AuthHandler) handleLockError(c *gin.Context, username string) {
	// 尝试获取用户锁定状态
	lockStatus, err := h.authService.GetLoginLockStatusByUsername(username)
	if err != nil {
		// 如果获取失败，返回基础锁定信息
		response.Error(c, errcode.ErrAccountLocked)
		return
	}

	// 返回包含锁定详情的错误
	c.JSON(errcode.ErrAccountLocked.HTTP, gin.H{
		"code":    errcode.ErrAccountLocked.Code,
		"message": h.formatLockMessage(lockStatus),
		"data":    lockStatus,
	})
}

// formatLockMessage 格式化锁定提示消息
func (h *AuthHandler) formatLockMessage(status *dto.LoginLockStatusResponse) string {
	if !status.Locked {
		return "登录失败次数过多，请稍后再试"
	}
	
	// 格式化剩余时间
	remaining := status.RemainingTime
	if remaining < 60 {
		return "账号已被锁定，请稍后再试"
	}
	
	minutes := remaining / 60
	if minutes < 60 {
		return "账号已被锁定，请 " + strconv.FormatInt(minutes, 10) + " 分钟后再试"
	}
	
	hours := minutes / 60
	remainingMinutes := minutes % 60
	if remainingMinutes > 0 {
		return "账号已被锁定，请 " + strconv.FormatInt(hours, 10) + " 小时 " + strconv.FormatInt(remainingMinutes, 10) + " 分钟后再试"
	}
	return "账号已被锁定，请 " + strconv.FormatInt(hours, 10) + " 小时后再试"
}

// Logout 用户登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, errcode.ErrTokenInvalid)
		return
	}

	if err := h.authService.Logout(claims); err != nil {
		response.InternalError(c)
		return
	}

	response.Success(c, nil)
}

// GetProfile 获取当前用户信息
// GET /api/v1/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, errcode.ErrTokenInvalid)
		return
	}

	profile, err := h.authService.GetProfile(userID)
	if err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, profile)
}

// UpdateProfile 更新当前用户个人信息
// PUT /api/v1/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if err := h.authService.UpdateProfile(userID, &req); err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, nil)
}

// ChangePassword 修改密码
// PUT /api/v1/auth/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入原密码和新密码")
		return
	}

	claims := middleware.GetClaims(c)
	if err := h.authService.ChangePassword(userID, &req, claims, c.ClientIP()); err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, e)
			return
		}
		response.InternalError(c)
		return
	}

	response.Success(c, nil)
}
