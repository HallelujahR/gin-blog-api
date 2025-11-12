package dao

import (
	"api/database"
	"api/models"
)

// 查询是否点赞（基于userID）
func GetLike(userID uint64, postID, commentID *uint64) (*models.Like, error) {
	var like models.Like
	db := database.GetDB().Where("user_id = ?", userID)
	if postID != nil {
		db = db.Where("post_id = ?", *postID)
	}
	if commentID != nil {
		db = db.Where("comment_id = ?", *commentID)
	}
	err := db.First(&like).Error
	return &like, err
}

// 点赞
func CreateLike(like *models.Like) error {
	return database.GetDB().Create(like).Error
}

// 取消点赞
func DeleteLike(like *models.Like) error {
	return database.GetDB().Delete(like).Error
}

// 统计点赞数
func CountLikes(postID, commentID *uint64) (int64, error) {
	db := database.GetDB().Model(&models.Like{})
	if postID != nil {
		db = db.Where("post_id = ?", *postID)
	}
	if commentID != nil {
		db = db.Where("comment_id = ?", *commentID)
	}
	var count int64
	err := db.Count(&count).Error
	return count, err
}
