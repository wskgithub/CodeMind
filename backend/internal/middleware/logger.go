package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 请求日志中间件
// 记录每个请求的方法、路径、状态码、耗时等信息
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		}

		// 添加用户信息（如果已认证）
		if userID, exists := c.Get(CtxKeyUserID); exists {
			fields = append(fields, zap.Int64("user_id", userID.(int64)))
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error("服务器错误", fields...)
		case statusCode >= 400:
			logger.Warn("客户端错误", fields...)
		default:
			logger.Info("请求完成", fields...)
		}
	}
}
