package service

import (
	"api/dao"
	"api/models"

	"golang.org/x/crypto/bcrypt"
)

// 注册新用户
func RegisterUser(username, email, plainPwd string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{Username: username, Email: email, PasswordHash: string(hash)}
	err = dao.CreateUser(&user)
	return &user, err
}

// 校验登录
func ValidateLogin(identifier, pwd string) (*models.User, error) {
	user, err := dao.GetUserByUsernameOrEmail(identifier, identifier)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pwd)) != nil {
		return nil, bcrypt.ErrMismatchedHashAndPassword
	}
	return user, nil
}

// 查用户详情
func UserDetail(id uint64) (*models.User, error) {
	return dao.GetUserByID(id)
}

// 更新用户
func UpdateUser(user *models.User) error {
	return dao.UpdateUser(user)
}

// 删除用户
func DeleteUser(id uint64) error {
	return dao.DeleteUser(id)
}

// 用户列表
func ListAllUsers() ([]models.User, error) {
	return dao.ListAllUsers()
}

// 修改用户密码（管理员功能）
func ChangeUserPassword(userID uint64, newPassword string) error {
	user, err := dao.GetUserByID(userID)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return dao.UpdateUser(user)
}

// 验证旧密码并修改密码
func ChangePasswordWithValidation(userID uint64, oldPassword, newPassword string) error {
	user, err := dao.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return bcrypt.ErrMismatchedHashAndPassword
	}

	// 生成新密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return dao.UpdateUser(user)
}
