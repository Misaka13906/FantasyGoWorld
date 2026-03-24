package db

import (
	"fantasy-go-world-be/internal/repository/model"
)

// CreateRoom 创建房间
func CreateRoom(room *model.Room) error {
	return DB.Create(room).Error
}

// GetRoomByID 根据 ID 查询房间 (预加载房主信息)
func GetRoomByID(id uint) (*model.Room, error) {
	var room model.Room
	if err := DB.Preload("Owner").First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

// ListPublicRooms 获取公开房间列表 (带分页)
func ListPublicRooms(offset, limit int) ([]model.Room, int64, error) {
	var rooms []model.Room
	var total int64

	query := DB.Model(&model.Room{}).Where("is_public = ?", true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Owner").Order("created_at DESC").Offset(offset).Limit(limit).Find(&rooms).Error; err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

// DeleteRoom 删除房间
func DeleteRoom(id uint) error {
	return DB.Delete(&model.Room{}, id).Error
}
