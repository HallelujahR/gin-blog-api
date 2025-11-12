package models

import (
	"time"
)

// Like 点赞表 - 记录IP对文章、评论的点赞行为
// post_id 和 comment_id 必须至少一个不为空
// user_id+post_id 或 user_id+comment_id 唯一
// user_id 可以是真实用户ID或虚拟用户ID（基于IP生成，范围1000000000-9999999999）
// 注意：此表不应有外键约束关联users表，因为虚拟user_id不存在于users表中
type Like struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:点赞记录ID" json:"id"`
	UserID    uint64    `gorm:"not null;index;comment:点赞用户ID（真实或虚拟）" json:"user_id"`
	PostID    *uint64   `gorm:"index;comment:被点赞文章ID" json:"post_id"`
	CommentID *uint64   `gorm:"index;comment:被点赞评论ID" json:"comment_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:点赞时间" json:"created_at"`
}

// TableName 设置表名为 likes
func (Like) TableName() string {
	return "likes"
}
