package routes

import (
	"api/internal/middleware"
	"api/internal/modules/content/controllers"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterCategoryRoutes(r *gin.Engine) {
	cat := r.Group("/api/categories")
	{
		cat.POST("", controllers.CreateCategory)
		cat.GET(":id", middleware.RateLimitMiddleware(120, time.Minute), controllers.GetCategory)
		cat.GET("", middleware.RateLimitMiddleware(120, time.Minute), controllers.ListCategories)
		cat.GET(":id/full", middleware.RateLimitMiddleware(60, time.Minute), controllers.GetCategoryFull)
		cat.PUT(":id", controllers.UpdateCategory)
		cat.DELETE(":id", controllers.DeleteCategory)
	}
}
