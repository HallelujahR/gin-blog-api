package controllers

import (
	"net/http"
	"strconv"

	"api/models"
	"api/service"
	"api/utils"

	"github.com/gin-gonic/gin"
)

// 创建评论
func CreateComment(c *gin.Context) {
	var req struct {
		Content      string  `json:"content" binding:"required"`
		AuthorName   string  `json:"author_name" binding:"required"`
		AuthorEmail  string  `json:"author_email" binding:"required"`
		AuthorUserID *uint64 `json:"author_user_id"`
		PostID       uint64  `json:"post_id" binding:"required"`
		ParentID     *uint64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment := models.Comment{
		Content:      req.Content,
		AuthorName:   req.AuthorName,
		AuthorEmail:  req.AuthorEmail,
		AuthorUserID: req.AuthorUserID,
		PostID:       req.PostID,
		ParentID:     req.ParentID,
		//获取评论请求来自的IP
		AuthorIP: utils.GetClientIP(c),
	}
	if err := service.CreateComment(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// 详情
func GetComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	comment, err := service.GetCommentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// 文章评论列表
func ListCommentsByPost(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Query("post_id"), 10, 64)
	comments, err := service.ListCommentsByPost(pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// 修改
func UpdateComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	comment, err := service.GetCommentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Content != "" {
		comment.Content = req.Content
	}
	if err = service.UpdateComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// 删除
func DeleteComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteComment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
