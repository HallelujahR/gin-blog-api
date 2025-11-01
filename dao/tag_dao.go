package dao

import (
	"api/database"
	"api/models"
)

func CreateTag(tag *models.Tag) error {
	return database.GetDB().Create(tag).Error
}
func GetTagByID(id uint64) (*models.Tag, error) {
	var tag models.Tag
	err := database.GetDB().First(&tag, id).Error
	return &tag, err
}
func ListTags() ([]models.Tag, error) {
	var tags []models.Tag
	err := database.GetDB().Find(&tags).Error
	return tags, err
}
func UpdateTag(tag *models.Tag) error {
	return database.GetDB().Save(tag).Error
}
func DeleteTag(id uint64) error {
	return database.GetDB().Delete(&models.Tag{}, id).Error
}
