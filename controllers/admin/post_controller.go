package admin

import (
	"api/models"
	"api/service"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 创建文章（管理后台）
// 支持两种格式：
// 1. FormData（有image文件时）
// 2. JSON（无文件时，字段名：category_ids, tag_ids）
func CreatePost(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var title, content, excerpt, status string
	var categoryIDs, tagIDs []uint64
	var coverImageURL string

	contentType := c.GetHeader("Content-Type")
	isFormData := contentType != "" && (strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/x-www-form-urlencoded"))

	if isFormData {
		// FormData格式（有文件时）
		title = c.PostForm("title")
		content = c.PostForm("content")
		excerpt = c.PostForm("excerpt")
		status = c.PostForm("status")

		// 处理图片文件
		if file, err := c.FormFile("image"); err == nil {
			filePath, err := handleImageUpload(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "图片上传失败: " + err.Error()})
				return
			}
			// 生成完整URL
			coverImageURL = service.GetFullFileURL(filePath)
		}

		// 处理分类和标签ID（FormData数组格式：category_ids[]）
		categoryIDs = parseFormIDArray(c.PostFormArray("category_ids[]"))
		if len(categoryIDs) == 0 {
			categoryIDs = parseFormIDArray(c.PostFormArray("category_ids"))
		}
		tagIDs = parseFormIDArray(c.PostFormArray("tag_ids[]"))
		if len(tagIDs) == 0 {
			tagIDs = parseFormIDArray(c.PostFormArray("tag_ids"))
		}
	} else {
		// JSON格式（无文件时）
		var req struct {
			Title       string   `json:"title" binding:"required"`
			Content     string   `json:"content" binding:"required"`
			Excerpt     string   `json:"excerpt"`
			CoverImage  string   `json:"cover_image"`
			Status      string   `json:"status"`
			CategoryIDs []uint64 `json:"category_ids"` // 前端使用category_ids
			TagIDs      []uint64 `json:"tag_ids"`      // 前端使用tag_ids
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误: " + err.Error()})
			return
		}
		title = req.Title
		content = req.Content
		excerpt = req.Excerpt
		status = req.Status
		// 处理cover_image：转换为完整URL
		if req.CoverImage != "" {
			coverImageURL = service.GetFullFileURL(req.CoverImage)
		}
		categoryIDs = req.CategoryIDs
		tagIDs = req.TagIDs
	}

	// 验证必填字段
	if title == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题和内容不能为空"})
		return
	}

	// 生成slug
	slug := service.GenerateSlug(title)
	slug = service.EnsureUniqueSlug(slug, 0)

	// 设置默认值
	if status == "" {
		status = "draft"
	}

	// 如果状态是published，设置发布时间
	var publishedAt *time.Time
	if status == "published" {
		now := time.Now()
		publishedAt = &now
	}

	post := &models.Post{
		Title:       title,
		Slug:        slug,
		Content:     content,
		Excerpt:     excerpt,
		CoverImage:  coverImageURL,
		AuthorID:    userID.(uint64),
		Status:      status,
		Visibility:  "public",
		PublishedAt: publishedAt,
	}

	if err := service.CreatePost(post, categoryIDs, tagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	// 确保返回的cover_image是完整URL
	if post.CoverImage != "" {
		post.CoverImage = service.GetFullFileURL(post.CoverImage)
	}

	c.JSON(http.StatusOK, gin.H{"post": post})
}

