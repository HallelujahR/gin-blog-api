package service

import (
	"api/configs"
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
// 默认连接本机 127.0.0.1:6379，可通过 REDIS_ADDR、REDIS_PASSWORD、REDIS_DB 覆盖。

func GetRedisClient() (*redis.Client, error) {
	redisOnce.Do(func() {
		cfg := configs.Load()
		if !cfg.RedisEnabled {
			redisClient = nil
			redisErr = nil
			return
		}
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			DB:       cfg.RedisDB,
			Password: cfg.RedisPassword,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		redisErr = redisClient.Ping(ctx).Err()
	})

	return redisClient, redisErr
}
