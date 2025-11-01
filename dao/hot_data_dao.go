package dao

import (
	"api/database"
	"api/models"
)

func CreateHotData(data *models.HotData) error {
	return database.GetDB().Create(data).Error
}
func ListHotData(dataType, period string) ([]models.HotData, error) {
	var hd []models.HotData
	db := database.GetDB().Where("data_type = ? AND period = ?", dataType, period)
	err := db.Find(&hd).Error
	return hd, err
}
func DeleteHotData(id uint64) error {
	return database.GetDB().Delete(&models.HotData{}, id).Error
}
