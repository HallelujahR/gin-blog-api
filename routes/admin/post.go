package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminPostRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		posts := adminGroup.Group("/posts")
		{
			posts.GET("", adminCtrl.ListPosts)           // 文章列表（支持分页和筛选）
			posts.POST("", adminCtrl.CreatePost)         // 创建文章
			posts.GET("/:id", adminCtrl.GetPost)        // 获取文章详情
			posts.PUT("/:id", adminCtrl.UpdatePost)      // 更新文章
			posts.DELETE("/:id", adminCtrl.DeletePost)   // 删除文章
		}
	}
}
