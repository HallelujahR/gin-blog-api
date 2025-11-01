package dao

import (
	"api/database"
	"api/models"
)

func CreateComment(c *models.Comment) error {
	return database.GetDB().Create(c).Error
}
func GetCommentByID(id uint64) (*models.Comment, error) {
	var comment models.Comment
	err := database.GetDB().First(&comment, id).Error
	return &comment, err
}
func ListCommentsByPost(postID uint64) ([]models.Comment, error) {
	var comments []models.Comment
	err := database.GetDB().Where("post_id = ?", postID).Find(&comments).Error
	return comments, err
}
func UpdateComment(c *models.Comment) error {
	return database.GetDB().Save(c).Error
}
func DeleteComment(id uint64) error {
	return database.GetDB().Delete(&models.Comment{}, id).Error
}
