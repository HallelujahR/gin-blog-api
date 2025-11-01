package routes

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(r *gin.Engine) {
	post := r.Group("/api/posts")
	{
		post.POST("", controllers.CreatePost)
		post.GET(":id", controllers.GetPost)
		post.GET("", controllers.ListPosts)
		post.PUT(":id", controllers.UpdatePost)
		post.DELETE(":id", controllers.DeletePost)
	}
}
