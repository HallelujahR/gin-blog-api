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

	// 生成30天快照
	regionCounts30 := make(map[string]int)
	uniqueIPSet30 := make(map[string]struct{})

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
				uniqueIPSet30[ip] = struct{}{}
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
				regionCounts30[region] += count
			}
		}
	}

	dto30 := dao.TrafficSnapshotDTO{
		LookbackDays:   defaultLookbackDays,
		UniqueVisitors: len(uniqueIPSet30),
		RegionCounts:   regionCounts30,
		GeneratedAt:    time.Now(),
	}
	if err := dao.SaveTrafficSnapshot(dto30); err != nil {
		return err
	}

	// 生成所有历史快照（从Redis中获取所有可用的日期键）
	// 注意：由于Redis键的TTL是45天，所以这里最多只能获取45天的数据
	regionCountsAll := make(map[string]int)
	uniqueIPSetAll := make(map[string]struct{})

	// 使用SCAN遍历所有analytics:uv:*键（更安全，不会阻塞Redis）
	var cursor uint64
	uvPattern := "analytics:uv:*"
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, uvPattern, 100).Result()
		if err != nil {
			break
		}
		for _, uvKey := range keys {
			if members, err := client.SMembers(ctx, uvKey).Result(); err == nil {
				for _, ip := range members {
					if ip == "" {
						continue
					}
					uniqueIPSetAll[ip] = struct{}{}
				}
			}
		}
		if cursor == 0 {
			break
		}
	}

	// 使用SCAN遍历所有analytics:region:*键
	cursor = 0
	regionPattern := "analytics:region:*"
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, regionPattern, 100).Result()
		if err != nil {
			break
		}
		for _, regionKey := range keys {
			if regionMap, err := client.HGetAll(ctx, regionKey).Result(); err == nil {
				for region, countStr := range regionMap {
					if region == "" {
						continue
					}
					count, convErr := strconv.Atoi(countStr)
					if convErr != nil {
						continue
					}
					regionCountsAll[region] += count
				}
			}
		}
		if cursor == 0 {
			break
		}
	}

	// 保存所有历史快照（使用0表示所有历史）
	dtoAll := dao.TrafficSnapshotDTO{
		LookbackDays:   0,
		UniqueVisitors: len(uniqueIPSetAll),
		RegionCounts:   regionCountsAll,
		GeneratedAt:    time.Now(),
	}
	return dao.SaveTrafficSnapshot(dtoAll)
}
