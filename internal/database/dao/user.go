package dao

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UID          string `gorm:"uniqueIndex;not null"`
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Email        string
	Phone        string
	// 个性签名
	PersonalSignature string
	// 段级位
	Level string `gorm:"default:'0'"`
	// 对局数据统计
	TotalGames  int `gorm:"default:0"`
	TotalWins   int `gorm:"default:0"`
	TotalLosses int `gorm:"default:0"`
	TotalDraws  int `gorm:"default:0"`
}

func (u *User) Create(tx *gorm.DB) error {
	return tx.Create(u).Error
}

func (u *User) GetByUID(tx *gorm.DB) error {
	return tx.Where("uid = ?", u.UID).First(u).Error
}

func (u *User) Update(tx *gorm.DB) error {
	return tx.Where("uid = ?", u.UID).Updates(u).Error
}

func (u *User) GetUserList(tx *gorm.DB, pageNum, pageSize int) ([]User, error) {
	var users []User
	err := tx.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, err
}

func (u *User) GetByUsername(tx *gorm.DB) error {
	return tx.Where("username = ?", u.Username).First(u).Error
}

func (u *User) SearchByUsername(tx *gorm.DB) ([]User, error) {
	var users []User
	err := tx.Where("username LIKE ?", "%"+u.Username+"%").Find(&users).Error
	return users, err
}
