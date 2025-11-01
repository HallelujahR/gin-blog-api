package database

import (
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
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
		// 生产环境配置(请替换为实际生产参数)
		return "prod_user:prod_password@tcp(prod_host:3306)/prod_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	// 默认本地开发环境配置
	return "root:10244201@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local"
}
