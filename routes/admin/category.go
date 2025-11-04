package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminCategoryRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		categories := adminGroup.Group("/categories")
		{
			categories.GET("", adminCtrl.ListAllCategories)   // 获取所有分类（用于文章编辑）
			categories.POST("", adminCtrl.CreateCategory)     // 创建分类
			categories.PUT("/:id", adminCtrl.UpdateCategory)  // 更新分类
			categories.DELETE("/:id", adminCtrl.DeleteCategory) // 删除分类
		}
	}
}
