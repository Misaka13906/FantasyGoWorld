package controller

import (
	"fantasy-go-world-be/internal/biz"
	"fantasy-go-world-be/pkg/e"
	"fantasy-go-world-be/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register 注册接口
// POST /auth/register
func Register(c *gin.Context) {
	var req biz.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, e.RequestFieldError, nil)
		return
	}

	if err := biz.Register(&req); err != nil {
		if err.Error() == "username already exists" {
			response.Error(c, http.StatusBadRequest, e.UsernameExists, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, e.DatabaseError, nil)
		return
	}

	response.Success(c, nil)
}

// Login 登录接口
// POST /auth/login
func Login(c *gin.Context) {
	var req biz.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, e.RequestFieldError, nil)
		return
	}

	res, err := biz.Login(&req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			response.Error(c, http.StatusUnauthorized, e.InvalidCredentials, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, e.ERROR, nil)
		return
	}

	// 将 access_token 写入 Cookie (SameSite=Lax, 非 httpOnly), 有效期 1h
	c.SetCookie("access_token", res.AccessToken, 3600, "/", "", false, false)
	// 将 Refresh Token 写入 httpOnly Cookie
	// TODO: Secure flag should be true in production (HTTPS)
	c.SetCookie("refresh_token", res.RefreshToken, 7*24*3600, "/", "", false, true)

	response.Success(c, res)
}

// Refresh 刷新 Token 接口
// POST /auth/refresh
func Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, http.StatusUnauthorized, e.Unauthorized, nil)
		return
	}

	accessToken, user, err := biz.RefreshToken(refreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, e.RefreshTokenExpired, nil)
		return
	}

	response.Success(c, gin.H{
		"access_token": accessToken,
		"user":         user,
	})
}

// Logout 退出登录接口
// POST /auth/logout
func Logout(c *gin.Context) {
	// 清除 Cookie
	c.SetCookie("access_token", "", -1, "/", "", false, false)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	response.Success(c, nil)
}
