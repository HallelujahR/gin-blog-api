package admin

import (
	"api/internal/middleware"
	adminCtrl "api/internal/modules/content/controllers/admin"

	"github.com/gin-gonic/gin"
)

func RegisterAdminGuestbookRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		guestbook := adminGroup.Group("/guestbook")
		{
			guestbook.GET("", adminCtrl.ListGuestbookMessages)
			guestbook.PUT("/:id/status", adminCtrl.UpdateGuestbookMessageStatus)
			guestbook.DELETE("/:id", adminCtrl.DeleteGuestbookMessage)
		}
	}
}
