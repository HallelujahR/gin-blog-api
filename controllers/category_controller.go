package controllers

import (
	"api/models"
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 创建分类
func CreateCategory(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Slug        string  `json:"slug" binding:"required"`
		Description string  `json:"description"`
		ParentID    *uint64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat := models.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    req.ParentID,
	}
	if err := service.CreateCategory(&cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": cat})
}

// 查详情（精简）
func GetCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cat, err := service.GetCategoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": cat})
}

// 查详情（含子、文章全结构）
func GetCategoryFull(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cat, err := service.GetCategoryByIDFull(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": cat})
}

// 列表
func ListCategories(c *gin.Context) {
	cats, err := service.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

// 修改
func UpdateCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cat, err := service.GetCategoryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Slug != "" {
		cat.Slug = req.Slug
	}
	if req.Description != "" {
		cat.Description = req.Description
	}
	if err = service.UpdateCategory(cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": cat})
}

// 删除
func DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
