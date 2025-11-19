package routes

import (
	"api/controllers"
	"api/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterCommentRoutes(r *gin.Engine) {
	cmt := r.Group("/api/comments")
	{
		cmt.POST("", middleware.RateLimitMiddleware(120, time.Minute), controllers.CreateComment)
		cmt.GET(":id", controllers.GetComment)
		cmt.GET("", controllers.ListCommentsByPost)
		cmt.PUT(":id", controllers.UpdateComment)
		cmt.DELETE(":id", controllers.DeleteComment)
	}
}
