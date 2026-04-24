package admin

import (
	"api/internal/modules/content/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListGuestbookMessages(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Q        string `form:"q"`
		Status   string `form:"status"`
		Sort     string `form:"sort"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := service.ListGuestbookMessagesWithPagination(req.Page, req.PageSize, req.Status, req.Q, req.Sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func UpdateGuestbookMessageStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	message, err := service.GetGuestbookMessageByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "留言不存在"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	message.Status = req.Status
	if err = service.UpdateGuestbookMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func DeleteGuestbookMessage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteGuestbookMessage(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
