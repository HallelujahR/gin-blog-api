package routes

import (
	"api/internal/middleware"
	"api/internal/modules/analytics/stats"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterStatsRoutes 注册统计信息相关的公开接口。
func RegisterStatsRoutes(r *gin.Engine) {
	r.GET("/api/stats", middleware.RateLimitMiddleware(60, time.Minute), stats.Handler)
}
