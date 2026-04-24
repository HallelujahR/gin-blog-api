package models

import "time"

// TrafficSnapshot 存储近 N 日的独立访客与地区统计快照，供 Stats API 与可视化使用。
type TrafficSnapshot struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	LookbackDays   int       `gorm:"uniqueIndex:idx_snapshot_lookback;not null" json:"lookback_days"`
	UniqueVisitors int       `gorm:"not null;default:0" json:"unique_visitors"`
	RegionJSON     string    `gorm:"type:longtext" json:"region_json"`
	GeneratedAt    time.Time `gorm:"autoCreateTime" json:"generated_at"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TrafficSnapshot) TableName() string { return "traffic_snapshots" }
