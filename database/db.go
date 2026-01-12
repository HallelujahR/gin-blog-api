package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"gorm.io/gorm/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var once sync.Once

// GetDB 获取全局唯一的DB实例
func GetDB() *gorm.DB {
	once.Do(func() {
		dsn := getDSN()
		var err error
		//db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			panic("数据库连接失败: " + err.Error())
		}

		// 配置数据库连接池参数以优化性能
		var sqlDB *sql.DB
		sqlDB, err = db.DB()
		if err != nil {
			panic("获取底层数据库连接失败: " + err.Error())
		}

		// 设置最大打开连接数（根据实际负载调整）
		sqlDB.SetMaxOpenConns(25)
		// 设置最大空闲连接数
		sqlDB.SetMaxIdleConns(10)
		// 设置连接的最大生命周期（避免长时间连接导致的问题）
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		// 设置连接的最大空闲时间
		sqlDB.SetConnMaxIdleTime(10 * time.Minute)

		// 验证连接是否可用
		if err := sqlDB.Ping(); err != nil {
			panic("数据库连接验证失败: " + err.Error())
		}
	})
	return db
}

// getDSN 根据环境变量切换本地/生产数据库配置
func getDSN() string {
	env := os.Getenv("BLOG_ENV") // BLOG_ENV 设为 prod 表示生产环境，默认是开发环境
	if env == "prod" {
		// 生产环境：从环境变量读取数据库配置
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		// 设置默认值
		if dbHost == "" {
			dbHost = "127.0.0.1"
		}
		if dbPort == "" {
			dbPort = "3306"
		}
		if dbUser == "" {
			dbUser = "blog_user"
		}
		if dbPassword == "" {
			dbPassword = "blog_password"
		}
		if dbName == "" {
			dbName = "blog"
		}

		// 构建DSN，添加性能优化参数
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
			dbUser, dbPassword, dbHost, dbPort, dbName)
		return dsn
	}
	// 默认本地开发环境配置，添加性能优化参数
	return "root:10244201@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s"
}
