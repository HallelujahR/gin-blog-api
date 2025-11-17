package service

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisOnce   sync.Once
	redisClient *redis.Client
	redisErr    error
)

// GetRedisClient 返回全局 Redis 客户端，使用懒加载方式初始化。
// 默认连接 127.0.0.1:6379，可通过 REDIS_ADDR、REDIS_PASSWORD、REDIS_DB 配置。
const (
	defaultRedisAddr         = "host.docker.internal:6379"
	defaultRedisFallbackAddr = "127.0.0.1:6379"
)

func GetRedisClient() (*redis.Client, error) {
	redisOnce.Do(func() {
		addr := defaultRedisAddr

		redisClient = redis.NewClient(&redis.Options{
			Addr: addr,
			DB:   0,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			redisClient = redis.NewClient(&redis.Options{
				Addr: defaultRedisFallbackAddr,
				DB:   0,
			})
			redisErr = redisClient.Ping(ctx).Err()
		} else {
			redisErr = nil
		}
	})

	return redisClient, redisErr
}
