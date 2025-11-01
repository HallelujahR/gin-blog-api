package routes

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine) {
	user := r.Group("/api/users")
	{
		user.POST("", controllers.Register)
		user.POST("/login", controllers.Login)
		user.GET(":id", controllers.UserDetail)
		user.PUT(":id", controllers.UpdateUser)
		user.DELETE(":id", controllers.DeleteUser)
	}
}
