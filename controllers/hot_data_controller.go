package controllers

import (
	"api/models"
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateHotData(c *gin.Context) {
	var req struct {
		// DataType  string `json:"data_type" binding:"required"`
		// DataKey   string `json:"data_key" binding:"required"`
		// DataValue string `json:"data_value" binding:"required"` // 业务层需转Json
		Period string `json:"period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hd := models.HotData{
		// DataType: req.DataType,
		// DataKey:  req.DataKey,
		Period: req.Period,
		// DataValue: 需解析
	}
	if err := service.CreateHotData(&hd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": hd})
}

func ListHotData(c *gin.Context) {
	// 获取查询参数
	dataType := c.Query("data_type") // 可选：trending_posts, popular_tags, active_users
	period := c.Query("period")      // 可选：daily, weekly, monthly, all_time

	// 默认返回前10条热点数据
	limit := 10

	// 如果前端指定了limit参数，使用指定值（但最多不超过20条）
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 20 {
				limit = 20 // 最多返回20条
			} else {
				limit = parsedLimit
			}
		}
	}

	list, err := service.ListHotData(dataType, period, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 返回格式：{ list: [...] }，符合前端期望
	c.JSON(http.StatusOK, gin.H{"list": list})
}

func DeleteHotData(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteHotData(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
