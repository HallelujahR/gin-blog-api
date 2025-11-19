package stats

import (
	"api/service"
	"sort"
	"strings"
)

// Aggregate 根据日志和数据库汇总结果生成统计数据。
// entries 仅用于计算独立访客和地区分布，summary 来源于数据库。
func Aggregate(entries []LogEntry, summary VisitSummary) StatsResult {
	visitorSet := make(map[string]struct{})
	regionCount := make(map[string]int)
	totalRegionEntries := 0

	for _, entry := range entries {
		if entry.IP != "" {
			visitorSet[entry.IP] = struct{}{}
		}
		// 只统计中国IP的地区分布
		if service.IsChinaIP(entry.IP) {
			if region, ok := normalizeRegion(entry.Region, entry.IP); ok {
				regionCount[region]++
				totalRegionEntries++
			}
		}
	}

	regionStats := make([]RegionStat, 0, len(regionCount))
	for region, count := range regionCount {
		percentage := 0.0
		if totalRegionEntries > 0 {
			percentage = float64(count*100) / float64(totalRegionEntries)
		}
		regionStats = append(regionStats, RegionStat{
			Name:       region,
			Percentage: round(percentage, 2),
		})
	}
	// 按访问次数降序排序（使用 count 变量）
	sort.Slice(regionStats, func(i, j int) bool {
		return regionCount[regionStats[i].Name] > regionCount[regionStats[j].Name]
	})

	return StatsResult{
		TotalVisits:        summary.TotalVisits,
		UniqueVisitors:     len(visitorSet),
		TopPosts:           summary.TopPosts,
		RegionDistribution: regionStats,
	}
}

// round 对浮点数进行四舍五入，保留 n 位小数。
func round(val float64, n int) float64 {
	power := 1.0
	for i := 0; i < n; i++ {
		power *= 10
	}
	return float64(int(val*power+0.5)) / power
}

// normalizeRegion 返回可统计的地区名，忽略本地访问并将空值标记为 UNKNOWN。
func normalizeRegion(value, ip string) (string, bool) {
	region := strings.TrimSpace(value)
	if region == "" {
		region = service.LookupRegion(ip)
	}
	if region == "" {
		region = "UNKNOWN"
	}
	if isIgnoredRegion(region) {
		return "", false
	}
	return region, true
}

func isIgnoredRegion(region string) bool {
	value := strings.TrimSpace(region)
	return strings.EqualFold(value, "local") || strings.EqualFold(value, "unknown")
}
