package biz

import (
	"errors"
	"fantasy-go-world-be/internal/repository/db"
	"fantasy-go-world-be/internal/repository/model"
	"golang.org/x/crypto/bcrypt"
)

// CreateRoom 创建房间
func CreateRoom(ownerID uint, description, password string, isPublic bool) (*model.Room, error) {
	var pwdHash string
	if password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		pwdHash = string(h)
	}

	room := &model.Room{
		OwnerID:     ownerID,
		Description: description,
		Password:    pwdHash,
		IsPublic:    isPublic,
		Status:      0, // 等待中
	}

	if err := db.CreateRoom(room); err != nil {
		return nil, err
	}

	return room, nil
}

// ListPublicRooms 获取公开房间列表 (带分页)
func ListPublicRooms(page, pageSize int) ([]model.Room, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return db.ListPublicRooms(offset, pageSize)
}

// GetRoomByID 根据ID查询房间
func GetRoomByID(roomId uint) (*model.Room, error) {
	return db.GetRoomByID(roomId)
}

// CloseRoom 关闭房间
// @logic-hint 鉴权检查: 此操作仅允许房主执行。错误码使用 e.ErrAuthFailed 或自行定制。
func CloseRoom(roomId, requestUid uint) error {
	room, err := db.GetRoomByID(roomId)
	if err != nil {
		return err
	}

	if room.OwnerID != requestUid {
		return errors.New("forbidden: not the owner")
	}

	// 阶段三：从记录中删除本房间，后续阶段可以在此发 WS 解散消息通知房间内其他人
	return db.DeleteRoom(roomId)
}
