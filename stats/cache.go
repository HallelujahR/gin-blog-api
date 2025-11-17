package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"api/service"

	"github.com/redis/go-redis/v9"
)

const (
	statsCacheKey = "stats:summary"
	localCacheTTL = 24 * time.Hour
)

var (
	localCache       StatsResult
	localCacheExpiry time.Time
	localCacheMu     sync.RWMutex
)

// GetStatsCache 从 Redis 获取缓存的统计结果，失败时退回本地缓存。
func GetStatsCache(ctx context.Context) (StatsResult, bool) {
	client, err := service.GetRedisClient()
	if err != nil {
		fmt.Printf("[stats] redis client error: %v\n", err)
		return getLocalCache()
	}

	value, err := client.Get(ctx, statsCacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			fmt.Println("[stats] redis cache miss, fallback to local cache")
		} else {
			fmt.Printf("[stats] redis get error: %v\n", err)
		}
		return getLocalCache()
	}

	var result StatsResult
	if err := json.Unmarshal(value, &result); err != nil {
		fmt.Printf("[stats] redis cache decode error: %v\n", err)
		return getLocalCache()
	}
	updateLocalCache(result, localCacheTTL)
	return result, true
}

// SetStatsCache 将统计结果写入 Redis，并更新本地缓存。
func SetStatsCache(ctx context.Context, data StatsResult, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	client, clientErr := service.GetRedisClient()
	if clientErr != nil {
		fmt.Printf("[stats] redis client error when set: %v\n", clientErr)
	} else if err := client.Set(ctx, statsCacheKey, payload, ttl).Err(); err != nil {
		fmt.Printf("[stats] redis set error: %v\n", err)
		clientErr = err
	}

	updateLocalCache(data, ttl)
	return clientErr
}

func getLocalCache() (StatsResult, bool) {
	localCacheMu.RLock()
	defer localCacheMu.RUnlock()
	if time.Now().Before(localCacheExpiry) && !localCache.GeneratedAt.IsZero() {
		return localCache, true
	}
	return StatsResult{}, false
}

func updateLocalCache(data StatsResult, ttl time.Duration) {
	localCacheMu.Lock()
	defer localCacheMu.Unlock()
	localCache = data
	localCacheExpiry = time.Now().Add(ttl)
}
