package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery handles panics and returns 500 errors with stack traces logged.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.ByteString("stack", debug.Stack()),
				)

				sendProtocolError(c, http.StatusInternalServerError, "internal server error")
				c.Abort()
			}
		}()

		c.Next()
	}
}
