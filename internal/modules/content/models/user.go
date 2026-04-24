package models

import (
	"time"
)

// User 用户表 - 存储系统所有用户信息，包括管理员、作者和普通用户
// 支持用户名、邮箱唯一，用户角色和状态、验证邮箱、最后登录等
type User struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement;comment:用户唯一ID" json:"id"`
	Username      string     `gorm:"uniqueIndex;size:50;not null;comment:用户名" json:"username"`
	Email         string     `gorm:"uniqueIndex;size:100;not null;comment:用户邮箱" json:"email"`
	PasswordHash  string     `gorm:"size:255;not null;comment:加密后的密码" json:"password_hash"`
	DisplayName   string     `gorm:"size:100;comment:显示名称" json:"display_name"`
	AvatarURL     string     `gorm:"size:255;comment:头像URL" json:"avatar_url"`
	Bio           string     `gorm:"type:text;comment:个人简介" json:"bio"`
	Website       string     `gorm:"size:200;comment:个人网站" json:"website"`
	Location      string     `gorm:"size:100;comment:所在地" json:"location"`
	Role          string     `gorm:"type:enum('admin','author','subscriber');default:'subscriber';comment:角色" json:"role"`
	Status        string     `gorm:"type:enum('active','inactive','banned');default:'active';comment:状态" json:"status"`
	EmailVerified bool       `gorm:"default:false;comment:邮箱是否验证" json:"email_verified"`
	LastLoginAt   *time.Time `gorm:"comment:最后登录时间" json:"last_login_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (User) TableName() string { return "users" }
