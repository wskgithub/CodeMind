// Package middleware 监控相关中间件
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// MonitorStats 监控统计接口
type MonitorStats interface {
	RecordRequestMetrics(statusCode int, responseTimeMs float64)
}

// RequestMonitor 请求性能监控中间件
// 记录每个请求的响应时间和状态码，用于计算 QPS 和响应时间统计
func RequestMonitor(stats MonitorStats) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 计算响应时间
		duration := time.Since(start)
		responseTimeMs := float64(duration.Microseconds()) / 1000.0
		
		// 获取状态码
		statusCode := c.Writer.Status()
		
		// 记录指标（异步，不阻塞响应）
		go stats.RecordRequestMetrics(statusCode, responseTimeMs)
	}
}
