package admin

import (
	"api/internal/middleware"
	adminCtrl "api/internal/modules/content/controllers/admin"

	"github.com/gin-gonic/gin"
)

func RegisterAdminPostRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		posts := adminGroup.Group("/posts")
		{
			posts.GET("", adminCtrl.ListPosts)                             // 文章列表（支持分页和筛选）
			posts.POST("/suggest-taxonomy", adminCtrl.SuggestPostTaxonomy) // 根据内容推荐分类和标签
			posts.POST("", adminCtrl.CreatePost)                           // 创建文章
			posts.GET("/:id", adminCtrl.GetPost)                           // 获取文章详情
			posts.PUT("/:id", adminCtrl.UpdatePost)                        // 更新文章
			posts.DELETE("/:id", adminCtrl.DeletePost)                     // 删除文章
		}
	}
}
