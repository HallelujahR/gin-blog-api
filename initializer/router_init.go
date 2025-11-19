package initializer

import (
	"api/middleware"
	"api/routes"
	adminRoutes "api/routes/admin"
	"api/service"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	// 初始化访问日志，失败时仅记录警告信息
	if writer, path, err := service.InitAccessLog(); err != nil {
		fmt.Fprintf(gin.DefaultErrorWriter, "警告: 访问日志初始化失败(%s): %v\n", path, err)
	} else if writer != nil {
		gin.DefaultWriter = io.MultiWriter(gin.DefaultWriter, writer)
	}

	r := gin.Default()
	// 跨域中间件：允许全部开发请求，支持Authorization头
	r.Use(middleware.CORSMiddleware())
	// 初始化上传目录
	if err := service.InitUploadDirs(); err != nil {
		// 如果初始化失败，记录错误但不中断启动
		fmt.Fprintf(gin.DefaultErrorWriter, "警告: 上传目录初始化失败: %v\n", err)
	}

	// 静态文件服务：提供上传文件的公开访问
	r.Static("/uploads", "./uploads")

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
	routes.RegisterStatsRoutes(r)

	// ========== 后台管理API（需要管理员权限）==========
	adminRoutes.RegisterAdminUserRoutes(r)
	adminRoutes.RegisterAdminPostRoutes(r)
	adminRoutes.RegisterAdminCategoryRoutes(r)
	adminRoutes.RegisterAdminTagRoutes(r)
	adminRoutes.RegisterAdminCommentRoutes(r)
	adminRoutes.RegisterAdminPageRoutes(r)   // 页面管理接口
	adminRoutes.RegisterAdminUploadRoutes(r) // 文件上传接口

	return r
}
