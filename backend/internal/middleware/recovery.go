package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery Panic 恢复中间件
// 捕获 panic 并返回 500 错误，同时记录堆栈信息
// 自动适配 LLM 代理端点的协议格式（需配合 SetLLMProtocol 使用）
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 详情和堆栈
				logger.Error("服务器 Panic 恢复",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.ByteString("stack", debug.Stack()),
				)

				sendProtocolError(c, http.StatusInternalServerError, "服务器内部错误")
				c.Abort()
			}
		}()

		c.Next()
	}
}
