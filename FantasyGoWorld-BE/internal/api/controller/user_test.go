package controller_test

import (
	"fantasy-go-world-be/internal/api/router"
	"fantasy-go-world-be/internal/config"
	"fantasy-go-world-be/internal/repository/db"
	"fantasy-go-world-be/internal/repository/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB() {
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
	db.CreateUser(&model.User{Username: "testuser1", PasswordHash: "x", Nickname: "User1"})
	db.CreateUser(&model.User{Username: "testuser2", PasswordHash: "x", Nickname: "User2"})
}

func TestGetOnlineList(t *testing.T) {
	setupUserTestDB()
	gin.SetMode(gin.TestMode)
	r := router.NewRouter()

	wList := httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/api/v1/user/list", nil)
	reqList.AddCookie(mockAuthCookie(1)) // Using mockAuthCookie from room_test.go
	r.ServeHTTP(wList, reqList)

	if wList.Code != http.StatusOK {
		t.Fatalf("GetOnlineList failed, expected 200, got %v: %v", wList.Code, wList.Body.String())
	}
}
