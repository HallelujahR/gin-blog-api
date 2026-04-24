package routes

import (
	"api/internal/middleware"
	"api/internal/modules/content/controllers"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterMomentRoutes(r *gin.Engine) {
	moments := r.Group("/api/moments")
	{
		moments.GET("", middleware.RateLimitMiddleware(120, time.Minute), controllers.ListMoments)
	}
}
