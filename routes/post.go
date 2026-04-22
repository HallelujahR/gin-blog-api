package routes

import (
	"api/controllers"
	"api/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(r *gin.Engine) {
	post := r.Group("/api/posts")
	{
		post.POST("", controllers.CreatePost)
		post.GET(":id", middleware.RateLimitMiddleware(300, time.Minute), controllers.GetPost)
		post.GET("", middleware.RateLimitMiddleware(180, time.Minute), controllers.ListPosts)
		post.PUT(":id", controllers.UpdatePost)
		post.DELETE(":id", controllers.DeletePost)
	}
}
