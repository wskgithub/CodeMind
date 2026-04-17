package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS configures cross-origin resource sharing
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
		config.AllowAllOrigins = true
		config.AllowCredentials = false
	}

	return cors.New(config)
}
