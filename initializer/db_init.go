package initializer

import (
	"api/database"
	"api/models"
)

func InitDB() {
	db := database.GetDB()
	db.AutoMigrate(
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
		&models.PostViewStat{},
		&models.TrafficSnapshot{},
	)
}
