package models

import (
	"time"
)

// Tag 文章标签表 - 用于文章多标签管理
type Tag struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;comment:标签唯一ID" json:"id"`
	Name        string    `gorm:"size:50;not null;comment:标签名称" json:"name"`
	Slug        string    `gorm:"size:50;not null;uniqueIndex;comment:标签URL标识" json:"slug"`
	Description string    `gorm:"type:text;comment:标签描述" json:"description"`
	PostCount   int       `gorm:"default:0;comment:文章数量" json:"post_count"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (Tag) TableName() string { return "tags" }
