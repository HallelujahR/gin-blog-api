package routes

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterTagRoutes(r *gin.Engine) {
	tag := r.Group("/api/tags")
	{
		tag.POST("", controllers.CreateTag)
		tag.GET(":id", controllers.GetTag)
		tag.GET("", controllers.ListTags)
		tag.PUT(":id", controllers.UpdateTag)
		tag.DELETE(":id", controllers.DeleteTag)
	}
}
