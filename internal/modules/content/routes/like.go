package routes

import (
	"api/internal/middleware"
	"api/internal/modules/content/controllers"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterLikeRoutes(r *gin.Engine) {
	lk := r.Group("/api/like")
	{
		lk.POST("/toggle", middleware.RateLimitMiddleware(40, time.Minute), controllers.ToggleLike)
		lk.GET("/count", middleware.RateLimitMiddleware(120, time.Minute), controllers.CountLikes)
	}
}
