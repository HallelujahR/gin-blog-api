package dao

import (
	"api/database"
	"api/models"
)

func CreateHotData(data *models.HotData) error {
	return database.GetDB().Create(data).Error
}

// ListHotData 获取热点数据列表
// dataType: 数据类型，可选（trending_posts, popular_tags, active_users）
// period: 统计周期，可选（daily, weekly, monthly, all_time）
// limit: 限制返回数量，默认10条
func ListHotData(dataType, period string, limit int) ([]models.HotData, error) {
	var hd []models.HotData
	db := database.GetDB()
	
	// 如果指定了dataType，添加过滤条件
	if dataType != "" {
		db = db.Where("data_type = ?", dataType)
	}
	
	// 如果指定了period，添加过滤条件
	if period != "" {
		db = db.Where("period = ?", period)
	}
	
	// 设置默认limit为10
	if limit <= 0 {
		limit = 10
	}
	
	// 按score降序排序，限制返回数量
	err := db.Order("score DESC").Limit(limit).Find(&hd).Error
	return hd, err
}

func DeleteHotData(id uint64) error {
	return database.GetDB().Delete(&models.HotData{}, id).Error
}
