package ws

import (
	"context"
	"encoding/json"
	"fantasy-go-world-be/internal/repository/store"
	"log"
	"strconv"
	"time"
)

// Hub 管理游戏大厅所有的客户端连接
type Hub struct {
	Sessions map[int]*UserSession
	Clients  map[int]*Client
	Rooms    map[int]*Room

	// Channels
	Register   chan *Client
	Unregister chan *Client
	Inbound    chan *Envelope // 处理从直接由 Client 到 Hub 的请求 (如 JOIN_ROOM, INVITE, HEARTBEAT 等)
	Notify     chan *GameActivatedEvent
}

var GlobalHub *Hub

// InitHub 初始化全局 Hub
func InitHub() *Hub {
	GlobalHub = &Hub{
		Sessions:   make(map[int]*UserSession),
		Clients:    make(map[int]*Client),
		Rooms:      make(map[int]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Inbound:    make(chan *Envelope, 256), // 控制层消息带一定缓冲
		Notify:     make(chan *GameActivatedEvent, 10),
	}
	return GlobalHub
}

// Run 启动 Hub 的主事务循环 (唯一的 Writer)
func (h *Hub) Run() {
	// 启动定期清理挂机用户的心跳检测 Ticker
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.Register:
			log.Printf("Client Registered: UID=%d", client.UserID)
			h.Clients[client.UserID] = client

			// 如果 Session 不存在则创建，标记在线，并写回 Redis
			if _, exists := h.Sessions[client.UserID]; !exists {
				h.Sessions[client.UserID] = &UserSession{
					UserID:       client.UserID,
					Nickname:     client.Nickname,
					Elo:          client.Elo,
					Status:       store.StatusIdle,
					ActiveGameID: 0,
					LastPulse:    time.Now(),
				}
				// 异步写 Redis, 避免阻塞 Hub
				go store.UpdateUserState(context.Background(), client.UserID, store.StatusIdle, "hall")
			} else {
				h.Sessions[client.UserID].LastPulse = time.Now()
			}

		case client := <-h.Unregister:
			log.Printf("Client Unregistered: UID=%d", client.UserID)
			// 断开连接，启动断线保护。这里暂时直接删缓存来简化 (按文档描述应该是起 Timer 监控)
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.SendCh)
			}
			// (后续应对 UserSession 启动 Grace Period 保留，我们先立即删除模拟不重连)
			delete(h.Sessions, client.UserID)
			go store.DeleteUserState(context.Background(), client.UserID)

		case env := <-h.Inbound:
			h.handleInbound(env)

		case evt := <-h.Notify:
			// Game activated event
			if b, ok := h.Sessions[evt.BlackUserID]; ok {
				b.ActiveGameID = evt.GameID
				b.Status = store.StatusGaming
			}
			if w, ok := h.Sessions[evt.WhiteUserID]; ok {
				w.ActiveGameID = evt.GameID
				w.Status = store.StatusGaming
			}

		case <-ticker.C:
			// 扫描检测没有心跳的死链接 (15秒没心跳视为断线)
			now := time.Now()
			for uid, session := range h.Sessions {
				if now.Sub(session.LastPulse) > 15*time.Second {
					if c, ok := h.Clients[uid]; ok {
						log.Printf("Client Timeout: UID=%d", uid)
						c.Conn.Close() // 强制关连接，触发 unregister
					}
				}
			}
		}
	}
}

// handleInbound 处理控制层网络事件
func (h *Hub) handleInbound(env *Envelope) {
	client := env.Client
	if client == nil {
		return
	}

	switch env.Type {
	case MsgTypeHeartbeat:
		if s, ok := h.Sessions[client.UserID]; ok {
			s.LastPulse = time.Now()
		}

	case MsgTypeJoinRoom:
		// 解析房间号
		var req struct {
			RoomID int `json:"room_id"`
		}
		if err := json.Unmarshal(env.Payload, &req); err == nil {
			room, exists := h.Rooms[req.RoomID]
			if !exists {
				// Room 尚未在内存时，尝试新创建并启动
				room = NewRoom(req.RoomID, h)
				h.Rooms[req.RoomID] = room
				go room.Run()
			}

			// Hub 向 Client 注入 RoomInbound
			client.RoomInbound.Store(&room.Inbound)

			// 投递让 Room 处理 Client 加成员
			room.register <- client

			// 更新 Redis
			go store.UpdateUserState(context.Background(), client.UserID, store.StatusIdle, strconv.Itoa(req.RoomID)) // simplistic for now
		}

	case MsgTypeLeaveRoom:
		// 读取现在的绑定关系清零
		if oldInb := client.RoomInbound.Load(); oldInb != nil {
			var req struct {
				RoomID int `json:"room_id"`
			}
			json.Unmarshal(env.Payload, &req)

			if room, exists := h.Rooms[req.RoomID]; exists {
				room.leave <- client
			}
			client.RoomInbound.Store(nil)
			go store.UpdateUserState(context.Background(), client.UserID, store.StatusIdle, "hall")
		}

	case MsgTypeInvite:
		// 路由转发，不带 Room 上下文
		var req struct {
			TargetID int `json:"target_id"`
			RoomID   int `json:"room_id"`
		}
		if err := json.Unmarshal(env.Payload, &req); err == nil {
			if target, ok := h.Clients[req.TargetID]; ok {
				target.SendCh <- SerializeMsg(env)
			}
		}
	}
}

// SerializeMsg 把响应结构转为 JSON byte 流
func SerializeMsg(env *Envelope) []byte {
	b, _ := json.Marshal(env)
	return b
}
