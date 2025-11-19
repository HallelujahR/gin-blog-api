package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type legacyEntry struct {
	Timestamp time.Time
	IP        string
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
}

func main() {
	logDir := flag.String("log-dir", "./logs", "旧日志所在目录")
	overwrite := flag.Bool("overwrite", false, "若目标 _raw.log 已存在是否覆盖")
	flag.Parse()

	entries, err := os.ReadDir(*logDir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, "_log") || strings.HasSuffix(name, "_raw.log") {
			continue
		}

		date := strings.TrimSuffix(name, "_log")
		src := filepath.Join(*logDir, name)
		dest := filepath.Join(*logDir, fmt.Sprintf("%s_raw.log", date))

		if !*overwrite {
			if _, err := os.Stat(dest); err == nil {
				fmt.Printf("目标文件已存在，跳过：%s\n", dest)
				continue
			}
		}

		if err := exportFile(src, dest); err != nil {
			fmt.Printf("转换失败 %s -> %s : %v\n", src, dest, err)
			continue
		}
		fmt.Printf("转换完成：%s -> %s\n", src, dest)
	}
}

func exportFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	scanner := bufio.NewScanner(in)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		entry, err := parseLegacyLine(line)
		if err != nil {
			continue
		}
		fmt.Fprintf(writer, "%s|%s|%s|%s|%d|%d|-\n",
			entry.Timestamp.Format(time.RFC3339Nano),
			entry.IP,
			entry.Method,
			entry.Path,
			entry.Status,
			entry.Latency.Microseconds(),
		)
	}

	return scanner.Err()
}

func parseLegacyLine(line string) (legacyEntry, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return legacyEntry{}, fmt.Errorf("空行")
	}
	if strings.HasPrefix(line, "[GIN]") {
		return parseGinLine(line)
	}
	return parseCustomLine(line)
}

func parseCustomLine(line string) (legacyEntry, error) {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return legacyEntry{}, fmt.Errorf("字段不足")
	}

	ts, err := time.Parse(time.RFC3339, fields[0])
	if err != nil {
		return legacyEntry{}, err
	}

	status, err := strconv.Atoi(fields[4])
	if err != nil {
		return legacyEntry{}, err
	}
	latency, err := time.ParseDuration(fields[5])
	if err != nil {
		return legacyEntry{}, err
	}

	return legacyEntry{
		Timestamp: ts,
		IP:        fields[1],
		Method:    fields[2],
		Path:      fields[3],
		Status:    status,
		Latency:   latency,
	}, nil
}

func parseGinLine(line string) (legacyEntry, error) {
	sections := strings.Split(line, "|")
	if len(sections) < 5 {
		return legacyEntry{}, fmt.Errorf("Gin 日志格式不正确")
	}

	tsPart := strings.TrimSpace(strings.TrimPrefix(sections[0], "[GIN]"))
	statusStr := strings.TrimSpace(sections[1])
	latencyStr := strings.TrimSpace(sections[2])
	ip := strings.TrimSpace(sections[3])
	methodPath := strings.TrimSpace(sections[4])

	ts, err := time.ParseInLocation("2006/01/02 - 15:04:05", tsPart, time.Local)
	if err != nil {
		return legacyEntry{}, err
	}
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		return legacyEntry{}, err
	}
	latency, err := time.ParseDuration(strings.ReplaceAll(latencyStr, " ", ""))
	if err != nil {
		return legacyEntry{}, err
	}

	method := ""
	path := ""
	if idx := strings.Index(methodPath, "\""); idx >= 0 {
		method = strings.TrimSpace(methodPath[:idx])
		path = strings.Trim(strings.TrimSpace(methodPath[idx:]), "\"")
	} else {
		parts := strings.Fields(methodPath)
		if len(parts) > 0 {
			method = parts[0]
		}
		if len(parts) > 1 {
			path = parts[1]
		}
	}

	return legacyEntry{
		Timestamp: ts.UTC(),
		IP:        ip,
		Method:    method,
		Path:      path,
		Status:    status,
		Latency:   latency,
	}, nil
}
