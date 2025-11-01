package routes

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterPageRoutes(r *gin.Engine) {
	pg := r.Group("/api/pages")
	{
		pg.POST("", controllers.CreatePage)
		pg.GET(":id", controllers.GetPage)
		pg.GET("", controllers.ListPages)
		pg.PUT(":id", controllers.UpdatePage)
		pg.DELETE(":id", controllers.DeletePage)
	}
}
