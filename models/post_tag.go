package models

import "time"

// PostTag 文章标签关联表
type PostTag struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:关联记录ID" json:"id"`
	PostID    uint64    `gorm:"index;not null;comment:文章ID" json:"post_id"`
	TagID     uint64    `gorm:"index;not null;comment:标签ID" json:"tag_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (PostTag) TableName() string { return "post_tags" }
