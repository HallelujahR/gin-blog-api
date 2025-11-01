package models

import (
	"time"
)

// Like 点赞表 - 记录用户对文章、评论的点赞行为
// post_id 和 comment_id 必须至少一个不为空
// user_id+post_id 或 user_id+comment_id 唯一
// 关联 users、posts、comments 表
type Like struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:点赞记录ID" json:"id"`
	UserID    uint64    `gorm:"not null;index;comment:点赞用户ID" json:"user_id"`
	PostID    *uint64   `gorm:"index;comment:被点赞文章ID" json:"post_id"`
	CommentID *uint64   `gorm:"index;comment:被点赞评论ID" json:"comment_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:点赞时间" json:"created_at"`
}

// TableName 设置表名为 likes
func (Like) TableName() string {
	return "likes"
}
