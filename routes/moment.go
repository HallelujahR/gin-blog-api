package routes

import (
	"api/controllers"
	"api/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterMomentRoutes(r *gin.Engine) {
	moments := r.Group("/api/moments")
	{
		moments.GET("", middleware.RateLimitMiddleware(120, time.Minute), controllers.ListMoments)
	}
}
