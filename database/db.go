package database

import (
	"api/configs"
	"database/sql"
	"fmt"
	"sync"

	"gorm.io/gorm/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var once sync.Once

// GetDB 获取全局唯一的DB实例
func GetDB() *gorm.DB {
	once.Do(func() {
		cfg := configs.Load()
		dsn := getDSN(cfg)
		var err error
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(cfg.DBLogLevel),
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

		sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

		// 验证连接是否可用
		if err := sqlDB.Ping(); err != nil {
			panic("数据库连接验证失败: " + err.Error())
		}
	})
	return db
}

// getDSN 根据环境变量切换本地/生产数据库配置
func getDSN(cfg configs.AppConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=15s&writeTimeout=15s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
}
