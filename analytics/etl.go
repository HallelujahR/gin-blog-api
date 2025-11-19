package analytics

import (
	"api/dao"
	"api/service"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	etlInterval = 5 * time.Minute
)

var etlOnce sync.Once

// StartETLWorker 启动定时任务，周期性地将 Redis 中的实时指标固化到数据库快照表，用于统计分析。
func StartETLWorker() {
	etlOnce.Do(func() {
		go runETL()
	})
}

func runETL() {
	ticker := time.NewTicker(etlInterval)
	defer ticker.Stop()

	for {
		if err := flushSnapshot(); err != nil {
			fmt.Printf("[analytics] flush snapshot error: %v\n", err)
		}
		<-ticker.C
	}
}

func flushSnapshot() error {
	client, err := service.GetRedisClient()
	if err != nil {
		return err
	}
	if client == nil {
		return fmt.Errorf("redis client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	regionCounts := make(map[string]int)
	uniqueIPSet := make(map[string]struct{})

	now := time.Now()
	for i := 0; i < defaultLookbackDays; i++ {
		date := now.AddDate(0, 0, -i).Format("20060102")
		uvKey := fmt.Sprintf("analytics:uv:%s", date)
		regionKey := fmt.Sprintf("analytics:region:%s", date)

		if members, err := client.SMembers(ctx, uvKey).Result(); err == nil {
			for _, ip := range members {
				if ip == "" {
					continue
				}
				uniqueIPSet[ip] = struct{}{}
			}
		}

		if regionMap, err := client.HGetAll(ctx, regionKey).Result(); err == nil {
			for region, countStr := range regionMap {
				if region == "" {
					continue
				}
				count, convErr := strconv.Atoi(countStr)
				if convErr != nil {
					continue
				}
				regionCounts[region] += count
			}
		}
	}

	dto := dao.TrafficSnapshotDTO{
		LookbackDays:   defaultLookbackDays,
		UniqueVisitors: len(uniqueIPSet),
		RegionCounts:   regionCounts,
		GeneratedAt:    time.Now(),
	}
	return dao.SaveTrafficSnapshot(dto)
}
