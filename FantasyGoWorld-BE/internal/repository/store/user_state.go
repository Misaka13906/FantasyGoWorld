package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StatusIdle   = "idle"
	StatusGaming = "gaming"
	StatusDND    = "dnd"
)

// UserState 存在 Redis 中代表用户的控制平面状态
type UserState struct {
	Status     string
	RoomID     string // 如果在大厅为 "hall"，房间内为房间 ID 的字符串形式 "123"
	LastActive string // 时间戳
}

// UpdateUserState 更新用户在 Redis 中的装态片段
func UpdateUserState(ctx context.Context, userID int, status, roomID string) error {
	if Redis == nil {
		return nil
	}
	key := fmt.Sprintf("user_state:%d", userID)
	return Redis.HSet(ctx, key, map[string]interface{}{
		"status":      status,
		"room_id":     roomID,
		"last_active": time.Now().Format(time.RFC3339),
	}).Err()
}

// GetUserState 获取用户的状态
func GetUserState(ctx context.Context, userID int) (*UserState, error) {
	if Redis == nil {
		return nil, nil
	}
	key := fmt.Sprintf("user_state:%d", userID)

	res, err := Redis.HGetAll(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil // Doesn't exist
	}

	return &UserState{
		Status:     res["status"],
		RoomID:     res["room_id"],
		LastActive: res["last_active"],
	}, nil
}

// DeleteUserState 删除用户状态（例如下线后结束宽限期清空）
func DeleteUserState(ctx context.Context, userID int) error {
	if Redis == nil {
		return nil
	}
	key := fmt.Sprintf("user_state:%d", userID)
	return Redis.Del(ctx, key).Err()
}
