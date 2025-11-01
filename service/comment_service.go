package service

import (
	"api/dao"
	"api/models"
)

func CreateComment(c *models.Comment) error {
	return dao.CreateComment(c)
}
func GetCommentByID(id uint64) (*models.Comment, error) {
	return dao.GetCommentByID(id)
}
func ListCommentsByPost(postID uint64) ([]models.Comment, error) {
	return dao.ListCommentsByPost(postID)
}
func UpdateComment(c *models.Comment) error {
	return dao.UpdateComment(c)
}
func DeleteComment(id uint64) error {
	return dao.DeleteComment(id)
}
