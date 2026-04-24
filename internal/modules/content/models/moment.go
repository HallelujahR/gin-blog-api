package models

import "time"

// Moment 碎碎念 / 动态
type Moment struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement;comment:动态唯一ID" json:"id"`
	Content     string     `gorm:"type:text;not null;comment:动态内容" json:"content"`
	Images      []string   `gorm:"serializer:json;type:json;comment:图片列表" json:"images"`
	Mood        string     `gorm:"size:50;comment:心情标记" json:"mood"`
	Status      string     `gorm:"type:enum('published','draft');default:'draft';comment:状态" json:"status"`
	PublishedAt *time.Time `gorm:"index;comment:发布时间" json:"published_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Moment) TableName() string { return "moments" }
