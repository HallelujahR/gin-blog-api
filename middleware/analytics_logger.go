package middleware

import (
	"api/analytics"
	"api/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsLogger 将请求时序异步写入日志文件与 Redis，构建“请求 → 中间件 → 原始日志”的链路。
func AnalyticsLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		event := analytics.Event{
			Timestamp: start,
			IP:        utils.GetClientIP(c),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    c.Writer.Status(),
			Latency:   time.Since(start),
			UserAgent: c.Request.UserAgent(),
		}
		analytics.Capture(event)
	}
}
