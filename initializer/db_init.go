package initializer

import (
	"api/database"
	"api/models"
)

func InitDB() {
	db := database.GetDB()
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
		panic("数据库自动迁移失败: " + err.Error())
	}
}
