package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户主表模型
// 严格对齐 docs/design/data-schema.md
type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(50);not null;uniqueIndex:uk_username" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string         `gorm:"type:varchar(50);not null" json:"nickname"`
	Email        string         `gorm:"type:varchar(255)" json:"email"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Bio          string         `gorm:"type:text" json:"bio"`
	Rank         string         `gorm:"type:varchar(5);not null;default:'18K'" json:"rank"`
	Elo          int            `gorm:"type:int;not null;default:1500" json:"elo"`
	AvatarURL    string         `gorm:"type:text" json:"avatar_url"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
