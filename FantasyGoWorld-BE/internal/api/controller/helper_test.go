package controller_test

import (
	"fantasy-go-world-be/internal/config"
	"fantasy-go-world-be/pkg/jwtauth"
	"net/http"
)

func mockAuthCookie(uid uint) *http.Cookie {
	cfg := config.GetConfig()
	token, _ := jwtauth.GenerateToken(uid, cfg.JWT.Secret, int64(cfg.JWT.AccessExpire))
	return &http.Cookie{Name: "access_token", Value: token}
}
