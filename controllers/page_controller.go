package controllers

import (
	"api/models"
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreatePage(c *gin.Context) {
	var req struct {
		Title    string  `json:"title" binding:"required"`
		Slug     string  `json:"slug" binding:"required"`
		Content  string  `json:"content" binding:"required"`
		Excerpt  string  `json:"excerpt"`
		ParentID *uint64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	page := models.Page{
		Title: req.Title, Slug: req.Slug, Content: req.Content, Excerpt: req.Excerpt, ParentID: req.ParentID,
	}
	if err := service.CreatePage(&page); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"page": page})
}
func GetPage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	page, err := service.GetPageByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"page": page})
}
func ListPages(c *gin.Context) {
	pages, err := service.ListPages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pages": pages})
}
func UpdatePage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	page, err := service.GetPageByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	var req struct {
		Title, Slug, Content, Excerpt string
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Title != "" {
		page.Title = req.Title
	}
	if req.Slug != "" {
		page.Slug = req.Slug
	}
	if req.Content != "" {
		page.Content = req.Content
	}
	if req.Excerpt != "" {
		page.Excerpt = req.Excerpt
	}
	if err = service.UpdatePage(page); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"page": page})
}
func DeletePage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePage(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
