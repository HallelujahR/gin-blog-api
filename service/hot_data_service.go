package service

import (
	"api/dao"
	"api/models"
)

func CreateHotData(hd *models.HotData) error {
	return dao.CreateHotData(hd)
}

// ListHotData 获取热点数据列表
// dataType: 数据类型，可选
// period: 统计周期，可选
// limit: 限制返回数量，默认10条
func ListHotData(dataType, period string, limit int) ([]models.HotData, error) {
	return dao.ListHotData(dataType, period, limit)
}

func DeleteHotData(id uint64) error {
	return dao.DeleteHotData(id)
}
