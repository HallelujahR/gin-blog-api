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

// 获取文章详情（包含完整的分类和标签信息）
func GetPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 使用GetPostWithFullRelations获取文章及其关联的分类和标签
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

	// 构建响应数据，符合前端期望的数据结构
	// 将categories和tags直接添加到post对象中，方便前端直接使用
	postWithRelations := gin.H{
		"id":            post.ID,
		"title":         post.Title,
		"slug":          post.Slug,
		"content":       post.Content,
		"excerpt":       post.Excerpt,
		"cover_image":   post.CoverImage,
		"author_id":     post.AuthorID,
		"status":        post.Status,
		"visibility":    post.Visibility,
		"view_count":    post.ViewCount,
		"like_count":    post.LikeCount,
		"comment_count": post.CommentCount,
		"published_at":  post.PublishedAt,
		"created_at":    post.CreatedAt,
		"updated_at":    post.UpdatedAt,
		"categories":    categories, // 直接包含完整的分类对象数组 [{id, name, slug}, ...]
		"tags":          tags,       // 直接包含完整的标签对象数组 [{id, name, slug}, ...]
		"category_ids":  categoryIDs,
		"tag_ids":       tagIDs,
	}

	// 返回响应，post对象中包含完整的categories和tags
	response := gin.H{
		"post":       postWithRelations,
		"categories": categories, // 顶层也返回，方便前端使用
		"tags":       tags,       // 顶层也返回，方便前端使用
	}

	c.JSON(http.StatusOK, response)
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
	posts, err := service.ListPostsWithParams(req.Page, req.Size, req.Q, req.Sort, req.Category, req.Tag, "published")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 确保列表中的cover_image都是完整URL
	for i := range posts {
		if posts[i].CoverImage != "" {
			posts[i].CoverImage = service.GetFullFileURL(posts[i].CoverImage)
		}
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
