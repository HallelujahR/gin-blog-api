package controllers

import (
	"net/http"
	"strconv"

	"api/models"
	"api/service"

	"github.com/gin-gonic/gin"
)

// 创建文章
func CreatePost(c *gin.Context) {
	var req struct {
		Title       string   `json:"title" binding:"required"`
		Slug        string   `json:"slug" binding:"required"`
		Content     string   `json:"content" binding:"required"`
		CategoryIDs []uint64 `json:"categories"`
		TagIDs      []uint64 `json:"tags"`
		AuthorID    uint64   `json:"author_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post := models.Post{
		Title:    req.Title,
		Slug:     req.Slug,
		Content:  req.Content,
		AuthorID: req.AuthorID,
	}
	if err := service.CreatePost(&post, req.CategoryIDs, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"post": post})
}

// 获取文章详情
func GetPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	post, err := service.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"post": post})
}

// 文章列表
func ListPosts(c *gin.Context) {
	//接收参数
	var req struct {
		Page     int    `form:"page" binding:"required"`
		Size     int    `form:"size" binding:"required"`
		Q        string `form:"q"`
		Sort     string `form:"sort"`
		Category string `form:"category"`
		Tag      string `form:"tag"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	posts, err := service.ListPostsWithParams(req.Page, req.Size, req.Q, req.Sort, req.Category, req.Tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// 修改
func UpdatePost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	post, err := service.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	var req struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Content string `json:"content"`
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

// 删除
func DeletePost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
