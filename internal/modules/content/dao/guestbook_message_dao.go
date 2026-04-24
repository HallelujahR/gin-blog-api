package dao

import (
	"api/internal/modules/content/models"
	"api/internal/platform/db"
)

func CreateGuestbookMessage(message *models.GuestbookMessage) error {
	return database.GetDB().Create(message).Error
}

func GetGuestbookMessageByID(id uint64) (*models.GuestbookMessage, error) {
	var message models.GuestbookMessage
	err := database.GetDB().First(&message, id).Error
	return &message, err
}

func UpdateGuestbookMessage(message *models.GuestbookMessage) error {
	return database.GetDB().Save(message).Error
}

func DeleteGuestbookMessage(id uint64) error {
	return database.GetDB().Delete(&models.GuestbookMessage{}, id).Error
}

func CountGuestbookMessages(status, q string) (int64, error) {
	var count int64
	db := database.GetDB().Model(&models.GuestbookMessage{})

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if q != "" {
		db = db.Where(
			"content LIKE ? OR author_name LIKE ? OR author_email LIKE ?",
			"%"+q+"%", "%"+q+"%", "%"+q+"%",
		)
	}

	err := db.Count(&count).Error
	return count, err
}

func ListGuestbookMessagesWithParams(page, pageSize int, status, q, sort string) ([]models.GuestbookMessage, error) {
	var messages []models.GuestbookMessage
	db := database.GetDB().Model(&models.GuestbookMessage{})

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if q != "" {
		db = db.Where(
			"content LIKE ? OR author_name LIKE ? OR author_email LIKE ?",
			"%"+q+"%", "%"+q+"%", "%"+q+"%",
		)
	}

	if sort != "" {
		db = db.Order("created_at " + sort)
	} else {
		db = db.Order("created_at DESC")
	}

	if pageSize > 0 {
		db = db.Limit(pageSize).Offset((page - 1) * pageSize)
	}

	err := db.Find(&messages).Error
	return messages, err
}
