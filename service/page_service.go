package service

import (
	"api/dao"
	"api/models"
)

func CreatePage(page *models.Page) error {
	return dao.CreatePage(page)
}
func GetPageByID(id uint64) (*models.Page, error) {
	return dao.GetPageByID(id)
}
func ListPages() ([]models.Page, error) {
	return dao.ListPages()
}
func UpdatePage(page *models.Page) error {
	return dao.UpdatePage(page)
}
func DeletePage(id uint64) error {
	return dao.DeletePage(id)
}
