package controller

import (
	"fantasy-go-world-be/internal/ws"
	"fantasy-go-world-be/pkg/e"
	"fantasy-go-world-be/pkg/response"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		// allow all cross origin for dev. Use exact domains for prod
		return true
	},
}

// ServeWS 升级并且转交控制权给 Hub
// @router /ws [get]
func ServeWS(c *gin.Context) {
	// Middleware 中间件已经通过 Header 或 Cookie-Fallback 解析出了 uid 与 nickname 等信息
	uidAny, exists := c.Get("uid")
	if !exists {
		response.Error(c, http.StatusUnauthorized, e.Unauthorized, nil)
		return
	}
	uid := uidAny.(uint)

	nicknameAny, _ := c.Get("nickname")
	nickname := "unknown"
	if nicknameAny != nil {
		nickname = nicknameAny.(string)
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	client := &ws.Client{
		Hub:      ws.GlobalHub,
		Conn:     conn,
		UserID:   int(uid),
		Nickname: nickname,
		Elo:      1500, // 默认桩，实际应该去 DB 查 Profile
		SendCh:   make(chan []byte, 256),
	}

	// 刚连接，必定没入室
	client.RoomInbound.Store(nil)

	// 把这个新加入的受信任连接塞入全局 Hub 队列
	client.Hub.Register <- client

	// 开启双泵 (由于是无限循环，所以用独立协程跑 Read，当前协程当作 Write 挂起)
	go client.ReadPump()
	client.WritePump()
}
