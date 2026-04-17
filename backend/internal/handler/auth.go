package handler

import (
	"strconv"

	"codemind/internal/middleware"
	"codemind/internal/model/dto"
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates an auth handler
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login authenticates a user
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "username and password required")
		return
	}

	resp, err := h.authService.Login(&req, c.ClientIP())
	if err != nil {
		if e, ok := err.(*errcode.ErrCode); ok {
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

func (h *AuthHandler) handleLockError(c *gin.Context, username string) {
	lockStatus, err := h.authService.GetLoginLockStatusByUsername(username)
	if err != nil {
		response.Error(c, errcode.ErrAccountLocked)
		return
	}

	c.JSON(errcode.ErrAccountLocked.HTTP, gin.H{
		"code":    errcode.ErrAccountLocked.Code,
		"message": h.formatLockMessage(lockStatus),
		"data":    lockStatus,
	})
}

func (h *AuthHandler) formatLockMessage(status *dto.LoginLockStatusResponse) string {
	if !status.Locked {
		return "too many failed attempts, try again later"
	}
	
	remaining := status.RemainingTime
	if remaining < 60 {
		return "account locked, try again later"
	}
	
	minutes := remaining / 60
	if minutes < 60 {
		return "account locked, try again in " + strconv.FormatInt(minutes, 10) + " minutes"
	}
	
	hours := minutes / 60
	remainingMinutes := minutes % 60
	if remainingMinutes > 0 {
		return "account locked, try again in " + strconv.FormatInt(hours, 10) + " hours " + strconv.FormatInt(remainingMinutes, 10) + " minutes"
	}
	return "account locked, try again in " + strconv.FormatInt(hours, 10) + " hours"
}

// Logout invalidates the current session
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

// GetProfile returns current user's profile
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

// UpdateProfile updates current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
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

// ChangePassword changes current user's password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "current and new password required")
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
