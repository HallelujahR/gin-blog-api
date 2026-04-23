package admin

import (
	"api/models"
	"api/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func ListMoments(c *gin.Context) {
	moments, err := service.ListAllMoments(200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"moments": moments})
}

func GetMoment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	moment, err := service.GetMomentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"moment": moment})
}

func CreateMoment(c *gin.Context) {
	var req struct {
		Content string   `json:"content" binding:"required"`
		Images  []string `json:"images"`
		Mood    string   `json:"mood"`
		Status  string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误: " + err.Error()})
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "内容不能为空"})
		return
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	moment := &models.Moment{
		Content: req.Content,
		Images:  req.Images,
		Mood:    strings.TrimSpace(req.Mood),
		Status:  status,
	}

	if err := service.CreateMoment(moment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"moment": moment})
}

func UpdateMoment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	moment, err := service.GetMomentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "动态不存在"})
		return
	}

	var req struct {
		Content string   `json:"content"`
		Images  []string `json:"images"`
		Mood    string   `json:"mood"`
		Status  string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误: " + err.Error()})
		return
	}

	if strings.TrimSpace(req.Content) != "" {
		moment.Content = strings.TrimSpace(req.Content)
	}
	if req.Images != nil {
		moment.Images = req.Images
	}
	if req.Mood != "" || c.GetHeader("Content-Type") == "application/json" {
		moment.Mood = strings.TrimSpace(req.Mood)
	}
	if req.Status != "" {
		moment.Status = req.Status
	}

	if err := service.UpdateMoment(moment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"moment": moment})
}

func DeleteMoment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteMoment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
