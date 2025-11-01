package controllers

import (
	"api/middleware"
	"api/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 注册
func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	user, err := service.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// 登录
func Login(c *gin.Context) {
	var req struct {
		Identifier string `json:"username"`
		Password   string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "格式错误"})
		return
	}
	user, err := service.ValidateLogin(req.Identifier, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	// 创建会话并返回token
	session, err := middleware.CreateSession(user.ID, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	service.UpdateUser(user)

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": session.SessionToken,
	})
}

// 查详情
func UserDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := service.UserDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// 更新
func UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := service.UserDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	var req struct {
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Bio         string `json:"bio"`
		Website     string `json:"website"`
		Location    string `json:"location"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.DisplayName = req.DisplayName
	user.AvatarURL = req.AvatarURL
	user.Bio = req.Bio
	user.Website = req.Website
	user.Location = req.Location
	if err = service.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// 删除
func DeleteUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
