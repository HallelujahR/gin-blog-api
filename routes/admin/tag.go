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
			tags.POST("", adminCtrl.CreateTag)
			tags.PUT("/:id", adminCtrl.UpdateTag)
			tags.DELETE("/:id", adminCtrl.DeleteTag)
		}
	}
}
