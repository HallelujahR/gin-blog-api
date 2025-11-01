package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminPostRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		posts := adminGroup.Group("/posts")
		{
			posts.POST("", adminCtrl.CreatePost)
			posts.PUT("/:id", adminCtrl.UpdatePost)
			posts.DELETE("/:id", adminCtrl.DeletePost)
		}
	}
}
