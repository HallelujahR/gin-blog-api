package admin

import (
	"api/models"
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 创建文章（管理后台）
func CreatePost(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req struct {
		Title       string   `json:"title" binding:"required"`
		Slug        string   `json:"slug" binding:"required"`
		Content     string   `json:"content" binding:"required"`
		CategoryIDs []uint64 `json:"categories"`
		TagIDs      []uint64 `json:"tags"`
		AuthorID    uint64   `json:"author_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 如果没有指定author_id，使用当前登录用户的ID
	if req.AuthorID == 0 {
		req.AuthorID = userID.(uint64)
	}
	
	post := &models.Post{
		Title:    req.Title,
		Slug:     req.Slug,
		Content:  req.Content,
		AuthorID: req.AuthorID,
		Status:   "published",
	}
	err := service.CreatePost(post, req.CategoryIDs, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"post": post})
}

// 删除文章（管理后台）
func DeletePost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 更新文章（管理后台）
func UpdatePost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	post, err := service.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	
	var req struct {
		Title   string   `json:"title"`
		Slug    string   `json:"slug"`
		Content string   `json:"content"`
		CategoryIDs []uint64 `json:"categories"`
		TagIDs      []uint64 `json:"tags"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Slug != "" {
		post.Slug = req.Slug
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	
	if err = service.UpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"post": post})
}
