package ws

import (
	"time"
)

// UserSession 对应业务模型 User 的运行时快照
type UserSession struct {
	UserID       int
	Nickname     string
	Elo          int
	Status       string    // "idle" | "gaming" | "dnd"
	ActiveGameID int       // 0=未参与对局（含观战）; >0=当前对局者
	LastPulse    time.Time // 心跳时间，心跳超时则标记断线
}

// MemRoom 对应业务模型 Room，房间运行态
type MemRoom struct {
	ID       int
	OwnerID  int
	IsPublic bool
	Status   string          // "waiting" | "ongoing"
	Members  map[int]*Member // UserID -> Member
	Config   *GameConfig     // 当前协商配置契约
	Instance *GameInstance   // Status="ongoing" 后非 nil
}

// Member 房间内成员
type Member struct {
	UserID int
	Role   string // "spectator" | "black" | "white"
}

// GameConfig 对局配置契约
type GameConfig struct {
	BoardSize       int     `json:"board_size"`
	RuleType        int     `json:"rule_type"`
	Komi            float64 `json:"komi"`
	TimeSystem      string  `json:"time_system"`
	MainTimeSeconds int     `json:"main_time_seconds"`
	ByoyomiPeriods  int     `json:"byoyomi_periods"`
	ByoyomiSeconds  int     `json:"byoyomi_seconds"`
	FirstBlackID    int     `json:"first_black_id"`
}

// GameTimer 计时器(简化版占位，阶段四关注连接，阶段五再完善业务逻辑)
type GameTimer struct{}

// GameInstance 对局运行态
type GameInstance struct {
	GameID     int
	Config     GameConfig
	Board      [19][19]int8
	StepSeq    int
	LastMoveAt time.Time
	Timer      *GameTimer
}

// GameActivatedEvent 用于 Room 激活后异步通知 Hub
type GameActivatedEvent struct {
	BlackUserID int
	WhiteUserID int
	GameID      int
}
