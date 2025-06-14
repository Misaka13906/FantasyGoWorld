package dao

import (
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	RID         int    `gorm:"column:room_id;uniqueIndex;not null"` // 房间 ID
	OwnerUID    string `gorm:"column:owner_uid;index;not null"`     // 房主用户 ID
	OwnerUser   User   `gorm:"foreignKey:OwnerUID;references:UID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Description string `gorm:"column:description"`                     // 房间描述
	IsPublic    bool   `gorm:"column:is_public;not null;default:true"` // 是否公开房间
	Password    string `gorm:"column:password"`                        // 房间密码，公开房间可为空
	// 房间状态，0为等待中，1为进行中，2为已结束
	Status        int    `gorm:"column:status;not null;default:0"` // 房间状态
	CurrentGameID string `gorm:"column:current_game_id;index"`     // 当前进行中的对局 ID
	CurrentGame   Game   `gorm:"foreignKey:CurrentGameID;references:GID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Participants  []User `gorm:"many2many:room_participants;foreignKey:RID;joinForeignKey:RoomID;References:UID;joinReferences:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // 参与者列表
}

func (r *Room) Create(tx *gorm.DB) error {
	return tx.Create(r).Error
}

func (r *Room) GetByRID(tx *gorm.DB) error {
	return tx.First(r, gorm.Expr("room_id = ?", r.RID)).Error
}

func (r *Room) Update(tx *gorm.DB) error {
	return tx.Model(r).Where("room_id = ?", r.RID).Updates(r).Error
}

func (r *Room) Delete(tx *gorm.DB) error {
	return tx.Where("room_id = ?", r.RID).Delete(r).Error
}

func (r *Room) GetRoomList(tx *gorm.DB, pageNum, pageSize int) ([]Room, error) {
	var rooms []Room
	return rooms, tx.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rooms).Error
}

func (r *Room) GetByOwnerUID(tx *gorm.DB) ([]Room, error) {
	var rooms []Room
	return rooms, tx.Where("owner_uid = ?", r.OwnerUID).Find(&rooms).Error
}

func (r *Room) GetByParticipantUID(tx *gorm.DB, uid string) ([]Room, error) {
	var rooms []Room
	return rooms, tx.Model(&Room{}).Joins("JOIN room_participants ON room_participants.room_id = rooms.room_id").
		Where("room_participants.user_id = ?", uid).Find(&rooms).Error
}
