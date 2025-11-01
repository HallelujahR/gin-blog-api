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
	//if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pwd)) != nil {
	//	return nil, bcrypt.ErrMismatchedHashAndPassword
	//}
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
