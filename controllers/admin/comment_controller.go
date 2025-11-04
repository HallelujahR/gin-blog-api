package admin

import (
	"api/models"
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 评论列表（管理后台，支持分页和筛选）
func ListComments(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		PostID   uint64 `form:"post_id"` // 可选：筛选特定文章
		Status   string `form:"status"`  // 可选：筛选状态 approved/pending/spam/trash
		Q        string `form:"q"`       // 可选：搜索关键词
		Sort     string `form:"sort"`    // 可选：排序 ASC/DESC
	}
	
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用分页服务
	result, err := service.ListCommentsWithPagination(
		req.Page, 
		req.PageSize, 
		req.PostID, 
		req.Status, 
		req.Q, 
		req.Sort,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// 获取评论详情（包含回复）
func GetComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	comment, replies, err := service.GetCommentWithReplies(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}
	
	response := gin.H{
		"comment": comment,
		"replies":  replies,
	}
	c.JSON(http.StatusOK, response)
}

// 删除单个评论（管理后台）
func DeleteComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteComment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 批量删除评论（管理后台）
func DeleteComments(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.DeleteComments(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "批量删除成功", "count": len(req.IDs)})
}

// 更新单个评论状态（管理后台）
func UpdateCommentStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	comment, err := service.GetCommentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"approved": true,
		"pending":  true,
		"spam":     true,
		"trash":    true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态值"})
		return
	}

	comment.Status = req.Status
	if err = service.UpdateComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// 批量更新评论状态（管理后台）
func UpdateCommentsStatus(c *gin.Context) {
	var req struct {
		IDs    []uint64 `json:"ids" binding:"required"`
		Status string   `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"approved": true,
		"pending":  true,
		"spam":     true,
		"trash":    true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态值"})
		return
	}

	if err := service.UpdateCommentsStatus(req.IDs, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "批量更新成功", "count": len(req.IDs)})
}

// 回复评论（管理后台）
func ReplyComment(c *gin.Context) {
	parentID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	
	// 验证父评论是否存在
	parent, err := service.GetCommentByID(parentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "父评论不存在"})
		return
	}

	var req struct {
		Content     string `json:"content" binding:"required"`
		AuthorName  string `json:"author_name" binding:"required"`
		AuthorEmail string `json:"author_email" binding:"required"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建回复评论
	reply := &models.Comment{
		Content:     req.Content,
		AuthorName:  req.AuthorName,
		AuthorEmail: req.AuthorEmail,
		PostID:      parent.PostID,
		ParentID:    &parentID,
		Status:      "approved", // 管理员回复自动审核通过
	}

	if err = service.CreateComment(reply); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "回复失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": reply})
}
