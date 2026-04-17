package middleware

import (
	"codemind/internal/pkg/errcode"
	"codemind/internal/pkg/response"
	"strings"

	jwtPkg "codemind/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	CtxKeyUserID       = "user_id"
	CtxKeyUsername     = "username"
	CtxKeyRole         = "role"
	CtxKeyDepartmentID = "department_id"
	CtxKeyClaims       = "claims"
)

// JWTAuth validates JWT tokens from Authorization header.
func JWTAuth(jwtManager *jwtPkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2) //nolint:mnd // intentional constant.
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

		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

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

// RequireRole restricts access to specified roles.
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

// RequireAdmin restricts access to super admins.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("super_admin")
}

// RequireManager restricts access to admins and dept managers.
func RequireManager() gin.HandlerFunc {
	return RequireRole("super_admin", "dept_manager")
}

// GetUserID returns the current user ID from context.
func GetUserID(c *gin.Context) int64 {
	id, _ := c.Get(CtxKeyUserID)
	if id == nil {
		return 0
	}
	return id.(int64)
}

// GetUserRole returns the current user role from context.
func GetUserRole(c *gin.Context) string {
	role, _ := c.Get(CtxKeyRole)
	if role == nil {
		return ""
	}
	return role.(string)
}

// GetDepartmentID returns the current user department ID from context.
func GetDepartmentID(c *gin.Context) *int64 {
	deptID, exists := c.Get(CtxKeyDepartmentID)
	if !exists {
		return nil
	}
	id := deptID.(int64)
	return &id
}

// GetClaims returns the JWT claims from context.
func GetClaims(c *gin.Context) *jwtPkg.Claims {
	claims, _ := c.Get(CtxKeyClaims)
	if claims == nil {
		return nil
	}
	return claims.(*jwtPkg.Claims)
}
