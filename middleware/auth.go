package middleware

import (
	"api/dao"
	"api/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 从Header中获取并验证token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息"})
			c.Abort()
			return
		}

		// 检查Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证token
		session, err := dao.GetSessionByToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 检查是否过期
		if time.Now().After(session.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证已过期"})
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := dao.GetUserByID(session.UserID)
		if err != nil || user.Status != "active" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被禁用"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("user", user)
		c.Next()
	}
}

// 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 作者或管理员权限中间件
func AuthorOrAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || (role != "admin" && role != "author") {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要作者或管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 可选认证中间件（登录用户可用，未登录也能访问）
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		session, err := dao.GetSessionByToken(token)
		if err != nil {
			c.Next()
			return
		}

		if time.Now().After(session.ExpiresAt) {
			c.Next()
			return
		}

		user, err := dao.GetUserByID(session.UserID)
		if err == nil && user.Status == "active" {
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
			c.Set("user", user)
		}
		c.Next()
	}
}

// 创建会话（用于登录后生成token）
func CreateSession(userID uint64, userAgent, ipAddress string) (*models.UserSession, error) {
	session := &models.UserSession{
		UserID:       userID,
		SessionToken: uuid.New().String() + "_" + uuid.New().String(),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7天有效期
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
	}
	return dao.CreateSession(session)
}

// 登出（删除会话）
func DeleteSession(token string) error {
	return dao.DeleteSession(token)
}
