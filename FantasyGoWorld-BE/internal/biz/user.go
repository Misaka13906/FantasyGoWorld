package biz

import (
	"errors"
	"fantasy-go-world-be/internal/repository/db"
	"fantasy-go-world-be/internal/repository/model"
)

// GetProfile 获取用户资料
func GetProfile(uid uint) (*model.User, error) {
	return db.GetUserByID(uid)
}

// UpdateProfile 更新用户资料 (暂时不支持修改敏感字典)
// @stub 阶段三仅提供占位符，需要限制改动范围
func UpdateProfile(uid uint, updates map[string]interface{}) error {
	// TODO: 实现按字典更新 db 字段，注意排除 elo 等只读属性
	return errors.New("not implemented yet")
}

// GetOnlineList 获取大厅在线用户列表
// @ai-context 阶段三暂时使用全库查询模拟在线列表，Phase 4 将改查 Redis store 中的在线状态。
func GetOnlineList(page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return db.GetOnlineList(offset, pageSize)
}
