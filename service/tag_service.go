package service

import (
	"api/dao"
	"api/models"
)

func CreateTag(tag *models.Tag) error {
	return dao.CreateTag(tag)
}
func GetTagByID(id uint64) (*models.Tag, error) {
	return dao.GetTagByID(id)
}
func ListTags() ([]models.Tag, error) {
	return dao.ListTags()
}
func UpdateTag(tag *models.Tag) error {
	return dao.UpdateTag(tag)
}
func DeleteTag(id uint64) error {
	return dao.DeleteTag(id)
}
