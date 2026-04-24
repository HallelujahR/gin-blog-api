package dao

import (
	"api/internal/modules/content/models"
	"api/internal/platform/db"
)

func CreateMoment(moment *models.Moment) error {
	return database.GetDB().Create(moment).Error
}

func GetMomentByID(id uint64) (*models.Moment, error) {
	var moment models.Moment
	err := database.GetDB().First(&moment, id).Error
	return &moment, err
}

func UpdateMoment(moment *models.Moment) error {
	return database.GetDB().Save(moment).Error
}

func DeleteMoment(id uint64) error {
	return database.GetDB().Delete(&models.Moment{}, id).Error
}

func ListMoments(status string, limit int) ([]models.Moment, error) {
	var moments []models.Moment
	db := database.GetDB().Model(&models.Moment{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	db = db.Order("COALESCE(published_at, created_at) DESC, id DESC")
	if limit > 0 {
		db = db.Limit(limit)
	}
	err := db.Find(&moments).Error
	return moments, err
}
