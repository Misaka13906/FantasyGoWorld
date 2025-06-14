package dao

import (
	"time"

	"gorm.io/gorm"
)

type Game struct {
	gorm.Model
	// ### 对局基本信息
	GID         uint   `gorm:"uniqueIndex;not null"`
	WhiteUserID string `gorm:"index;not null"`
	BlackUserID string `gorm:"index;not null"`
	WhiteUser   User   `gorm:"foreignKey:WhiteUserID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	BlackUser   User   `gorm:"foreignKey:BlackUserID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// ### 对局设置
	RuleID  string     `gorm:"not null"` // 规则 ID
	TimerID string     `gorm:"not null"` // 计时器 ID
	Rule    *GameRule  `gorm:"foreignKey:RuleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Timer   *GameTimer `gorm:"foreignKey:TimerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// 允许向对手申请悔棋
	PermitUndo bool `gorm:"not null;default:false"`
	// ### 对局结果
	IsFinished bool `gorm:"not null;default:false"`
	//
	WinnerID string `gorm:"index"`
	LoserID  string `gorm:"index"`
	Winner   User   `gorm:"foreignKey:WinnerID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Loser    User   `gorm:"foreignKey:LoserID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// 数目，中盘，超时，平局，未结束
	EndType string
	// 胜负分差，正数表示白胜，负数表示黑胜
	ScoreDiff float64 `gorm:"not null;default:0"`
	// 最大落子标号，用于计算对局总手数
	MaxMoveNum int `gorm:"not null;default:0"`
	// 对局开始、结束时间
	StartTime time.Time `gorm:"not null"`
	EndTime   time.Time `gorm:"not null"`
}

type Move struct {
	gorm.Model
	// 对局 ID
	GameID uint `gorm:"index;not null;uniqueIndex:idx_game_move"`
	// 落子序列号，表示在对局中的第几手
	MoveSequence int `gorm:"not null;uniqueIndex:idx_game_move"`
	// 落子位置，格式如 "D4"
	Position string `gorm:"not null"`
	// 落子颜色，true 为黑子，false 为白子
	StoneColor bool `gorm:"not null"`
	Game       Game `gorm:"foreignKey:GameID;references:GID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (g *Game) Create(tx *gorm.DB) error {
	return tx.Create(g).Error
}

func (g *Game) GetByGID(tx *gorm.DB) error {
	return tx.Where("game_id = ?", g.GID).First(g).Error
}

func (g *Game) GetUserTotalGames(tx *gorm.DB, uid string) error {
	// 将用户执黑和执白的对局都统计在内
	return tx.Model(g).
		Select("count(*) as total_games").
		Where("is_finished AND (white_user_id = ? OR black_user_id = ?)", uid, uid).
		First(g).Error
}

func (g *Game) GetUserTotalWins(tx *gorm.DB, uid string) error {
	// 统计用户的胜利次数
	return tx.Model(g).
		Select("count(*) as total_wins").
		Where("is_finished AND winner = ?", uid).
		First(g).Error
}

func (g *Game) GetUserTotalLosses(tx *gorm.DB, uid string) error {
	// 统计用户的失败次数
	return tx.Model(g).
		Select("count(*) as total_losses").
		Where("is_finished AND loser = ?", uid).
		First(g).Error
}
