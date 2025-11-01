package service

import (
	"api/dao"
	"api/models"
)

func CreateCategory(c *models.Category) error {
	return dao.CreateCategory(c)
}
func GetCategoryByID(id uint64) (*models.Category, error) {
	return dao.GetCategoryByID(id)
}

// 新增
func GetCategoryByIDFull(id uint64) (*models.Category, error) {
	return dao.GetCategoryByIDFull(id)
}
func ListCategories() ([]models.Category, error) {
	return dao.ListCategories()
}
func UpdateCategory(c *models.Category) error {
	return dao.UpdateCategory(c)
}
func DeleteCategory(id uint64) error {
	return dao.DeleteCategory(id)
}
