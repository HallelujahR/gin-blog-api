package models

import (
	"time"
)

// Category 文章分类表 - 支持多级分类
// 完整 GORM 结构：自关联Parent/Children、多对多Posts
type Category struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;comment:分类唯一ID" json:"id"`
	Name        string     `gorm:"size:100;not null;comment:分类名称" json:"name"`
	Slug        string     `gorm:"size:100;not null;uniqueIndex;comment:分类URL标识" json:"slug"`
	Description string     `gorm:"type:text;comment:分类描述" json:"description"`
	ParentID    *uint64    `gorm:"index;comment:父级分类ID" json:"parent_id"`
	Parent      *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	SortOrder   int        `gorm:"default:0;comment:排序" json:"sort_order"`
	PostCount   int        `gorm:"default:0;comment:文章数量" json:"post_count"`
	Posts       []Post     `gorm:"many2many:post_categories;joinForeignKey:CategoryID;joinReferences:PostID" json:"posts,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Category) TableName() string { return "categories" }
