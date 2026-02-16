package middleware

import (
	"strings"

	jwtPkg "codemind/internal/pkg/jwt"
	"codemind/internal/pkg/response"
	"codemind/internal/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// 上下文键常量
const (
	CtxKeyUserID       = "user_id"
	CtxKeyUsername     = "username"
	CtxKeyRole         = "role"
	CtxKeyDepartmentID = "department_id"
	CtxKeyClaims       = "claims"
)

// JWTAuth JWT 认证中间件
// 从 Authorization Header 提取并验证 JWT Token
func JWTAuth(jwtManager *jwtPkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 提取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			response.Error(c, errcode.ErrTokenInvalid.WithMessage(err.Error()))
			c.Abort()
			return
		}

		// 将用户信息注入上下文
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Set(CtxKeyClaims, claims)
		if claims.DepartmentID != nil {
			c.Set(CtxKeyDepartmentID, *claims.DepartmentID)
		}

		c.Next()
	}
}

// RequireRole 角色权限校验中间件
// 只允许指定角色的用户访问
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(CtxKeyRole)
		if !exists {
			response.Error(c, errcode.ErrForbidden)
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		response.Error(c, errcode.ErrForbidden)
		c.Abort()
	}
}

// RequireAdmin 仅超级管理员可访问
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("super_admin")
}

// RequireManager 管理员或部门经理可访问
func RequireManager() gin.HandlerFunc {
	return RequireRole("super_admin", "dept_manager")
}

// GetUserID 从上下文获取当前用户 ID
func GetUserID(c *gin.Context) int64 {
	id, _ := c.Get(CtxKeyUserID)
	if id == nil {
		return 0
	}
	return id.(int64)
}

// GetUserRole 从上下文获取当前用户角色
func GetUserRole(c *gin.Context) string {
	role, _ := c.Get(CtxKeyRole)
	if role == nil {
		return ""
	}
	return role.(string)
}

// GetDepartmentID 从上下文获取当前用户部门 ID
func GetDepartmentID(c *gin.Context) *int64 {
	deptID, exists := c.Get(CtxKeyDepartmentID)
	if !exists {
		return nil
	}
	id := deptID.(int64)
	return &id
}

// GetClaims 从上下文获取 JWT Claims
func GetClaims(c *gin.Context) *jwtPkg.Claims {
	claims, _ := c.Get(CtxKeyClaims)
	if claims == nil {
		return nil
	}
	return claims.(*jwtPkg.Claims)
}