// 处理图片上传的辅助函数
func handleImageUpload(file *multipart.FileHeader) (string, error) {
	// 验证文件类型
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".svg": true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return "", fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 生成唯一文件名
	filename := service.GenerateFileName(file.Filename)

	// 保存文件
	filePath, err := service.SaveUploadedFile(file, filename)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// 解析表单中的ID数组
func parseFormIDArray(values []string) []uint64 {
	var ids []uint64
	for _, val := range values {
		if id, err := strconv.ParseUint(val, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// 获取文章详情（管理后台）
func GetPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	post, categories, tags, err := service.GetPostWithFullRelations(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 确保返回空数组而不是nil
	if categories == nil {
		categories = []models.Category{}
	}
	if tags == nil {
		tags = []models.Tag{}
	}

	// 提取分类和标签ID数组
	categoryIDs := make([]uint64, len(categories))
	for i, cat := range categories {
		categoryIDs[i] = cat.ID
	}
	tagIDs := make([]uint64, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}

	// 将分类和标签数据添加到post对象中（供前端直接使用）
	if post.CategoryIDs == nil {
		post.CategoryIDs = categoryIDs
	}
	if post.TagIDs == nil {
		post.TagIDs = tagIDs
	}

	// 确保返回的cover_image是完整URL
	if post.CoverImage != "" {
		post.CoverImage = service.GetFullFileURL(post.CoverImage)
	}

	// 返回文章信息及完整的分类和标签信息
	response := gin.H{
		"post": post,
		// 顶层也返回完整的分类和标签信息（便于前端直接使用）
		"categories": categories,
		"tags":       tags,
		// 兼容性：也返回ID数组（顶层）
		"category_ids": categoryIDs,
		"tag_ids":      tagIDs,
	}
	c.JSON(http.StatusOK, response)
}

// 删除文章（管理后台）
func DeletePost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePost(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 更新文章（管理后台）
// 支持两种格式：
// 1. FormData（有image文件时）
// 2. JSON（无文件时，字段名：category_ids, tag_ids）
func UpdatePost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	post, err := service.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	var title, content, excerpt, status string
	var categoryIDs, tagIDs []uint64
	var coverImageURL string

	contentType := c.GetHeader("Content-Type")
	isFormData := contentType != "" && (strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/x-www-form-urlencoded"))

	if isFormData {
		// FormData格式（有文件时）
		title = c.PostForm("title")
		content = c.PostForm("content")
		excerpt = c.PostForm("excerpt")
		status = c.PostForm("status")
		coverImageURL = c.PostForm("cover_image")

		// 处理图片文件
		if file, err := c.FormFile("image"); err == nil {
			filePath, err := handleImageUpload(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "图片上传失败: " + err.Error()})
				return
			}
			// 生成完整URL
			coverImageURL = service.GetFullFileURL(filePath)
		}

		// 处理分类和标签ID（FormData数组格式：category_ids[]）
		categoryIDs = parseFormIDArray(c.PostFormArray("category_ids[]"))
		if len(categoryIDs) == 0 {
			categoryIDs = parseFormIDArray(c.PostFormArray("category_ids"))
		}
		tagIDs = parseFormIDArray(c.PostFormArray("tag_ids[]"))
		if len(tagIDs) == 0 {
			tagIDs = parseFormIDArray(c.PostFormArray("tag_ids"))
		}
	} else {
		// JSON格式（无文件时）
		var req struct {
			Title       string   `json:"title"`
			Content     string   `json:"content"`
			Excerpt     string   `json:"excerpt"`
			CoverImage  string   `json:"cover_image"`
			Status      string   `json:"status"`
			CategoryIDs []uint64 `json:"category_ids"` // 前端使用category_ids
			TagIDs      []uint64 `json:"tag_ids"`      // 前端使用tag_ids
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误: " + err.Error()})
			return
		}
		title = req.Title
		content = req.Content
		excerpt = req.Excerpt
		status = req.Status
		// 处理cover_image：转换为完整URL
		if req.CoverImage != "" {
			coverImageURL = service.GetFullFileURL(req.CoverImage)
		}
		categoryIDs = req.CategoryIDs
		tagIDs = req.TagIDs
	}

	hasUpdates := false

	// 更新标题
	if title != "" {
		post.Title = title
		hasUpdates = true
		// 如果标题改变，自动更新slug
		newSlug := service.GenerateSlug(title)
		post.Slug = service.EnsureUniqueSlug(newSlug, id)
	}

	// 更新内容
	if content != "" {
		post.Content = content
		hasUpdates = true
	}

	// 更新摘要（允许清空）
	if excerpt != "" || c.GetHeader("Content-Type") == "application/json" {
		post.Excerpt = excerpt
		hasUpdates = true
	}

	// 更新封面图（允许清空）
	if coverImageURL != "" || c.GetHeader("Content-Type") == "application/json" {
		post.CoverImage = coverImageURL
		hasUpdates = true
	}

	// 更新状态
	if status != "" {
		oldStatus := post.Status
		post.Status = status
		hasUpdates = true
		// 如果状态从非published变为published，设置发布时间
		if oldStatus != "published" && status == "published" && post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		}
		// 如果从published改为其他状态，清空发布时间
		if oldStatus == "published" && status != "published" {
			post.PublishedAt = nil
		}
	}

	// 更新文章基本信息
	if hasUpdates {
		if err = service.UpdatePost(post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
			return
		}
	}

	// 更新分类和标签（如果提供了）
	if categoryIDs != nil || tagIDs != nil {
		if categoryIDs == nil {
			categoryIDs = []uint64{}
		}
		if tagIDs == nil {
			tagIDs = []uint64{}
		}
		if err = service.UpdatePostCategoriesAndTags(id, categoryIDs, tagIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新分类标签失败: " + err.Error()})
			return
		}
	}

	// 重新获取文章（包含最新的关联数据）
	updatedPost, categories, tags, _ := service.GetPostWithFullRelations(id)

	// 确保返回的cover_image是完整URL
	if updatedPost.CoverImage != "" {
		updatedPost.CoverImage = service.GetFullFileURL(updatedPost.CoverImage)
	}

	// 构建响应
	response := gin.H{
		"post":       updatedPost,
		"categories": categories,
		"tags":       tags,
	}
	c.JSON(http.StatusOK, response)
}

// 文章列表（管理后台，支持分页和筛选）
func ListPosts(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Size     int    `form:"size"` // 兼容旧参数名
		Q        string `form:"q"`
		Sort     string `form:"sort"`
		Category string `form:"category"`
		Tag      string `form:"tag"`
		Status   string `form:"status"` // 管理员可以筛选所有状态
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 兼容旧参数名
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = req.Size
	}

	// 使用分页服务
	result, err := service.ListPostsWithPagination(req.Page, pageSize, req.Q, req.Sort, req.Category, req.Tag, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	// 确保列表中的cover_image都是完整URL
	for i := range result.Posts {
		if result.Posts[i].CoverImage != "" {
			result.Posts[i].CoverImage = service.GetFullFileURL(result.Posts[i].CoverImage)
		}
	}

	c.JSON(http.StatusOK, result)
}
