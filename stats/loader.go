package stats

import (
	"context"
	"os"
	"time"

	"api/service"
)

// LoadRecentEntries 读取最近 days 天的访问日志并解析。
func LoadRecentEntries(ctx context.Context, days int) ([]LogEntry, error) {
	if days <= 0 {
		days = 30
	}

	entries := make([]LogEntry, 0, days*100)
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		path, err := service.AccessLogPathFor(date)
		if err != nil {
			return entries, err
		}

		select {
		case <-ctx.Done():
			return entries, ctx.Err()
		default:
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return entries, err
		}
		if !info.Mode().IsRegular() {
			continue
		}

		dailyEntries, err := ParseLogFile(ctx, path)
		if err != nil {
			return entries, err
		}
		entries = append(entries, dailyEntries...)
	}
	return entries, nil
}
