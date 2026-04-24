package models

import (
	"time"
)

// UserSession 用户会话表 - 管理用户登录状态和会话信息
type UserSession struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:会话唯一ID" json:"id"`
	UserID       uint64    `gorm:"index;not null;comment:关联用户ID" json:"user_id"`
	SessionToken string    `gorm:"uniqueIndex;size:255;not null;comment:会话令牌" json:"session_token"`
	ExpiresAt    time.Time `gorm:"not null;comment:会话过期时间" json:"expires_at"`
	UserAgent    string    `gorm:"type:text;comment:用户浏览器信息" json:"user_agent"`
	IPAddress    string    `gorm:"size:45;comment:用户IP地址" json:"ip_address"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

func (UserSession) TableName() string { return "user_sessions" }
