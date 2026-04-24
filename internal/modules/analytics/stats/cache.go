package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"api/internal/platform/redisstore"

	"github.com/redis/go-redis/v9"
)

const statsCacheKey = "stats:summary"

// GetStatsCache 从 Redis 获取缓存的统计结果。
func GetStatsCache(ctx context.Context) (StatsResult, bool) {
	client, err := redisstore.GetClient()
	if err != nil {
		fmt.Printf("[stats] redis client error: %v\n", err)
		return StatsResult{}, false
	}
	if client == nil {
		return StatsResult{}, false
	}

	value, err := client.Get(ctx, statsCacheKey).Bytes()
	if err != nil {
		if err != redis.Nil {
			fmt.Printf("[stats] redis get error: %v\n", err)
		}
		return StatsResult{}, false
	}

	var result StatsResult
	if err := json.Unmarshal(value, &result); err != nil {
		fmt.Printf("[stats] redis cache decode error: %v\n", err)
		return StatsResult{}, false
	}
	return result, true
}

// SetStatsCache 将统计结果写入 Redis。
func SetStatsCache(ctx context.Context, data StatsResult, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	client, clientErr := redisstore.GetClient()
	if clientErr != nil {
		fmt.Printf("[stats] redis client error when set: %v\n", clientErr)
		return clientErr
	}
	if client == nil {
		return nil
	}
	if err := client.Set(ctx, statsCacheKey, payload, ttl).Err(); err != nil {
		fmt.Printf("[stats] redis set error: %v\n", err)
		return err
	}
	return nil
}
