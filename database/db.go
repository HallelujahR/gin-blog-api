package database

import (
	"fmt"
	"gorm.io/gorm/logger"
	"os"
	"sync"

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
			dbHost = "mysql"
		}
		if dbPort == "" {
			dbPort = "3306"
		}
		if dbUser == "" {
			dbUser = "blog_user"
		}
		if dbName == "" {
			dbName = "blog"
		}
		
		// 构建DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPassword, dbHost, dbPort, dbName)
		return dsn
	}
	// 默认本地开发环境配置
	return "root:10244201@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local"
}
