package stats

import (
	"context"
	"encoding/json"
	"time"

	"api/service"

	"github.com/redis/go-redis/v9"
)

const statsCacheKey = "stats:summary"

// GetStatsCache 从 Redis 获取缓存的统计结果。
func GetStatsCache(ctx context.Context) (StatsResult, bool) {
	client, err := service.GetRedisClient()
	if err != nil {
		return StatsResult{}, false
	}

	value, err := client.Get(ctx, statsCacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return StatsResult{}, false
		}
		return StatsResult{}, false
	}

	var result StatsResult
	if err := json.Unmarshal(value, &result); err != nil {
		return StatsResult{}, false
	}
	return result, true
}

// SetStatsCache 将统计结果写入 Redis，并设置 TTL。
func SetStatsCache(ctx context.Context, data StatsResult, ttl time.Duration) error {
	client, err := service.GetRedisClient()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return client.Set(ctx, statsCacheKey, payload, ttl).Err()
}
