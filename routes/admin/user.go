package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminUserRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		users := adminGroup.Group("/users")
		{
			users.GET("", adminCtrl.ListUsers)
			users.DELETE("/:id", adminCtrl.DeleteUser)
			users.PUT("/:id/status", adminCtrl.UpdateUserStatus)
			users.PUT("/:id/role", adminCtrl.UpdateUserRole)
			users.PUT("/:id/password", adminCtrl.ChangeUserPassword)
		}
	}
}
