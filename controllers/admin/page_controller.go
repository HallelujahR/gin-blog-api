package admin

import (
	"api/models"
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 创建页面（管理后台）
func CreatePage(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Slug    string `json:"slug" binding:"required"`
		Content string `json:"content" binding:"required"`
		Excerpt string `json:"excerpt"`
		Status  string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误: " + err.Error()})
		return
	}

	// 设置默认值
	status := req.Status
	if status == "" {
		status = "draft"
	}

	page := &models.Page{
		Title:   req.Title,
		Slug:    req.Slug,
		Content: req.Content,
		Excerpt: req.Excerpt,
		Status:  status,
	}

	if err := service.CreatePage(page); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// 获取页面详情（管理后台）
func GetPage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	page, err := service.GetPageByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "页面不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// 更新页面（管理后台）
func UpdatePage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	page, err := service.GetPageByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "页面不存在"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Content string `json:"content"`
		Excerpt string `json:"excerpt"`
		Status  string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误: " + err.Error()})
		return
	}

	// 更新字段
	if req.Title != "" {
		page.Title = req.Title
	}
	if req.Slug != "" {
		page.Slug = req.Slug
	}
	if req.Content != "" {
		page.Content = req.Content
	}
	if req.Excerpt != "" || c.GetHeader("Content-Type") == "application/json" {
		page.Excerpt = req.Excerpt
	}
	if req.Status != "" {
		page.Status = req.Status
	}

	if err = service.UpdatePage(page); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// 删除页面（管理后台）
func DeletePage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePage(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 页面列表（管理后台）
func ListPages(c *gin.Context) {
	pages, err := service.ListPages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pages": pages})
}

