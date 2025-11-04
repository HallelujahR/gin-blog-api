package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminTagRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		tags := adminGroup.Group("/tags")
		{
			tags.GET("", adminCtrl.ListAllTags)          // 获取所有标签（用于文章编辑）
			tags.POST("", adminCtrl.CreateTag)           // 创建标签
			tags.PUT("/:id", adminCtrl.UpdateTag)        // 更新标签
			tags.DELETE("/:id", adminCtrl.DeleteTag)      // 删除标签
		}
	}
}
