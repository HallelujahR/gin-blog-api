package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminMomentRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		moments := adminGroup.Group("/moments")
		{
			moments.GET("", adminCtrl.ListMoments)
			moments.POST("", adminCtrl.CreateMoment)
			moments.GET("/:id", adminCtrl.GetMoment)
			moments.PUT("/:id", adminCtrl.UpdateMoment)
			moments.DELETE("/:id", adminCtrl.DeleteMoment)
		}
	}
}
