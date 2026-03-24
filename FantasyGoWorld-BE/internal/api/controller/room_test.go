package controller_test

import (
	"bytes"
	"encoding/json"
	"fantasy-go-world-be/internal/api/router"
	"fantasy-go-world-be/internal/config"
	"fantasy-go-world-be/internal/repository/db"
	"fantasy-go-world-be/internal/repository/model"
	"fantasy-go-world-be/pkg/jwtauth"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupRoomTestDB() {
	config.GlobalConfig = &config.Config{}
	cfg := config.GetConfig()
	cfg.JWT.Secret = "test-secret"
	cfg.JWT.AccessExpire = 3600

	sqliteDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test db")
	}

	sqliteDB.AutoMigrate(&model.User{}, &model.Room{})
	db.DB = sqliteDB
	db.CreateUser(&model.User{Username: "roomowner", PasswordHash: "x", Nickname: "Owner"})
}

func mockAuthCookie(uid uint) *http.Cookie {
	token, _ := jwtauth.GenerateToken(uid, "test-secret", 3600)
	return &http.Cookie{Name: "access_token", Value: token}
}

func TestRoomEndpoints(t *testing.T) {
	setupRoomTestDB()
	gin.SetMode(gin.TestMode)
	r := router.NewRouter()

	// 1. Create Room
	body := map[string]interface{}{
		"description": "Test Room",
		"is_public":   true,
		"password":    "",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/room", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(mockAuthCookie(1))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("CreateRoom failed, expected 200, got %v: %v", w.Code, w.Body.String())
	}

	// 2. List Public Rooms
	wList := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/api/v1/room/list", nil)
	reqList.AddCookie(mockAuthCookie(1))
	r.ServeHTTP(wList, reqList)

	if wList.Code != http.StatusOK {
		t.Fatalf("ListPublicRooms failed, got %v: %v", wList.Code, wList.Body.String())
	}

	// 3. Get Room Info
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/api/v1/room/1", nil)
	reqGet.AddCookie(mockAuthCookie(1))
	r.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("GetRoomByID failed, got %v: %v", wGet.Code, wGet.Body.String())
	}

	// 4. Close Room
	wClose := httptest.NewRecorder()
	reqClose, _ := http.NewRequest("DELETE", "/api/v1/room/1", nil)
	reqClose.AddCookie(mockAuthCookie(1))
	r.ServeHTTP(wClose, reqClose)

	if wClose.Code != http.StatusOK {
		t.Fatalf("CloseRoom failed, got %v: %v", wClose.Code, wClose.Body.String())
	}
}
