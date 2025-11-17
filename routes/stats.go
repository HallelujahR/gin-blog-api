package routes

import (
	"api/stats"

	"github.com/gin-gonic/gin"
)

// RegisterStatsRoutes 注册统计信息相关的公开接口。
func RegisterStatsRoutes(r *gin.Engine) {
	r.GET("/api/stats", stats.Handler)
}
