package admin

import (
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 评论列表（管理后台）
func ListComments(c *gin.Context) {
	var postID *uint64
	if v := c.Query("post_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		postID = &id
	}
	
	comments, err := service.ListCommentsByPost(*postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// 删除评论（管理后台）
func DeleteComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteComment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 更新评论状态（管理后台）
func UpdateCommentStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	comment, err := service.GetCommentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "不存在"})
		return
	}
	
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	comment.Status = req.Status
	if err = service.UpdateComment(comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comment": comment})
}
