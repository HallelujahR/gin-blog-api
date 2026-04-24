package main

import (
	"api/internal/modules/content/models"
	"api/internal/platform/db"

	"gorm.io/gorm"
)

func InitDB() {
	db := database.GetDB()
	ensureTable(db, &models.User{})
	ensureTable(db, &models.UserSession{})
	ensureTable(db, &models.Category{})
	ensureTable(db, &models.Tag{})
	ensureTable(db, &models.Post{})
	ensureTable(db, &models.PostCategory{})
	ensureTable(db, &models.PostTag{})
	ensureTable(db, &models.Comment{})
	ensureTable(db, &models.Like{})
	ensureTable(db, &models.HotData{})
	ensureTable(db, &models.Page{})
	ensureTable(db, &models.GuestbookMessage{})
	ensureTable(db, &models.Moment{})
	ensureTable(db, &models.PostViewStat{})
	ensureTable(db, &models.TrafficSnapshot{})
	ensureTable(db, &models.ImageCompressStats{})
	ensureTable(db, &models.ImageCompressJob{})
}

func ensureTable(db *gorm.DB, model interface{}) {
	if db.Migrator().HasTable(model) {
		return
	}
	if err := db.AutoMigrate(model); err != nil {
		panic("数据库自动建表失败: " + err.Error())
	}
}
