package dao

import (
	"api/database"
	"api/models"
)

func CreatePage(page *models.Page) error {
	return database.GetDB().Create(page).Error
}
func GetPageByID(id uint64) (*models.Page, error) {
	var page models.Page
	err := database.GetDB().First(&page, id).Error
	return &page, err
}

// GetPageBySlug 通过slug获取页面
func GetPageBySlug(slug string) (*models.Page, error) {
	var page models.Page
	err := database.GetDB().Where("slug = ? AND status = ?", slug, "published").First(&page).Error
	return &page, err
}
func ListPages() ([]models.Page, error) {
	var pages []models.Page
	err := database.GetDB().Find(&pages).Error
	return pages, err
}
func UpdatePage(page *models.Page) error {
	return database.GetDB().Save(page).Error
}
func DeletePage(id uint64) error {
	return database.GetDB().Delete(&models.Page{}, id).Error
}
