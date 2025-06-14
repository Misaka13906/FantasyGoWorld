package dao

import (
	"gorm.io/gorm"
)

type GameRule struct {
	gorm.Model
	// 是否为自定义规则
	IsCustom bool `gorm:"not null;default:false"`
	// 规则名称
	Name string `gorm:"not null"`
	// 规则类型，0为中国规则，1为日本规则
	RuleType int `gorm:"not null;default:0"`
	// 贴目，如 7.5，6.5 等
	Komi float64 `gorm:"not null;default:7.5"`
	// 棋份，0为分先，1为让先，2以上为让子
	Handicap int `gorm:"not null;default:0"`
	// 让子摆放位置，格式：{"D4", "Q16"}，存储为 JSON 数组
	HandicapPositions []string `gorm:"type:text;not null"`
}

type GameTimer struct {
	gorm.Model
	// 是否为自定义计时器
	IsCustom bool `gorm:"not null;default:false"`
	// 计时器名称
	Name string `gorm:"not null"`
	// 计时规则：读秒制，加秒制，包干制
	TimerType int `gorm:"not null;default:0"`
	// 基础时间，单位为秒
	BaseTime int `gorm:"not null;default:0"`
	// 读秒/加秒时间，单位为秒
	ByoyomiTime int `gorm:"not null;default:0"`
	// 读秒次数
	ByoyomiCount int `gorm:"not null;default:0"`
}

func (gr *GameRule) Create(tx *gorm.DB) error {
	return tx.Create(gr).Error
}

func (gr *GameRule) GetByID(tx *gorm.DB, id uint) error {
	return tx.Where("id = ?", id).First(gr).Error
}

func (gr *GameRule) GetDefaultList(tx *gorm.DB) ([]GameRule, error) {
	var rules []GameRule
	err := tx.Where("is_custom = ?", false).Find(&rules).Error
	return rules, err
}

func (gr *GameRule) Update(tx *gorm.DB) error {
	return tx.Where("id = ?", gr.ID).Updates(gr).Error
}

func (gt *GameTimer) Create(tx *gorm.DB) error {
	return tx.Create(gt).Error
}

func (gt *GameTimer) GetByID(tx *gorm.DB, id uint) error {
	return tx.Model(&GameTimer{}).Where("id = ?", id).First(gt).Error
}

func (gt *GameTimer) Update(tx *gorm.DB) error {
	return tx.Model(&GameTimer{}).Where("id = ?", gt.ID).Updates(gt).Error
}

func (gt *GameTimer) GetDefaultList(tx *gorm.DB) ([]GameTimer, error) {
	var timers []GameTimer
	err := tx.Where("is_custom = ?", false).Find(&timers).Error
	return timers, err
}
