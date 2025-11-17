package stats

import (
	"context"
	"sync"
	"time"
)

var schedulerOnce sync.Once

func ensureScheduler() {
	schedulerOnce.Do(func() {
		go scheduleDailyRefresh()
	})
}

func scheduleDailyRefresh() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		timer := time.NewTimer(next.Sub(now))
		<-timer.C
		runRefresh()
	}
}

func runRefresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	entries, err := LoadRecentEntries(ctx, 30)
	if err != nil {
		return
	}
	result := Aggregate(entries, 3)
	result.GeneratedAt = time.Now().UTC()
	_ = SetStatsCache(ctx, result, defaultCacheTTL)
}
