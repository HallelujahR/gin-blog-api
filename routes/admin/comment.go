package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminCommentRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		comments := adminGroup.Group("/comments")
		{
			comments.GET("", adminCtrl.ListComments)                // 评论列表（分页、筛选）
			comments.GET("/:id", adminCtrl.GetComment)             // 获取评论详情（含回复）
			comments.DELETE("/:id", adminCtrl.DeleteComment)       // 删除单个评论
			comments.POST("/batch-delete", adminCtrl.DeleteComments) // 批量删除
			comments.PUT("/:id/status", adminCtrl.UpdateCommentStatus) // 更新单个评论状态
			comments.POST("/batch-status", adminCtrl.UpdateCommentsStatus) // 批量更新状态
			comments.POST("/:id/reply", adminCtrl.ReplyComment)    // 回复评论
		}
	}
}
