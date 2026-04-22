package configs

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm/logger"
)

type AppConfig struct {
	Env string

	BaseURL string

	HTTPPort     string
	PprofEnabled bool
	PprofPort    string

	AnalyticsEnabled bool
	AnalyticsETL     time.Duration
	AccessLogEnabled bool

	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
	DBLogLevel        logger.LogLevel

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisEnabled  bool
}

var (
	cfg     AppConfig
	cfgOnce sync.Once
)

func Load() AppConfig {
	cfgOnce.Do(func() {
		cfg = AppConfig{
			Env:              envString("BLOG_ENV", "dev"),
			BaseURL:          envString("API_BASE_URL", "http://localhost:8080"),
			HTTPPort:         envString("HTTP_PORT", "8080"),
			PprofEnabled:     envBool("ENABLE_PPROF", false),
			PprofPort:        envString("PPROF_PORT", "6060"),
			AnalyticsEnabled: envBool("ENABLE_ANALYTICS", false),
			AnalyticsETL:     envDuration("ANALYTICS_ETL_INTERVAL", 30*time.Minute),
			AccessLogEnabled: envBool("ENABLE_ACCESS_LOG", true),
			DBHost:           envString("DB_HOST", "127.0.0.1"),
			DBPort:           envString("DB_PORT", "3306"),
			DBUser:           envString("DB_USER", defaultDBUser()),
			DBPassword:       envString("DB_PASSWORD", defaultDBPassword()),
			DBName:           envString("DB_NAME", "blog"),
			DBMaxOpenConns:   envInt("DB_MAX_OPEN_CONNS", 8),
			DBMaxIdleConns:   envInt("DB_MAX_IDLE_CONNS", 3),
			DBConnMaxLifetime: envDuration(
				"DB_CONN_MAX_LIFETIME",
				30*time.Minute,
			),
			DBConnMaxIdleTime: envDuration(
				"DB_CONN_MAX_IDLE_TIME",
				5*time.Minute,
			),
			DBLogLevel:    envLogLevel("DB_LOG_LEVEL", logger.Warn),
			RedisAddr:     envString("REDIS_ADDR", "127.0.0.1:6379"),
			RedisPassword: envString("REDIS_PASSWORD", ""),
			RedisDB:       envInt("REDIS_DB", 0),
			RedisEnabled:  envBool("ENABLE_REDIS", true),
		}
	})
	return cfg
}

func GetBaseURL() string {
	return Load().BaseURL
}

func SetBaseURL(url string) {
	url = strings.TrimSpace(url)
	if url == "" {
		return
	}
	cfgOnce.Do(func() {
		cfg = Load()
	})
	cfg.BaseURL = url
}

func defaultDBUser() string {
	if strings.EqualFold(envString("BLOG_ENV", "dev"), "prod") {
		return "blog_user"
	}
	return "root"
}

func defaultDBPassword() string {
	if strings.EqualFold(envString("BLOG_ENV", "dev"), "prod") {
		return "blog_password"
	}
	return "10244201"
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envLogLevel(key string, fallback logger.LogLevel) logger.LogLevel {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn", "warning":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return fallback
	}
}
