package analytics

import (
	"api/service"
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	defaultLookbackDays = 30
	eventBufferSize     = 2048
	redisKeyTTL         = 45 * 24 * time.Hour
	rawLogSuffix        = "_raw.log"
)

var (
	initOnce    sync.Once
	eventCh     chan Event
	rawWriter   *service.DailyLogWriter
	rawWriterMu sync.Mutex
)

// Capture 异步接收一条请求事件，将其放入队列供后台写入器消费。
func Capture(event Event) {
	initOnce.Do(func() {
		eventCh = make(chan Event, eventBufferSize)
		go consumeEvents()
	})

	select {
	case eventCh <- event:
	default:
		fmt.Printf("[analytics] event channel full, drop request %s %s\n", event.Method, event.Path)
	}
}

func consumeEvents() {
	for evt := range eventCh {
		if err := writeRaw(evt); err != nil {
			fmt.Printf("[analytics] write raw log failed: %v\n", err)
		}
		if err := cacheEvent(evt); err != nil {
			fmt.Printf("[analytics] cache event failed: %v\n", err)
		}
	}
}

func writeRaw(event Event) error {
	rawWriterMu.Lock()
	defer rawWriterMu.Unlock()
	if rawWriter == nil {
		writer, err := service.NewDailyLogWriterWithSuffix(rawLogSuffix)
		if err != nil {
			return err
		}
		rawWriter = writer
	}
	line := fmt.Sprintf("%s|%s|%s|%s|%d|%d|%s\n",
		event.Timestamp.Format(time.RFC3339Nano),
		event.IP,
		event.Method,
		event.Path,
		event.Status,
		event.Latency.Microseconds(),
		event.UserAgent,
	)
	_, err := rawWriter.Write([]byte(line))
	return err
}

func cacheEvent(event Event) error {
	client, err := service.GetRedisClient()
	if err != nil || client == nil {
		if err != nil {
			return err
		}
		return fmt.Errorf("redis client unavailable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	dateKey := event.Timestamp.Format("20060102")
	uvKey := fmt.Sprintf("analytics:uv:%s", dateKey)
	reqKey := fmt.Sprintf("analytics:req:%s", dateKey)
	regionKey := fmt.Sprintf("analytics:region:%s", dateKey)

	pipe := client.Pipeline()
	if event.IP != "" {
		pipe.SAdd(ctx, uvKey, event.IP)
		pipe.Expire(ctx, uvKey, redisKeyTTL)
	}
	pipe.Incr(ctx, reqKey)
	pipe.Expire(ctx, reqKey, redisKeyTTL)

	if event.IP != "" && service.IsChinaIP(event.IP) {
		region := service.LookupRegion(event.IP)
		if region != "" {
			pipe.HIncrBy(ctx, regionKey, region, 1)
			pipe.Expire(ctx, regionKey, redisKeyTTL)
		}
	}

	_, execErr := pipe.Exec(ctx)
	return execErr
}
