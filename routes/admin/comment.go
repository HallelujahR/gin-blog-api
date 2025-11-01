package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminCommentRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		comments := adminGroup.Group("/comments")
		{
			comments.GET("", adminCtrl.ListComments)
			comments.DELETE("/:id", adminCtrl.DeleteComment)
			comments.PUT("/:id/status", adminCtrl.UpdateCommentStatus)
		}
	}
}
