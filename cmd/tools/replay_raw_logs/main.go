package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"api/service"
	"github.com/redis/go-redis/v9"
)

const redisKeyTTL = 45 * 24 * time.Hour

type rawEntry struct {
	Timestamp time.Time
	IP        string
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	UserAgent string
}

func main() {
	logDir := flag.String("log-dir", "./logs", "raw 日志目录")
	startDate := flag.String("start", "", "起始日期 (YYYY-MM-DD)，为空则不限")
	endDate := flag.String("end", "", "结束日期 (YYYY-MM-DD)，为空则不限")
	flag.Parse()

	dates := discoverDates(*logDir)
	if len(dates) == 0 {
		fmt.Println("未找到任何 *_raw.log 文件")
		return
	}

	start, end := parseRange(*startDate, *endDate)

	client, err := service.GetRedisClient()
	if err != nil || client == nil {
		fmt.Printf("Redis 客户端不可用: %v\n", err)
		os.Exit(1)
	}

	for _, date := range dates {
		if !withinRange(date, start, end) {
			continue
		}
		path := filepath.Join(*logDir, fmt.Sprintf("%s_raw.log", date.Format("2006-01-02")))
		if err := replayFile(client, path); err != nil {
			fmt.Printf("重放失败 %s: %v\n", path, err)
			continue
		}
		fmt.Printf("重放完成：%s\n", path)
	}
}

func discoverDates(dir string) []time.Time {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return nil
	}
	dates := make([]time.Time, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, "_raw.log") {
			continue
		}
		dateStr := strings.TrimSuffix(name, "_raw.log")
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			dates = append(dates, t)
		}
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates
}

func parseRange(start, end string) (time.Time, time.Time) {
	var startTime time.Time
	var endTime time.Time
	if start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startTime = t
		}
	}
	if end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endTime = t
		}
	}
	return startTime, endTime
}

func withinRange(date, start, end time.Time) bool {
	if !start.IsZero() && date.Before(start) {
		return false
	}
	if !end.IsZero() && date.After(end) {
		return false
	}
	return true
}

func replayFile(client *redis.Client, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		entry, err := parseRawLine(line)
		if err != nil {
			continue
		}
		if err := writeToRedis(client, entry); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func parseRawLine(line string) (rawEntry, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 6 {
		return rawEntry{}, fmt.Errorf("字段不足")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return rawEntry{}, err
	}
	status, err := strconv.Atoi(parts[4])
	if err != nil {
		return rawEntry{}, err
	}
	latencyMicros, err := strconv.ParseInt(parts[5], 10, 64)
	if err != nil {
		return rawEntry{}, err
	}
	userAgent := ""
	if len(parts) > 6 {
		userAgent = strings.Join(parts[6:], "|")
	}
	return rawEntry{
		Timestamp: ts,
		IP:        parts[1],
		Method:    parts[2],
		Path:      parts[3],
		Status:    status,
		Latency:   time.Duration(latencyMicros) * time.Microsecond,
		UserAgent: userAgent,
	}, nil
}

func writeToRedis(client *redis.Client, entry rawEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dateKey := entry.Timestamp.Format("20060102")
	uvKey := fmt.Sprintf("analytics:uv:%s", dateKey)
	reqKey := fmt.Sprintf("analytics:req:%s", dateKey)
	regionKey := fmt.Sprintf("analytics:region:%s", dateKey)

	pipe := client.Pipeline()
	if entry.IP != "" {
		pipe.SAdd(ctx, uvKey, entry.IP)
		pipe.Expire(ctx, uvKey, redisKeyTTL)
	}
	pipe.Incr(ctx, reqKey)
	pipe.Expire(ctx, reqKey, redisKeyTTL)

	if entry.IP != "" && service.IsChinaIP(entry.IP) {
		region := service.LookupRegion(entry.IP)
		if region != "" {
			pipe.HIncrBy(ctx, regionKey, region, 1)
			pipe.Expire(ctx, regionKey, redisKeyTTL)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}
