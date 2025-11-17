package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	logDirName    = "logs"
	logFileSuffix = "_log"
)

// DailyLogWriter 根据当前日期写入日志文件，文件名形如 2006-01-02_log。
type DailyLogWriter struct {
	mu          sync.Mutex
	dir         string
	currentDate string
	file        *os.File
}

// NewDailyLogWriter 创建一个新的日志写入器。
func NewDailyLogWriter() (*DailyLogWriter, error) {
	dir, err := AccessLogDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir failed: %w", err)
	}
	writer := &DailyLogWriter{dir: dir}
	if err := writer.ensureFile(); err != nil {
		return nil, err
	}
	return writer, nil
}

// Write 实现 io.Writer 接口。
func (w *DailyLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.ensureFile(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

// CurrentPath 返回当前日志文件路径。
func (w *DailyLogWriter) CurrentPath() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.currentDate == "" {
		return ""
	}
	return filepath.Join(w.dir, fmt.Sprintf("%s%s", w.currentDate, logFileSuffix))
}

// Close 关闭当前文件（可选）。
func (w *DailyLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *DailyLogWriter) ensureFile() error {
	date := time.Now().Format("2006-01-02")
	if date == w.currentDate && w.file != nil {
		return nil
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	path := filepath.Join(w.dir, fmt.Sprintf("%s%s", date, logFileSuffix))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file failed: %w", err)
	}
	w.file = file
	w.currentDate = date
	return nil
}

// InitAccessLog 初始化访问日志写入器。
func InitAccessLog() (io.Writer, string, error) {
	writer, err := NewDailyLogWriter()
	if err != nil {
		return nil, "", err
	}
	return writer, writer.CurrentPath(), nil
}

// AccessLogDir 返回日志目录的绝对路径。
func AccessLogDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, logDirName), nil
}

// AccessLogPath 返回当天日志文件路径。
func AccessLogPath() (string, error) {
	return AccessLogPathFor(time.Now())
}

// AccessLogPathFor 返回指定日期的日志文件路径。
func AccessLogPathFor(t time.Time) (string, error) {
	dir, err := AccessLogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s%s", t.Format("2006-01-02"), logFileSuffix)), nil
}
