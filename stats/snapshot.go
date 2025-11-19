package stats

import (
	"api/dao"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// LoadTrafficSnapshot 从数据库中加载近 N 日的快照数据，若不存在返回零值结构。
func LoadTrafficSnapshot(days int) (SnapshotData, error) {
	snapshot, err := dao.GetTrafficSnapshot(days)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SnapshotData{RegionCounts: map[string]int{}}, nil
		}
		return SnapshotData{}, err
	}

	data := SnapshotData{
		UniqueVisitors: snapshot.UniqueVisitors,
		RegionCounts:   map[string]int{},
	}
	if snapshot.RegionJSON != "" {
		if err := json.Unmarshal([]byte(snapshot.RegionJSON), &data.RegionCounts); err != nil {
			return SnapshotData{}, err
		}
	}
	return data, nil
}
