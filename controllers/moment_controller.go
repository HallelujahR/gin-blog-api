package controllers

import (
	"api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListMoments(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	moments, err := service.ListPublishedMoments(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询动态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"moments": moments})
}
