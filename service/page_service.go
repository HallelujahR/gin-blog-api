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

// GetPageBySlug 通过slug获取已发布的页面
func GetPageBySlug(slug string) (*models.Page, error) {
	return dao.GetPageBySlug(slug)
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
