package routes

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterHotDataRoutes(r *gin.Engine) {
	hd := r.Group("/api/hotdata")
	{
		hd.POST("", controllers.CreateHotData)
		hd.GET("", controllers.ListHotData)
		hd.DELETE(":id", controllers.DeleteHotData)
	}
}
