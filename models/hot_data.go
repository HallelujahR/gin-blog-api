package models

import (
	"time"

	"gorm.io/datatypes"
)

// HotData 热点数据表 - 缓存热门文章、标签、用户等
type HotData struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement;comment:热点数据ID" json:"id"`
	DataType     string         `gorm:"type:enum('trending_posts','popular_tags','active_users');not null;comment:数据类型" json:"data_type"`
	DataKey      string         `gorm:"size:100;not null;comment:数据键名" json:"data_key"`
	DataValue    datatypes.JSON `gorm:"type:json;not null;comment:数据值(json格式)" json:"data_value"`
	Score        float64        `gorm:"type:decimal(10,4);default:0;comment:热度分数" json:"score"`
	Period       string         `gorm:"type:enum('daily','weekly','monthly','all_time');default:'all_time';comment:统计周期" json:"period"`
	CalculatedAt time.Time      `gorm:"autoCreateTime;comment:计算时间" json:"calculated_at"`
	ExpiresAt    *time.Time     `gorm:"comment:过期时间" json:"expires_at"`
}

func (HotData) TableName() string { return "hot_data" }
