package controller_test

import (
	"bytes"
	"encoding/json"
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

func setupTestDB() {
	config.GlobalConfig = &config.Config{}
	cfg := config.GetConfig()
	cfg.JWT.Secret = "test-secret"
	cfg.JWT.AccessExpire = 3600
	cfg.JWT.RefreshExpire = 3600 * 24 * 7

	sqliteDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test db")
	}

	sqliteDB.AutoMigrate(&model.User{})
	db.DB = sqliteDB
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return router.NewRouter()
}

func TestRegister(t *testing.T) {
	setupTestDB()
	r := setupTestRouter()

	body := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
		"nickname": "Test User",
		"rank":     "18K",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %v: %v", w.Code, w.Body.String())
	}
}

func TestLoginAndRefreshLogOut(t *testing.T) {
	setupTestDB()
	r := setupTestRouter()

	// 1. Register
	body := map[string]interface{}{
		"username": "testuser2",
		"password": "password123",
		"nickname": "Test User2",
		"rank":     "18K",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// 2. Login
	loginBody := map[string]interface{}{
		"username": "testuser2",
		"password": "password123",
	}
	jsonLogin, _ := json.Marshal(loginBody)

	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("Login failed: %v", wLogin.Body.String())
	}

	// Read refresh_token cookie
	var refreshToken string
	var accessToken string
	for _, cookie := range wLogin.Result().Cookies() {
		if cookie.Name == "refresh_token" {
			refreshToken = cookie.Value
		}
		if cookie.Name == "access_token" {
			accessToken = cookie.Value
		}
	}

	if refreshToken == "" {
		t.Fatal("Expected refresh_token cookie")
	}

	if accessToken == "" {
		t.Fatal("Expected access_token cookie")
	}

	// 3. Refresh
	wRefresh := httptest.NewRecorder()
	reqRefresh, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
	reqRefresh.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	r.ServeHTTP(wRefresh, reqRefresh)

	if wRefresh.Code != http.StatusOK {
		t.Fatalf("Refresh failed: %v", wRefresh.Body.String())
	}

	// 4. Logout (requires auth)
	wLogout := httptest.NewRecorder()
	reqLogout, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	// Pass access_token
	reqLogout.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	r.ServeHTTP(wLogout, reqLogout)

	if wLogout.Code != http.StatusOK {
		t.Fatalf("Logout failed: %v", wLogout.Body.String())
	}

	// Verify logout deleted cookies
	logoutCookies := wLogout.Result().Cookies()
	var rtCleared bool
	var atCleared bool
	for _, c := range logoutCookies {
		if c.Name == "refresh_token" && c.Value == "" {
			rtCleared = true
		}
		if c.Name == "access_token" && c.Value == "" {
			atCleared = true
		}
	}
	if !rtCleared || !atCleared {
		t.Fatal("Expected access_token and refresh_token to be cleared")
	}
}
