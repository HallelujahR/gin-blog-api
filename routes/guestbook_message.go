package routes

import (
	"api/controllers"
	"api/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterGuestbookRoutes(r *gin.Engine) {
	guestbook := r.Group("/api/guestbook")
	{
		guestbook.GET("", middleware.RateLimitMiddleware(120, time.Minute), controllers.ListApprovedGuestbookMessages)
		guestbook.POST("", middleware.RateLimitMiddleware(10, time.Minute), controllers.CreateGuestbookMessage)
	}
}
