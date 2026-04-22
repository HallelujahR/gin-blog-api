package controllers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

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
	req.Content = strings.TrimSpace(req.Content)
	req.AuthorName = strings.TrimSpace(req.AuthorName)
	req.AuthorEmail = strings.TrimSpace(req.AuthorEmail)
	if req.PostID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id 无效"})
		return
	}
	if len([]rune(req.Content)) < 2 || len([]rune(req.Content)) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容长度需在 2 到 2000 个字符之间"})
		return
	}
	if len([]rune(req.AuthorName)) < 1 || len([]rune(req.AuthorName)) > 40 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "昵称长度需在 1 到 40 个字符之间"})
		return
	}
	if !isValidEmail(req.AuthorEmail) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
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

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func isValidEmail(email string) bool {
	if len(email) > 120 {
		return false
	}
	return emailPattern.MatchString(email)
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
