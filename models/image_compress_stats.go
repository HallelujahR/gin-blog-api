package models

import "time"

// ImageCompressStats 保存全局累计的图片压缩统计信息。
// 当前设计为单行表，记录“历史累计成功处理”的图片总数与字节数。
type ImageCompressStats struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TotalImages         int64     `gorm:"not null;default:0" json:"total_images"`
	TotalOriginalBytes  int64     `gorm:"not null;default:0" json:"total_original_bytes"`
	TotalCompressedBytes int64    `gorm:"not null;default:0" json:"total_compressed_bytes"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ImageCompressStats) TableName() string { return "image_compress_stats" }


