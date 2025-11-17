package stats

import "time"

// LogEntry 表示从访问日志中解析出的单条记录。
// 在解析阶段会尽量容错：若某些字段缺失，则该行会被忽略，不影响整体统计。
type LogEntry struct {
	Timestamp  time.Time // 日志时间戳，统一为UTC
	IP         string    // 访客IP，用于计算独立访客
	Method     string    // HTTP 方法
	Path       string    // 请求路径，用于判定热门文章
	Status     int       // 响应状态码
	DurationMs int64     // 请求耗时（毫秒）
	Region     string    // 访客地区（可能为空字符串）
}

// TopPost 用于描述热门文章条目。
// Path 为文章请求路径，Count 为访问次数。
// PostID 和 Title 会在后续阶段通过数据库补充。
type TopPost struct {
	PostID uint64 `json:"post_id,omitempty"`
	Title  string `json:"title,omitempty"`
	Path   string `json:"path"`
	Count  int    `json:"count"`
}

// RegionStat 表示单个地区的访问统计信息。
// Percentage 范围为0-100，保留两位小数，方便前端直接展示。
type RegionStat struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// StatsResult 为 /api/stats 接口返回的整体统计结果。
// GeneratedAt 表示统计数据生成时间（UTC）。
type StatsResult struct {
	TotalVisits        int          `json:"total_visits"`
	UniqueVisitors     int          `json:"unique_visitors"`
	TopPosts           []TopPost    `json:"top_posts"`
	RegionDistribution []RegionStat `json:"region_distribution"`
	GeneratedAt        time.Time    `json:"generated_at"`
}

// VisitSummary 汇总数据库中的访问和热门文章信息。
type VisitSummary struct {
	TotalVisits int
	TopPosts    []TopPost
}
