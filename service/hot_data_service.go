package service

import (
	"api/dao"
	"api/models"
)

func CreateHotData(hd *models.HotData) error {
	return dao.CreateHotData(hd)
}
func ListHotData(dataType, period string) ([]models.HotData, error) {
	return dao.ListHotData(dataType, period)
}
func DeleteHotData(id uint64) error {
	return dao.DeleteHotData(id)
}
