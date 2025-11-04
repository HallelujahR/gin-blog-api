package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminPageRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		pages := adminGroup.Group("/pages")
		{
			pages.GET("", adminCtrl.ListPages)           // 页面列表
			pages.POST("", adminCtrl.CreatePage)         // 创建页面
			pages.GET("/:id", adminCtrl.GetPage)        // 获取页面详情
			pages.PUT("/:id", adminCtrl.UpdatePage)     // 更新页面
			pages.DELETE("/:id", adminCtrl.DeletePage)  // 删除页面
		}
	}
}

