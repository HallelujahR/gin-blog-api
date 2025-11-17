package stats

import (
	"api/dao"
	"api/service"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Aggregate 统计所有日志条目并生成汇总结果。
// topN 表示热门文章返回数量。
func Aggregate(entries []LogEntry, topN int) StatsResult {
	if topN <= 0 {
		topN = 3
	}

	totalVisits := len(entries)
	visitorSet := make(map[string]struct{})

	type postStat struct {
		Count  int
		Path   string
		PostID uint64
	}
	posts := make(map[string]*postStat)

	regionCount := make(map[string]int)

	for _, entry := range entries {
		if entry.IP != "" {
			visitorSet[entry.IP] = struct{}{}
		}

		sanitizedPath := sanitizePath(entry.Path)
		if sanitizedPath != "" {
			stat := posts[sanitizedPath]
			if stat == nil {
				stat = &postStat{Path: sanitizedPath, PostID: extractPostID(sanitizedPath)}
				posts[sanitizedPath] = stat
			}
			stat.Count++
		}

		region := strings.TrimSpace(entry.Region)
		if region == "" {
			region = service.LookupRegion(entry.IP)
		}
		if region == "" {
			region = "UNKNOWN"
		}
		regionCount[region]++
	}

	topPosts := make([]TopPost, 0, len(posts))
	for _, stat := range posts {
		topPosts = append(topPosts, TopPost{
			PostID: stat.PostID,
			Path:   stat.Path,
			Count:  stat.Count,
		})
	}
	sort.Slice(topPosts, func(i, j int) bool {
		return topPosts[i].Count > topPosts[j].Count
	})
	if len(topPosts) > topN {
		topPosts = topPosts[:topN]
	}
	enrichTopPosts(topPosts)

	regionStats := make([]RegionStat, 0, len(regionCount))
	for region, count := range regionCount {
		percentage := 0.0
		if totalVisits > 0 {
			percentage = float64(count*100) / float64(totalVisits)
		}
		regionStats = append(regionStats, RegionStat{
			Name:       region,
			Count:      count,
			Percentage: round(percentage, 2),
		})
	}
	sort.Slice(regionStats, func(i, j int) bool {
		return regionStats[i].Count > regionStats[j].Count
	})

	return StatsResult{
		TotalVisits:        totalVisits,
		UniqueVisitors:     len(visitorSet),
		TopPosts:           topPosts,
		RegionDistribution: regionStats,
	}
}

// sanitizePath 仅保留文章详情请求路径，其他路径返回空字符串。
func sanitizePath(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "/api/posts/") {
		return stripQuery(path)
	}
	if strings.HasPrefix(path, "/posts/") {
		return stripQuery(path)
	}
	return ""
}

func stripQuery(path string) string {
	if idx := strings.Index(path, "?"); idx >= 0 {
		return path[:idx]
	}
	return path
}

func extractPostID(path string) uint64 {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return 0
	}
	last := parts[len(parts)-1]
	id, err := strconv.ParseUint(last, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func enrichTopPosts(topPosts []TopPost) {
	var wg sync.WaitGroup
	for i := range topPosts {
		tp := &topPosts[i]
		if tp.PostID == 0 {
			continue
		}
		wg.Add(1)
		go func(target *TopPost) {
			defer wg.Done()
			post, err := dao.GetPostByID(target.PostID)
			if err != nil {
				return
			}
			target.Title = post.Title
			target.Path = fmt.Sprintf("/posts/%d", post.ID)
		}(tp)
	}
	wg.Wait()
}

// round 对浮点数进行四舍五入，保留 n 位小数。
func round(val float64, n int) float64 {
	power := 1.0
	for i := 0; i < n; i++ {
		power *= 10
	}
	return float64(int(val*power+0.5)) / power
}
