package service

import (
	"api/dao"
	"api/models"
)

func ToggleLike(userID uint64, postID, commentID *uint64) (string, error) {
	like, err := dao.GetLike(userID, postID, commentID)
	if err == nil && like.ID != 0 {
		err = dao.DeleteLike(like)
		return "取消点赞", err
	}
	mod := &models.Like{UserID: userID, PostID: postID, CommentID: commentID}
	err = dao.CreateLike(mod)
	return "点赞成功", err
}

func CountLikes(postID, commentID *uint64) (int64, error) {
	return dao.CountLikes(postID, commentID)
}
