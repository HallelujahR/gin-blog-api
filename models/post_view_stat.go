package models

import "time"

// PostViewStat 记录文章每天的访问次数，用于近30日统计。
type PostViewStat struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    uint64    `gorm:"index:idx_post_view_date,priority:1;not null" json:"post_id"`
	Date      time.Time `gorm:"type:date;index:idx_post_view_date,priority:2;not null" json:"date"`
	Views     int       `gorm:"default:0;not null" json:"views"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PostViewStat) TableName() string { return "post_view_stats" }
