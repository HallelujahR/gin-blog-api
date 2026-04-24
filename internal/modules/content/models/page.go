package models

import "time"

// Page 静态页面表
type Page struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;comment:页面唯一ID" json:"id"`
	Title     string    `gorm:"size:200;not null;comment:标题" json:"title"`
	Slug      string    `gorm:"size:100;not null;uniqueIndex;comment:URL标识" json:"slug"`
	Content   string    `gorm:"type:longtext;not null;comment:内容" json:"content"`
	Excerpt   string    `gorm:"type:text;comment:页面摘要" json:"excerpt"`
	Template  string    `gorm:"size:50;default:'default';comment:页面模板" json:"template"`
	Status    string    `gorm:"type:enum('published','draft');default:'draft';comment:页面状态" json:"status"`
	MenuOrder int       `gorm:"default:0;comment:菜单排序" json:"menu_order"`
	ParentID  *uint64   `gorm:"index;comment:父页面ID" json:"parent_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Page) TableName() string { return "pages" }
