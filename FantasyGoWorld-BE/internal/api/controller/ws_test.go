package controller_test

import (
	"encoding/json"
	"fantasy-go-world-be/internal/api/router"
	"fantasy-go-world-be/internal/config"
	"fantasy-go-world-be/internal/repository/db"
	"fantasy-go-world-be/internal/repository/model"
	"fantasy-go-world-be/internal/ws"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

func setupWSTestDB() {
	config.GlobalConfig = &config.Config{}
	cfg := config.GetConfig()
	cfg.JWT.Secret = "test-secret"
	cfg.JWT.AccessExpire = 3600

	sqliteDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test db")
	}

	sqliteDB.AutoMigrate(&model.User{})
	db.DB = sqliteDB
	db.CreateUser(&model.User{Username: "wsuser", PasswordHash: "xxx", Nickname: "WS Tester"})
}

func TestWSConnectionAndHeartbeat(t *testing.T) {
	setupWSTestDB()
	// Disable real Redis for this unit/server test
	// We might need to mock store.Redis if ws.Hub uses it...
	// For now, let's just use try-catch/if-nil in the code or mock it.
	
	gin.SetMode(gin.TestMode)
	hub := ws.InitHub()
	go hub.Run()

	r := router.NewRouter()
	s := httptest.NewServer(r)
	defer s.Close()

	// Convert http URL to ws URL
	wsUrl := "ws" + strings.TrimPrefix(s.URL, "http") + "/api/v1/ws"
	// Wait! My router registered it at /ws directly, not /api/v1/ws?
	// Let's recheck router.go...
	// r.GET("/ws", middleware.JWTAuth(), controller.ServeWS)
	wsUrl = "ws" + strings.TrimPrefix(s.URL, "http") + "/ws"

	// Add cookie for auth
	cookie := mockAuthCookie(1)
	header := http.Header{}
	header.Add("Cookie", cookie.String())

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsUrl, header)
	if err != nil {
		t.Fatalf("Failed to dial WS: %v", err)
	}
	defer conn.Close()

	// 1. Test Heartbeat Send
	hb := ws.Envelope{
		Type:    ws.MsgTypeHeartbeat,
		RoomID:  "hall",
		Payload: json.RawMessage("{}"),
	}
	hbBytes, _ := json.Marshal(hb)
	if err := conn.WriteMessage(websocket.TextMessage, hbBytes); err != nil {
		t.Fatalf("Failed to write heartbeat: %v", err)
	}

	// 2. Test wait for Hub to update session (this is internal state, hard to verify via network without a GetStatus msg)
	// But we can check if the connection stays alive.
	time.Sleep(1 * time.Second)
	
	// Write something again to verify connection is open
	if err := conn.WriteMessage(websocket.TextMessage, hbBytes); err != nil {
		t.Fatalf("Connection should still be alive: %v", err)
	}
}
