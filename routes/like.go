package routes

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterLikeRoutes(r *gin.Engine) {
	lk := r.Group("/api/like")
	{
		lk.POST("/toggle", controllers.ToggleLike)
		lk.GET("/count", controllers.CountLikes)
	}
}
