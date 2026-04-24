package routes

import (
	"api/internal/middleware"
	"api/internal/modules/content/controllers"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterHotDataRoutes(r *gin.Engine) {
	hd := r.Group("/api/hotdata")
	{
		hd.POST("", controllers.CreateHotData)
		hd.GET("", middleware.RateLimitMiddleware(60, time.Minute), controllers.ListHotData)
		hd.DELETE(":id", controllers.DeleteHotData)
	}
}
