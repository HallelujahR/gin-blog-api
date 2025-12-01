package dao

import (
	"api/database"
	"api/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AddImageCompressStats 将一次成功任务的统计累加到全局统计表中。
// nImages: 本次成功压缩的图片数量
// originalBytes / compressedBytes: 本次任务对应的累计原始/压缩字节数
func AddImageCompressStats(nImages int64, originalBytes, compressedBytes int64) error {
	if nImages <= 0 || originalBytes < 0 || compressedBytes < 0 {
		return nil
	}

	db := database.GetDB()

	// 设计为单行表，这里通过主键 ID=1 实现幂等累加。
	stats := models.ImageCompressStats{
		ID:                  1,
		TotalImages:         nImages,
		TotalOriginalBytes:  originalBytes,
		TotalCompressedBytes: compressedBytes,
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_images":          gorm.Expr("total_images + ?", nImages),
			"total_original_bytes":  gorm.Expr("total_original_bytes + ?", originalBytes),
			"total_compressed_bytes": gorm.Expr("total_compressed_bytes + ?", compressedBytes),
			"updated_at":            time.Now(),
		}),
	}).Create(&stats).Error
}

// GetImageCompressStats 读取全局累计压缩统计，若不存在则返回零值。
func GetImageCompressStats() (models.ImageCompressStats, error) {
	var stats models.ImageCompressStats
	err := database.GetDB().First(&stats, 1).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ImageCompressStats{}, nil
		}
		return models.ImageCompressStats{}, err
	}
	return stats, nil
}


