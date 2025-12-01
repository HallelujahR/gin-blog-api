package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 配置CORS跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 允许来源：
		// - 生产环境: riverlog.cn
		// - 本地后端: localhost:8080
		// - 本地前端开发: localhost:5173
		AllowAllOrigins: false,
		AllowOrigins:    []string{"http://riverlog.cn", "http://localhost:8080", "http://localhost:5173"},

		// 允许的HTTP方法
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},

		// 允许的请求头，重要：包含Authorization以便携带token
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},

		// 暴露的响应头
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
		},

		// 允许携带凭证（如cookies）
		AllowCredentials: true,

		// 预检请求缓存时间
		MaxAge: 12 * time.Hour,

		// 允许私有网络
		AllowPrivateNetwork: true,
	})
}
