package initializer

import (
	"api/database"
	"api/models"
)

func InitDB() {
	db := database.GetDB()
	migrations := []interface{}{
		&models.User{},
		&models.UserSession{},
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
	}

	if !db.Migrator().HasTable(&models.Category{}) {
		migrations = append([]interface{}{&models.Category{}}, migrations...)
	}

	for _, model := range migrations {
		if err := db.AutoMigrate(model); err != nil {
			panic("数据库自动迁移失败: " + err.Error())
		}
	}
}
