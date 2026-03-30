package model

import (
	"time"
)

// Room 房间模型
// 对齐 docs/design/data-schema.md
type Room struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OwnerID       uint      `gorm:"not null;index" json:"owner_id"`
	Description   string    `gorm:"type:text" json:"description"`
	IsPublic      bool      `gorm:"not null;default:1" json:"is_public"`
	Password      string    `gorm:"type:varchar(100)" json:"-"`
	Status        uint8     `gorm:"not null;default:0;comment:0=等待中 1=进行中 2=已结束" json:"status"`
	CurrentGameID *uint     `gorm:"index" json:"current_game_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Owner User `gorm:"foreignKey:OwnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"owner,omitempty"`
}

func (Room) TableName() string {
	return "rooms"
}
