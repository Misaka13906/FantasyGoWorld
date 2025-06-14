package controller

import (
	"net/http"

	"github.com/Misaka13906/FantasyGoWorld/internal/api/model/request"
	"github.com/Misaka13906/FantasyGoWorld/internal/api/model/response"
	service "github.com/Misaka13906/FantasyGoWorld/internal/service/user"
	"github.com/Misaka13906/FantasyGoWorld/pkg/jwt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithHttpCode(c, http.StatusBadRequest)
		return
	}

	user := service.User.Begin()
	if user.Error != nil {
		zap.L().Error("Failed to begin transaction: ", zap.Error(user.Error))
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}
	committed := false
	defer func() {
		err := user.Rollback().Error
		if err != nil && !committed {
			zap.L().Error("Failed to rollback transaction: ", zap.Error(err))
		}
	}()

	newUser, err := user.CreateUser(&req)
	if err != nil {
		if err.Error() == response.MsgMap[response.ErrorCodeConflict] {
			response.Error(c, response.ErrorCodeConflict)
			return
		}
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}

	if err := user.Commit().Error; err != nil {
		zap.L().Error("Failed to commit transaction: ", zap.Error(err))
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}
	committed = true

	claims := jwt.NewClaims(newUser.UID, 24) // Token valid for 1 day
	tokenString, err := jwt.GenerateToken(claims)
	if err != nil {
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}

	response.SuccessWithData(c, gin.H{
		"token":    tokenString,
		"uid":      newUser.UID,
		"username": newUser.Username,
	})
}

func Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithHttpCode(c, http.StatusBadRequest)
		return
	}

	user, err := service.User.GetByUsername(req.Username)
	if err != nil {
		response.ErrorWithHttpCode(c, http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil { // Password does not match
		zap.L().Error("Invalid username or password", zap.Error(err))
		response.ErrorWithHttpCode(c, http.StatusUnauthorized)
		return
	}

	claims := jwt.NewClaims(user.UID, 24) // Token valid for 1 day
	tokenString, err := jwt.GenerateToken(claims)
	if err != nil {
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}

	response.SuccessWithData(c, gin.H{
		"token":    tokenString,
		"uid":      user.UID,
		"username": user.Username,
	})
}
