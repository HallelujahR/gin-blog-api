package analytics

import "time"

// Event 表示一次 HTTP 请求的原始埋点事件，由 Gin 中间件异步写入。
type Event struct {
	Timestamp time.Time
	IP        string
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	UserAgent string
}
