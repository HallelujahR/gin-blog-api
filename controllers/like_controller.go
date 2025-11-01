package controllers

import (
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 点赞或取消点赞
func ToggleLike(c *gin.Context) {
	var req struct {
		UserID    uint64  `json:"user_id" binding:"required"`
		PostID    *uint64 `json:"post_id"`
		CommentID *uint64 `json:"comment_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.PostID == nil && req.CommentID == nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	msg, err := service.ToggleLike(req.UserID, req.PostID, req.CommentID)
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
		postID = &id
	}
	if v := c.Query("comment_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		commentID = &id
	}
	n, err := service.CountLikes(postID, commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": n})
}
