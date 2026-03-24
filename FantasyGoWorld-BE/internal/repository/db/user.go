package db

import (
	"fantasy-go-world-be/internal/repository/model"
)

// CreateUser 创建新用户
func CreateUser(user *model.User) error {
	return DB.Create(user).Error
}

// GetUserByUsername 根据用户名查询用户
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据 ID 查询用户
func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetOnlineList 获取大厅在线用户列表 (带分页)
// @stub: 阶段三暂使用全量用户列表，后续阶段四接入 Redis 在线状态过滤
func GetOnlineList(offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	if err := DB.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := DB.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
