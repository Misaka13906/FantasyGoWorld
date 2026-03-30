package ws

import (
	"encoding/json"
)

// MsgType 定义所有的 WebSocket 协议动作常量
const (
	MsgTypeHeartbeat   = "HEARTBEAT"
	MsgTypeConnect     = "CONNECT"    // 内部控制帧
	MsgTypeDisconnect  = "DISCONNECT" // 内部控制帧
	MsgTypeJoinRoom    = "JOIN_ROOM"
	MsgTypeLeaveRoom   = "LEAVE_ROOM"
	MsgTypeInvite      = "INVITE"
	MsgTypeInviteReply = "INVITE_REPLY"
	MsgTypeProposal    = "PROPOSAL"
	MsgTypeMove        = "MOVE"
	MsgTypePass        = "PASS"
	MsgTypeResign      = "RESIGN"
)

// Envelope 是所有 WebSocket 消息的外层包装协议
// 参考: docs/design/api-spec.md
type Envelope struct {
	Type      string          `json:"type"`                // 消息类型，全大写
	Seq       int             `json:"seq"`                 // 序列号
	RoomID    string          `json:"room_id"`             // 所属房间（大厅为 "hall"）
	Payload   json.RawMessage `json:"payload"`             // 业务载荷，随 Type 不同而不同
	Timestamp int64           `json:"timestamp,omitempty"` // 毫秒时间戳（仅发出附加，前端可选接收）

	// 以下字段不序列化，仅服务端内部路由使用
	Client *Client `json:"-"`
}
