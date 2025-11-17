package middleware

import (
	"net/http"
	"sync"
	time "time"

	"github.com/gin-gonic/gin"
)

// limiterEntry 保存单个IP在当前时间窗口内的请求计数信息。
// windowStart 记录窗口起始时间，count 记录已收到的请求数量。
type limiterEntry struct {
	count       int
	windowStart time.Time
}

// RateLimitMiddleware 返回一个基于IP的简单限流中间件。
// limit 表示单位窗口允许的最大请求数量，window 为时间窗口长度。
// 例如：limit=120、window=1分钟，代表同一IP在一分钟内最多允许120次请求。
// 超出限制后会返回 HTTP 429 Too Many Requests。
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	var (
		mu      sync.Mutex
		buckets = make(map[string]*limiterEntry)
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		entry, exists := buckets[ip]
		if !exists {
			// 新IP直接创建计数器，允许通过
			buckets[ip] = &limiterEntry{count: 1, windowStart: now}
			mu.Unlock()
			c.Next()
			return
		}

		// 如果已超出窗口，则重置计数器
		if now.Sub(entry.windowStart) > window {
			entry.windowStart = now
			entry.count = 1
			mu.Unlock()
			c.Next()
			return
		}

		entry.count++
		if entry.count > limit {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		mu.Unlock()

		c.Next()
	}
}
