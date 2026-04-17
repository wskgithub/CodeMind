package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// MonitorStats defines monitor statistics interface.
type MonitorStats interface {
	RecordRequestMetrics(statusCode int, responseTimeMs float64)
}

// RequestMonitor records request response time and status code for QPS and response time statistics.
func RequestMonitor(stats MonitorStats) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		responseTimeMs := float64(duration.Microseconds()) / 1000.0

		statusCode := c.Writer.Status()

		go stats.RecordRequestMetrics(statusCode, responseTimeMs)
	}
}
