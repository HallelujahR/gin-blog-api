package models

import (
	"time"
)

// Post 文章表 - 博客核心内容表
// 多对多: Categories/Tags
type Post struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;comment:文章唯一ID" json:"id"`
	Title         string     `gorm:"size:200;not null;comment:标题" json:"title"`
	Slug          string     `gorm:"size:200;not null;uniqueIndex;comment:URL标识" json:"slug"`
	Excerpt       string     `gorm:"type:text;comment:文章摘要" json:"excerpt"`
	Content       string     `gorm:"type:longtext;not null;comment:文章内容" json:"content"`
	CoverImage    string     `gorm:"size:255;comment:封面图片URL" json:"cover_image"`
	AuthorID      uint64     `gorm:"index;not null;comment:作者用户ID" json:"author_id"`
	Status        string     `gorm:"type:enum('published','draft','pending','trash');default:'draft';comment:状态" json:"status"`
	Visibility    string     `gorm:"type:enum('public','private','password');default:'public';comment:可见性" json:"visibility"`
	Password      string     `gorm:"size:100;comment:访问密码" json:"password"`
	CommentStatus string     `gorm:"type:enum('open','closed');default:'open';comment:评论状态" json:"comment_status"`
	ViewCount     int        `gorm:"default:0;comment:阅读数" json:"view_count"`
	LikeCount     int        `gorm:"default:0;comment:点赞数量" json:"like_count"`
	CommentCount  int        `gorm:"default:0;comment:评论数量" json:"comment_count"`
	PublishedAt   *time.Time `gorm:"comment:发布时间" json:"published_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	Categories    []Category `gorm:"many2many:post_categories;joinForeignKey:PostID;joinReferences:CategoryID" json:"categories,omitempty"`
}

func (Post) TableName() string { return "posts" }
