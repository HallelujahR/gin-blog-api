package initializer

import (
	"api/middleware"
	"api/routes"
	adminRoutes "api/routes/admin"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	// 跨域中间件：允许全部开发请求，支持Authorization头
	r.Use(middleware.CORSMiddleware())

	// ========== 前端用户访问API（公开或认证用户访问）==========
	// 保持原有接口路径不变，确保向前兼容
	routes.RegisterUserRoutes(r)
	routes.RegisterPostRoutes(r)
	routes.RegisterCommentRoutes(r)
	routes.RegisterCategoryRoutes(r)
	routes.RegisterTagRoutes(r)
	routes.RegisterLikeRoutes(r)
	routes.RegisterPageRoutes(r)
	routes.RegisterHotDataRoutes(r)

	// ========== 后台管理API（需要管理员权限）==========
	adminRoutes.RegisterAdminUserRoutes(r)
	adminRoutes.RegisterAdminPostRoutes(r)
	adminRoutes.RegisterAdminCategoryRoutes(r)
	adminRoutes.RegisterAdminTagRoutes(r)
	adminRoutes.RegisterAdminCommentRoutes(r)

	return r
}
