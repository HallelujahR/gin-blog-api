package models

import "time"

// Comment 文章评论表，支持多级评论
type Comment struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:评论唯一ID" json:"id"`
	Content      string    `gorm:"type:text;not null;comment:评论内容" json:"content"`
	AuthorName   string    `gorm:"size:100;not null;comment:评论者名称" json:"author_name"`
	AuthorEmail  string    `gorm:"size:100;not null;comment:评论者邮箱" json:"author_email"`
	AuthorURL    string    `gorm:"size:200;comment:评论者网站" json:"author_url"`
	AuthorIP     string    `gorm:"size:45;comment:评论者IP" json:"author_ip"`
	AuthorUserID *uint64   `gorm:"index;comment:评论者用户ID" json:"author_user_id"`
	PostID       uint64    `gorm:"index;not null;comment:关联文章ID" json:"post_id"`
	ParentID     *uint64   `gorm:"index;comment:父评论ID" json:"parent_id"`
	Status       string    `gorm:"type:enum('approved','pending','spam','trash');default:'pending';comment:评论状态" json:"status"`
	LikeCount    int       `gorm:"default:0;comment:评论点赞数" json:"like_count"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Comment) TableName() string { return "comments" }
