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

// 批量删除评论
func DeleteComments(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return database.GetDB().Where("id IN ?", ids).Delete(&models.Comment{}).Error
}

// 统计评论总数（用于分页）
func CountComments(postID uint64, status, q string) (int64, error) {
	var count int64
	db := database.GetDB().Model(&models.Comment{})
	
	if postID > 0 {
		db = db.Where("post_id = ?", postID)
	}
	
	if status != "" {
		db = db.Where("status = ?", status)
	}
	
	if q != "" {
		db = db.Where("content LIKE ? OR author_name LIKE ? OR author_email LIKE ?", 
			"%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	
	err := db.Count(&count).Error
	return count, err
}

// 评论列表带分页和筛选
func ListCommentsWithParams(page, pageSize int, postID uint64, status, q, sort string) ([]models.Comment, error) {
	var comments []models.Comment
	db := database.GetDB().Model(&models.Comment{})
	
	if postID > 0 {
		db = db.Where("post_id = ?", postID)
	}
	
	if status != "" {
		db = db.Where("status = ?", status)
	}
	
	if q != "" {
		db = db.Where("content LIKE ? OR author_name LIKE ? OR author_email LIKE ?", 
			"%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	
	// 排序
	if sort != "" {
		db = db.Order("created_at " + sort)
	} else {
		db = db.Order("created_at DESC")
	}
	
	// 分页
	if pageSize > 0 {
		db = db.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	
	err := db.Find(&comments).Error
	return comments, err
}

// 批量更新评论状态
func UpdateCommentsStatus(ids []uint64, status string) error {
	if len(ids) == 0 {
		return nil
	}
	return database.GetDB().Model(&models.Comment{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// 获取评论的回复列表
func GetCommentReplies(commentID uint64) ([]models.Comment, error) {
	var comments []models.Comment
	err := database.GetDB().
		Where("parent_id = ?", commentID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}
