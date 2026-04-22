package controllers

import (
	"api/service"
	"api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 点赞或取消点赞（基于IP，不依赖用户ID）
func ToggleLike(c *gin.Context) {
	var req struct {
		PostID    *uint64 `json:"post_id"`
		CommentID *uint64 `json:"comment_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.PostID == nil && req.CommentID == nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if req.PostID != nil && *req.PostID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id 无效"})
		return
	}
	if req.CommentID != nil && *req.CommentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_id 无效"})
		return
	}

	// 获取客户端IP
	ip := utils.GetClientIP(c)

	msg, err := service.ToggleLike(ip, req.PostID, req.CommentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

// 统计点赞数
func CountLikes(c *gin.Context) {
	var postID, commentID *uint64
	if v := c.Query("post_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		if id > 0 {
			postID = &id
		}
	}
	if v := c.Query("comment_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		if id > 0 {
			commentID = &id
		}
	}
	if postID == nil && commentID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	n, err := service.CountLikes(postID, commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": n})
}
