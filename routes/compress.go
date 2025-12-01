package routes

import (
	adminCtrl "api/controllers/admin"

	"github.com/gin-gonic/gin"
)

// RegisterCompressRoutes 注册公共图片压缩接口（无需登录，任何人可用）。
func RegisterCompressRoutes(r *gin.Engine) {
	// 公开工具类接口前缀，可按需调整，例如 /api/tools
	tool := r.Group("/api/tools")
	{
		// 图片压缩异步任务：任何人可调用
		tool.POST("/image-compress/start", adminCtrl.StartCompressJob)
		tool.GET("/image-compress/stream", adminCtrl.StreamCompressProgress)
		tool.GET("/image-compress/stats", adminCtrl.GetCompressStats)
	}
}


