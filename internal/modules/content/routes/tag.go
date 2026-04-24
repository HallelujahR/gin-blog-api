package routes

import (
	"api/internal/middleware"
	"api/internal/modules/content/controllers"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterTagRoutes(r *gin.Engine) {
	tag := r.Group("/api/tags")
	{
		tag.POST("", controllers.CreateTag)
		tag.GET(":id", controllers.GetTag)
		tag.GET("", middleware.RateLimitMiddleware(120, time.Minute), controllers.ListTags)
		tag.PUT(":id", controllers.UpdateTag)
		tag.DELETE(":id", controllers.DeleteTag)
	}
}
