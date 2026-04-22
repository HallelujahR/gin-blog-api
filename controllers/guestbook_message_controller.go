package controllers

import (
	"api/models"
	"api/service"
	"api/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateGuestbookMessage(c *gin.Context) {
	var req struct {
		Content     string `json:"content" binding:"required"`
		AuthorName  string `json:"author_name" binding:"required"`
		AuthorEmail string `json:"author_email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	req.AuthorName = strings.TrimSpace(req.AuthorName)
	req.AuthorEmail = strings.TrimSpace(req.AuthorEmail)

	if len([]rune(req.Content)) < 2 || len([]rune(req.Content)) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "留言内容长度需在 2 到 2000 个字符之间"})
		return
	}
	if len([]rune(req.AuthorName)) < 1 || len([]rune(req.AuthorName)) > 40 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "昵称长度需在 1 到 40 个字符之间"})
		return
	}
	if !isValidEmail(req.AuthorEmail) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱格式不正确"})
		return
	}

	message := &models.GuestbookMessage{
		Content:     req.Content,
		AuthorName:  req.AuthorName,
		AuthorEmail: req.AuthorEmail,
		AuthorIP:    utils.GetClientIP(c),
		Status:      "pending",
	}

	if err := service.CreateGuestbookMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "留言提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"notice":  "留言已提交，审核通过后会展示在留言板中",
	})
}

func ListApprovedGuestbookMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := service.ListGuestbookMessagesWithPagination(page, pageSize, "approved", "", "DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询留言失败"})
		return
	}

	c.JSON(http.StatusOK, result)
}
