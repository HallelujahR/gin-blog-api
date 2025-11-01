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
		DataType  string `json:"data_type" binding:"required"`
		DataKey   string `json:"data_key" binding:"required"`
		DataValue string `json:"data_value" binding:"required"` // 业务层需转Json
		Period    string `json:"period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hd := models.HotData{
		DataType: req.DataType,
		DataKey:  req.DataKey,
		Period:   req.Period,
		// DataValue: 需解析
	}
	if err := service.CreateHotData(&hd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": hd})
}

func ListHotData(c *gin.Context) {
	dt := c.Query("data_type")
	period := c.Query("period")
	list, err := service.ListHotData(dt, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
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
