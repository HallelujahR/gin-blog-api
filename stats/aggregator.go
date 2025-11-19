package stats

import (
	"sort"
)

// SnapshotData 表示来自数据库快照的汇总结果。
type SnapshotData struct {
	UniqueVisitors int
	RegionCounts   map[string]int
}

// Aggregate 根据快照与数据库摘要构建最终响应。
func Aggregate(snapshot SnapshotData, summary VisitSummary) StatsResult {
	regionStats := buildRegionStats(snapshot.RegionCounts)
	return StatsResult{
		TotalVisits:        summary.TotalVisits,
		UniqueVisitors:     snapshot.UniqueVisitors,
		TopPosts:           summary.TopPosts,
		RegionDistribution: regionStats,
	}
}

func buildRegionStats(regionCounts map[string]int) []RegionStat {
	total := 0
	for _, count := range regionCounts {
		if count > 0 {
			total += count
		}
	}
	regionStats := make([]RegionStat, 0, len(regionCounts))
	for region, count := range regionCounts {
		if count <= 0 {
			continue
		}
		percentage := 0.0
		if total > 0 {
			percentage = round(float64(count*100)/float64(total), 2)
		}
		regionStats = append(regionStats, RegionStat{
			Name:       region,
			Percentage: percentage,
		})
	}
	sort.Slice(regionStats, func(i, j int) bool {
		return regionCounts[regionStats[i].Name] > regionCounts[regionStats[j].Name]
	})
	return regionStats
}

// round 对浮点数进行四舍五入，保留 n 位小数。
func round(val float64, n int) float64 {
	power := 1.0
	for i := 0; i < n; i++ {
		power *= 10
	}
	return float64(int(val*power+0.5)) / power
}
