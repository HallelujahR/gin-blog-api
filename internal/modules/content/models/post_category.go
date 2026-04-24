package models

import "time"

// PostCategory 文章分类关联表
type PostCategory struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;comment:关联记录ID" json:"id"`
	PostID     uint64    `gorm:"index;not null;comment:文章ID" json:"post_id"`
	CategoryID uint64    `gorm:"index;not null;comment:分类ID" json:"category_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (PostCategory) TableName() string { return "post_categories" }
