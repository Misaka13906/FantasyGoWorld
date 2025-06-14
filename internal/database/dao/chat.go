package dao

import (
	"gorm.io/gorm"
)

// 房间关闭后销毁
type ChatMessage struct {
	gorm.Model
	MID        string `gorm:"column:mid;uniqueIndex;not null"` // 消息 ID
	SenderUID  string `gorm:"index;not null"`                  // 发送者用户 ID
	SenderUser User   `gorm:"foreignKey:SenderUID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	RoomID     string `gorm:"index;not null"` // 房间 ID
	Room       Room   `gorm:"foreignKey:RoomID;references:RID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Content    string `gorm:"not null"` // 消息内容
	Timestamp  int64  `gorm:"not null"` // 消息发送时间戳
}

type LetterMessage struct {
	gorm.Model
	LID          string `gorm:"column:lid;uniqueIndex;not null"` // 消息 ID
	SenderUID    string `gorm:"index;not null"`                  // 发送者用户 ID
	SenderUser   User   `gorm:"foreignKey:SenderUID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ReceiverUID  string `gorm:"index;not null"` // 接收者用户 ID
	ReceiverUser User   `gorm:"foreignKey:ReceiverUID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Content      string `gorm:"not null"` // 消息内容
	Timestamp    int64  `gorm:"not null"` // 消息发送时间戳
}
