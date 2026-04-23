package initializer

import (
	"api/database"
	"api/models"
)

func InitDB() {
	db := database.GetDB()
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		panic("关闭外键检查失败: " + err.Error())
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.Category{},
		&models.Tag{},
		&models.Post{},
		&models.PostCategory{},
		&models.PostTag{},
		&models.Comment{},
		&models.Like{},
		&models.HotData{},
		&models.Page{},
		&models.GuestbookMessage{},
		&models.PostViewStat{},
		&models.TrafficSnapshot{},
		&models.ImageCompressStats{},
		&models.ImageCompressJob{},
	); err != nil {
		_ = db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error
		panic("数据库自动迁移失败: " + err.Error())
	}

	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		panic("恢复外键检查失败: " + err.Error())
	}
}
