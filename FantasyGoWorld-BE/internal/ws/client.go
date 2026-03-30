package ws

import (
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 15 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 65536 // 64KB
)

// Client 代理对单个 WebSocket 连接的所有访问
type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	UserID      int
	Nickname    string
	Elo         int
	SendCh      chan []byte                    // WritePump 消费；其他 Goroutine 仅可发不可读
	RoomInbound atomic.Pointer[chan *Envelope] // Hub 写一次，ReadPump 持续读取
}

// ReadPump 将从 WebSocket 连接读取的消息推送到 Hub 或 Room。
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var env Envelope
		if err := json.Unmarshal(message, &env); err != nil {
			log.Printf("invalid payload format from client %d: %v", c.UserID, err)
			continue
		}

		env.Client = c

		// 路由决策逻辑: 根据 docs/design/ws-design.md
		switch env.Type {
		// 这些需要由 Hub 也就是控制平面统一处理的协调事件
		case MsgTypeHeartbeat, MsgTypeJoinRoom, MsgTypeLeaveRoom, MsgTypeInvite, MsgTypeInviteReply:
			c.Hub.Inbound <- &env

		// 这些完全内部消化在 Room 内的高频事件
		case MsgTypeProposal, MsgTypeMove, MsgTypePass, MsgTypeResign:
			// 从 atomic 中安全的借出当前的绑定通道
			if chPtr := c.RoomInbound.Load(); chPtr != nil {
				(*chPtr) <- &env
			} else {
				log.Printf("client %d trying to send room message but not in room", c.UserID)
			}

		default:
			log.Printf("unknown message type %s from client %d", env.Type, c.UserID)
		}
	}
}

// WritePump 提取队列数据，唯一的途径将其以帧的方式推往网络层
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.SendCh:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.SendCh)
			for i := 0; i < n; i++ {
				w.Write(<-c.SendCh)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
