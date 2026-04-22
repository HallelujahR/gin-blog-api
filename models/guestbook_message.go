package models

import "time"

// GuestbookMessage 留言板消息
type GuestbookMessage struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;comment:留言唯一ID" json:"id"`
	Content     string    `gorm:"type:text;not null;comment:留言内容" json:"content"`
	AuthorName  string    `gorm:"size:80;not null;comment:留言者昵称" json:"author_name"`
	AuthorEmail string    `gorm:"size:120;not null;comment:留言者邮箱" json:"author_email"`
	AuthorIP    string    `gorm:"size:45;comment:留言者IP" json:"author_ip"`
	Status      string    `gorm:"type:enum('approved','pending','spam','trash');default:'pending';comment:留言状态" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (GuestbookMessage) TableName() string { return "guestbook_messages" }
