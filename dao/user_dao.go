package dao

import (
	"api/database"
	"api/models"
	"database/sql"
	"errors"
	"time"
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

// 用户列表
func ListAllUsers() ([]models.User, error) {
	var users []models.User
	err := database.GetDB().Find(&users).Error
	return users, err
}

func validateDate(date string) (string, error) {
	if date == "0000-00-00" || date == "" {
		return "", errors.New("invalid date value")
	}
	return date, nil
}

func validateRole(role string) (string, error) {
	validRoles := map[string]bool{
		"admin": true,
		"user":  true,
		"guest": true,
	}
	if !validRoles[role] {
		return "", errors.New("invalid role value")
	}
	return role, nil
}

func InsertUser(db *sql.DB, username, role, createdAt string) error {
	// Validate 'created_at'
	validDate, err := validateDate(createdAt)
	if err != nil || validDate == "" {
		validDate = time.Now().Format("2006-01-02") // Use current date as default
	}

	// Validate 'role'
	validRole, err := validateRole(role)
	if err != nil {
		return err
	}

	query := "INSERT INTO users (username, role, created_at) VALUES (?, ?, ?)"
	_, err = db.Exec(query, username, validRole, validDate)
	if err != nil {
		return err
	}
	return nil
}
