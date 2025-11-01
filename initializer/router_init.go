package initializer

import (
	"api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	// 跨域中间件：允许全部开发请求
	r.Use(cors.Default())
	// 集中注册所有路由
	routes.RegisterUserRoutes(r)
	routes.RegisterPostRoutes(r)
	routes.RegisterCommentRoutes(r)
	routes.RegisterCategoryRoutes(r)
	routes.RegisterTagRoutes(r)
	routes.RegisterLikeRoutes(r)
	routes.RegisterPageRoutes(r)
	routes.RegisterHotDataRoutes(r)
	return r
}
