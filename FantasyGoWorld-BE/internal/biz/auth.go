package biz

import (
	"errors"
	"fantasy-go-world-be/internal/config"
	"fantasy-go-world-be/internal/repository/db"
	"fantasy-go-world-be/internal/repository/model"
	"fantasy-go-world-be/pkg/jwtauth"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
	Rank     string `json:"rank" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"-"` // Handled via Cookie in controller
	User         *model.User `json:"user"`
}

// Register 用户注册
func Register(req *RegisterRequest) error {
	// 1. 检查用户是否已存在
	_, err := db.GetUserByUsername(req.Username)
	if err == nil {
		return errors.New("username already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 2. 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 3. 创建用户
	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
		Rank:         req.Rank,
		Elo:          1500, // 初始分
	}

	return db.CreateUser(user)
}

// Login 用户登录
func Login(req *LoginRequest) (*LoginResponse, error) {
	// 1. 查询用户
	user, err := db.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. 生成 Access Token
	cfg := config.GetConfig()
	accessToken, err := jwtauth.GenerateToken(user.ID, cfg.JWT.Secret, cfg.JWT.AccessExpire)
	if err != nil {
		return nil, err
	}

	// 4. 生成 Refresh Token
	refreshToken, err := jwtauth.GenerateToken(user.ID, cfg.JWT.Secret, cfg.JWT.RefreshExpire)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshToken 刷新 Token
func RefreshToken(oldRefreshToken string) (string, error) {
	cfg := config.GetConfig()
	claims, err := jwtauth.ParseToken(oldRefreshToken, cfg.JWT.Secret)
	if err != nil {
		return "", err
	}

	// 重新生成 Access Token
	return jwtauth.GenerateToken(claims.UserID, cfg.JWT.Secret, cfg.JWT.AccessExpire)
}
