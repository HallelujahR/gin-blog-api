package dao

import (
	"api/database"
	"api/models"
	"time"
)

// CreateImageCompressJob 在任务创建时写入一条记录。
func CreateImageCompressJob(id string, totalImages int, status string) error {
	if id == "" {
		return nil
	}
	job := models.ImageCompressJob{
		ID:          id,
		Status:      status,
		TotalImages: int64(totalImages),
	}
	return database.GetDB().Create(&job).Error
}

// FinishImageCompressJob 在任务完成（成功或失败）时更新结果信息。
func FinishImageCompressJob(id string, status string, originalTotal, compressedTotal int64, tarPath, errorMessage string) error {
	if id == "" {
		return nil
	}
	update := map[string]interface{}{
		"status":          status,
		"original_total":  originalTotal,
		"compressed_total": compressedTotal,
		"finished_at":     time.Now(),
	}
	if tarPath != "" {
		update["tar_path"] = tarPath
	}
	if errorMessage != "" {
		update["error_message"] = errorMessage
	}
	return database.GetDB().
		Model(&models.ImageCompressJob{}).
		Where("id = ?", id).
		Updates(update).Error
}

// CountImageCompressJobs 返回累计任务数量。
func CountImageCompressJobs() (int64, error) {
	var count int64
	err := database.GetDB().Model(&models.ImageCompressJob{}).Count(&count).Error
	return count, err
}


