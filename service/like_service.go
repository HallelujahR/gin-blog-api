package service

import (
	"api/dao"
	"api/database"
	"api/models"
	"gorm.io/gorm"
)

// ToggleLike 切换点赞状态（基于IP，通过IP生成虚拟userID，确保同一IP只能点赞一次）
func ToggleLike(ip string, postID, commentID *uint64) (string, error) {
	// 使用IP生成虚拟userID
	virtualUserID := ipToVirtualUserID(ip)
	
	like, err := dao.GetLike(virtualUserID, postID, commentID)
	if err == nil && like.ID != 0 {
		// 已点赞，取消点赞
		err = dao.DeleteLike(like)
		if err == nil && postID != nil {
			// 更新文章点赞数
			updatePostLikeCount(*postID, -1)
		}
		return "取消点赞", err
	}
	// 未点赞，添加点赞
	mod := &models.Like{UserID: virtualUserID, PostID: postID, CommentID: commentID}
	err = dao.CreateLike(mod)
	if err == nil && postID != nil {
		// 更新文章点赞数
		updatePostLikeCount(*postID, 1)
	}
	return "点赞成功", err
}

// ipToVirtualUserID 将IP转换为虚拟userID（使用简单的hash算法）
func ipToVirtualUserID(ip string) uint64 {
	// 使用IP字符串的hash值作为虚拟userID
	// 这里使用简单的字符串hash，确保同一IP总是生成相同的userID
	hash := uint64(0)
	for _, char := range ip {
		hash = hash*31 + uint64(char)
	}
	// 使用一个大的偏移量，避免与真实userID冲突（假设真实userID不会超过1000000）
	return hash%9000000000 + 1000000000
}

// updatePostLikeCount 更新文章点赞数
func updatePostLikeCount(postID uint64, delta int) {
	db := database.GetDB()
	db.Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + ?", delta))
}

func CountLikes(postID, commentID *uint64) (int64, error) {
	return dao.CountLikes(postID, commentID)
}
