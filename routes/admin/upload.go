package admin

import (
	adminCtrl "api/controllers/admin"
	"api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAdminUploadRoutes(r *gin.Engine) {
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware()) // 需要管理员权限
	{
		upload := adminGroup.Group("/upload")
		{
			upload.POST("/file", adminCtrl.UploadFile)   // 上传单个文件
			upload.POST("/image", adminCtrl.UploadImage) // 上传图片
			upload.POST("/files", adminCtrl.UploadFiles) // 批量上传文件
			upload.DELETE("/file", adminCtrl.DeleteFile) // 删除文件
			upload.GET("/files", adminCtrl.ListFiles)    // 获取文件列表
		}
	}
}
