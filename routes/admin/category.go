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
			categories.POST("", adminCtrl.CreateCategory)
			categories.PUT("/:id", adminCtrl.UpdateCategory)
			categories.DELETE("/:id", adminCtrl.DeleteCategory)
		}
	}
}
