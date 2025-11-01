package dao

import (
	"api/database"
	"api/models"
)

func CreateCategory(c *models.Category) error {
	return database.GetDB().Create(c).Error
}
func GetCategoryByID(id uint64) (*models.Category, error) {
	var cat models.Category
	err := database.GetDB().First(&cat, id).Error
	return &cat, err
}

// 查询带Posts和Children（推荐service使用）
func GetCategoryByIDFull(id uint64) (*models.Category, error) {
	var cat models.Category
	err := database.GetDB().
		Preload("Parent").
		Preload("Children").
		Preload("Posts").
		First(&cat, id).Error
	return &cat, err
}
func ListCategories() ([]models.Category, error) {
	var cats []models.Category
	err := database.GetDB().Find(&cats).Error
	return cats, err
}
func UpdateCategory(c *models.Category) error {
	return database.GetDB().Save(c).Error
}
func DeleteCategory(id uint64) error {
	return database.GetDB().Delete(&models.Category{}, id).Error
}
