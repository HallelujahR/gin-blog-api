package stats

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	errEmptyPath      = errors.New("log path is empty")
	errFileNotRegular = errors.New("log path is not a regular file")
)

// ParseLogFile 解析访问日志文件并返回 LogEntry 列表。
// 支持两种格式：
// 1) 自定义格式：timestamp ip method path status duration region
// 2) Gin 默认格式：[GIN] 2006/01/02 - 15:04:05 | status | latency | ip | method "path"
// 如果日志不存在则返回空数组，便于上层降级处理。
func ParseLogFile(ctx context.Context, path string) ([]LogEntry, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errEmptyPath
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []LogEntry{}, nil
		}
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errFileNotRegular
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lineCh := make(chan string, 1024)
	entryCh := make(chan LogEntry, 1024)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	workerCount := runtime.NumCPU()
	if workerCount < 2 {
		workerCount = 2
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for line := range lineCh {
				entry, parseErr := parseLine(line)
				if parseErr != nil {
					continue
				}
				select {
				case entryCh <- entry:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case lineCh <- line:
			case <-ctx.Done():
				close(lineCh)
				return
			}
		}
		if scanErr := scanner.Err(); scanErr != nil && !errors.Is(scanErr, io.EOF) {
			errCh <- scanErr
		}
		close(lineCh)
	}()

	go func() {
		wg.Wait()
		close(entryCh)
	}()

	entries := make([]LogEntry, 0, 1024)
	for {
		select {
		case entry, ok := <-entryCh:
			if !ok {
				return entries, nil
			}
			entries = append(entries, entry)
		case workerErr := <-errCh:
			return entries, workerErr
		case <-ctx.Done():
			return entries, ctx.Err()
		}
	}
}

// parseLine 尝试匹配多种日志格式。
func parseLine(line string) (LogEntry, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return LogEntry{}, errors.New("empty line")
	}

	if strings.HasPrefix(line, "[GIN]") {
		return parseGinLine(line)
	}

	fields := strings.Fields(line)
	if len(fields) < 7 {
		return LogEntry{}, errors.New("invalid log line")
	}

	ts, err := time.Parse(time.RFC3339, fields[0])
	if err != nil {
		return LogEntry{}, err
	}

	ip := fields[1]
	method := fields[2]
	path := fields[3]

	status, err := strconv.Atoi(fields[4])
	if err != nil {
		return LogEntry{}, err
	}

	duration, err := time.ParseDuration(fields[5])
	if err != nil {
		return LogEntry{}, err
	}

	region := fields[6]

	return LogEntry{
		Timestamp:  ts,
		IP:         ip,
		Method:     method,
		Path:       path,
		Status:     status,
		DurationMs: duration.Milliseconds(),
		Region:     region,
	}, nil
}

// parseGinLine 解析 Gin 默认格式的日志行。
func parseGinLine(line string) (LogEntry, error) {
	sections := strings.Split(line, "|")
	if len(sections) < 5 {
		return LogEntry{}, errors.New("invalid gin log line")
	}

	tsPart := strings.TrimSpace(strings.TrimPrefix(sections[0], "[GIN]"))
	statusStr := strings.TrimSpace(sections[1])
	latencyStr := strings.TrimSpace(sections[2])
	ip := strings.TrimSpace(sections[3])
	methodPath := strings.TrimSpace(sections[4])

	ts, err := time.ParseInLocation("2006/01/02 - 15:04:05", tsPart, time.Local)
	if err != nil {
		return LogEntry{}, err
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		return LogEntry{}, err
	}

	duration, err := time.ParseDuration(strings.ReplaceAll(latencyStr, " ", ""))
	if err != nil {
		return LogEntry{}, err
	}

	method := ""
	path := ""
	if idx := strings.Index(methodPath, "\""); idx >= 0 {
		method = strings.TrimSpace(methodPath[:idx])
		path = strings.TrimSpace(strings.Trim(methodPath[idx:], "\""))
	} else {
		parts := strings.Fields(methodPath)
		if len(parts) > 0 {
			method = parts[0]
		}
		if len(parts) > 1 {
			path = parts[1]
		}
	}

	return LogEntry{
		Timestamp:  ts.UTC(),
		IP:         ip,
		Method:     method,
		Path:       path,
		Status:     status,
		DurationMs: duration.Milliseconds(),
		Region:     "",
	}, nil
}
