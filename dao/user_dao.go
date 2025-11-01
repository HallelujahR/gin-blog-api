package dao

import (
	"api/database"
	"api/models"
)

// 创建用户
func CreateUser(user *models.User) error {
	return database.GetDB().Create(user).Error
}

// 通过ID查找用户
func GetUserByID(id uint64) (*models.User, error) {
	var user models.User
	err := database.GetDB().First(&user, id).Error
	return &user, err
}

// 通过用户名或邮箱查找
func GetUserByUsernameOrEmail(username, email string) (*models.User, error) {
	var user models.User
	db := database.GetDB()
	if username != "" {
		err := db.Where("username = ?", username).First(&user).Error
		return &user, err
	}
	err := db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// 更新用户
func UpdateUser(user *models.User) error {
	return database.GetDB().Save(user).Error
}

// 删除用户
func DeleteUser(id uint64) error {
	return database.GetDB().Delete(&models.User{}, id).Error
}
