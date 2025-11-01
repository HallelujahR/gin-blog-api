package dao

import (
	"api/database"
	"api/models"
)

// 创建会话
func CreateSession(session *models.UserSession) (*models.UserSession, error) {
	err := database.GetDB().Create(session).Error
	return session, err
}

// 通过token查找会话
func GetSessionByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	err := database.GetDB().Where("session_token = ?", token).First(&session).Error
	return &session, err
}

// 删除会话
func DeleteSession(token string) error {
	return database.GetDB().Where("session_token = ?", token).Delete(&models.UserSession{}).Error
}

// 删除用户的所有会话
func DeleteAllUserSessions(userID uint64) error {
	return database.GetDB().Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}
