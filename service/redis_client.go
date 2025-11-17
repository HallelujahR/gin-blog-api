package service

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisOnce   sync.Once
	redisClient *redis.Client
	redisErr    error
)

const (
	defaultRedisAddr = "127.0.0.1:6379"
)

// GetRedisClient 返回全局 Redis 客户端，使用懒加载方式初始化。
// 默认连接本机 127.0.0.1:6379，可通过 REDIS_ADDR、REDIS_PASSWORD、REDIS_DB 覆盖。

func GetRedisClient() (*redis.Client, error) {
	redisOnce.Do(func() {
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = defaultRedisAddr
		}
		password := os.Getenv("REDIS_PASSWORD")

		db := 0
		if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
			if parsed, err := parseRedisDB(dbStr); err == nil {
				db = parsed
			}
		}

		redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			DB:       db,
			Password: password,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		redisErr = redisClient.Ping(ctx).Err()
	})

	return redisClient, redisErr
}

func parseRedisDB(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, err
	}
	return parsed, nil
}
