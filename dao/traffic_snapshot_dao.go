package dao

import (
	"api/database"
	"api/models"
	"encoding/json"
	"time"

	"gorm.io/gorm/clause"
)

// TrafficSnapshotDTO 用于在 ETL 流程中传输聚合后的指标。
type TrafficSnapshotDTO struct {
	LookbackDays   int
	UniqueVisitors int
	RegionCounts   map[string]int
	GeneratedAt    time.Time
}

// SaveTrafficSnapshot 将最新的统计结果 upsert 到数据库，供下游查询使用。
func SaveTrafficSnapshot(dto TrafficSnapshotDTO) error {
	payload, err := json.Marshal(dto.RegionCounts)
	if err != nil {
		return err
	}

	snapshot := models.TrafficSnapshot{
		LookbackDays:   dto.LookbackDays,
		UniqueVisitors: dto.UniqueVisitors,
		RegionJSON:     string(payload),
		GeneratedAt:    dto.GeneratedAt,
	}

	return database.GetDB().
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "lookback_days"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"unique_visitors": snapshot.UniqueVisitors, "region_json": snapshot.RegionJSON, "generated_at": snapshot.GeneratedAt, "updated_at": time.Now()}),
		}).
		Create(&snapshot).Error
}

// GetTrafficSnapshot 读取指定天数窗口的最新快照，没有可返回 gorm.ErrRecordNotFound。
func GetTrafficSnapshot(days int) (models.TrafficSnapshot, error) {
	var snapshot models.TrafficSnapshot
	err := database.GetDB().
		Where("lookback_days = ?", days).
		Order("generated_at DESC").
		First(&snapshot).Error
	return snapshot, err
}
