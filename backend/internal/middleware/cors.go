package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 跨域资源共享中间件
// allowedOrigins 为空时默认允许所有域名（开发环境），但会关闭 AllowCredentials
func CORS(allowedOrigins []string) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		MaxAge: 12 * time.Hour,
	}

	if len(allowedOrigins) > 0 {
		config.AllowOrigins = allowedOrigins
		config.AllowCredentials = true
	} else {
		// 未配置白名单时允许所有域名，但禁止携带凭证（符合 CORS 规范）
		config.AllowAllOrigins = true
		config.AllowCredentials = false
	}

	return cors.New(config)
}
