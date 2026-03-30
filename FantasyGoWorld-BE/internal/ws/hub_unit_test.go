package ws

import (
	"fantasy-go-world-be/internal/repository/store"
	"testing"
	"time"
)

func TestHubRegistration(t *testing.T) {
	h := InitHub()
	go h.Run()

	c := &Client{
		Hub:    h,
		UserID: 42,
		Nickname: "testuser",
		SendCh: make(chan []byte, 1),
	}

	h.Register <- c

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if _, ok := h.Clients[42]; !ok {
		t.Error("Client not registered in hub")
	}

	if sess, ok := h.Sessions[42]; !ok {
		t.Error("Session not created in hub")
	} else if sess.Status != store.StatusIdle {
		t.Errorf("Expected status idle, got %s", sess.Status)
	}
}

func TestHubHeartbeatUpdate(t *testing.T) {
	h := InitHub()
	go h.Run()

	c := &Client{Hub: h, UserID: 1, Nickname: "u1"}
	h.Register <- c
	time.Sleep(50 * time.Millisecond)

	initialPulse := h.Sessions[1].LastPulse

	env := &Envelope{
		Type:   MsgTypeHeartbeat,
		Client: c,
	}
	h.Inbound <- env

	time.Sleep(50 * time.Millisecond)

	if !h.Sessions[1].LastPulse.After(initialPulse) {
		t.Error("LastPulse not updated after heartbeat")
	}
}
