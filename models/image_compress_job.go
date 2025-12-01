package models

import "time"

// ImageCompressJob 记录每一次图片压缩任务（对应一个 job_id），便于审计与统计。
type ImageCompressJob struct {
	ID                string    `gorm:"primaryKey;size:191" json:"id"` // 与服务层 CompressJob.ID 对应
	Status            string    `gorm:"index;size:32;not null" json:"status"`
	TotalImages       int64     `json:"total_images"`
	OriginalTotal     int64     `json:"original_total"`
	CompressedTotal   int64     `json:"compressed_total"`
	TarPath           string    `gorm:"size:512" json:"tar_path"`
	ErrorMessage      string    `gorm:"size:1024" json:"error_message"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	FinishedAt        *time.Time `json:"finished_at"`
}

func (ImageCompressJob) TableName() string { return "image_compress_jobs" }


