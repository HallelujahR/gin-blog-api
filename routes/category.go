package routes

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterCategoryRoutes(r *gin.Engine) {
	cat := r.Group("/api/categories")
	{
		cat.POST("", controllers.CreateCategory)
		cat.GET(":id", controllers.GetCategory)
		cat.GET("", controllers.ListCategories)
		cat.GET(":id/full", controllers.GetCategoryFull)
		cat.PUT(":id", controllers.UpdateCategory)
		cat.DELETE(":id", controllers.DeleteCategory)
	}
}
