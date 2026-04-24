package routes

import (
	"api/internal/middleware"
	"api/internal/modules/content/controllers"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterCommentRoutes(r *gin.Engine) {
	cmt := r.Group("/api/comments")
	{
		cmt.POST("", middleware.RateLimitMiddleware(20, time.Minute), controllers.CreateComment)
		cmt.GET(":id", middleware.RateLimitMiddleware(120, time.Minute), controllers.GetComment)
		cmt.GET("", middleware.RateLimitMiddleware(120, time.Minute), controllers.ListCommentsByPost)
		cmt.PUT(":id", controllers.UpdateComment)
		cmt.DELETE(":id", controllers.DeleteComment)
	}
}
