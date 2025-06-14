package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Misaka13906/FantasyGoWorld/internal/api/model/request"
	"github.com/Misaka13906/FantasyGoWorld/internal/api/model/response"
	"github.com/Misaka13906/FantasyGoWorld/internal/database/dao"
	"github.com/Misaka13906/FantasyGoWorld/internal/database/dto"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var User = UserServ{dao.DB}

type UserServ struct {
	*gorm.DB
}

func (a UserServ) Begin() UserServ {
	a.DB = a.DB.Begin()
	return a
}

func (a UserServ) GetUserInfo(uid string) (*dto.UserProfile, error) {
	user := dao.User{UID: uid}

	err := user.GetByUID(a.DB)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			zap.L().Error("User not found: ", zap.String("uid", uid))
			return nil, err
		}
		zap.L().Error("Failed to get user by UID: ", zap.String("uid", uid), zap.Error(err))
		return nil, err
	}

	return &dto.UserProfile{
		Username:          user.Username,
		PersonalSignature: user.PersonalSignature,
		Level:             user.Level,
		TotalGames:        user.TotalGames,
		TotalWins:         user.TotalWins,
		TotalLosses:       user.TotalLosses,
	}, nil
}

func (a UserServ) CreateUser(userData *request.RegisterRequest) (*dto.UserBasicInfo, error) {
	user := dao.User{
		Username: userData.Username,
	}
	exists, err := a.CheckUsernameExists(user.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		zap.L().Error("User already exists: ", zap.String("username", userData.Username))
		return nil, errors.New(response.MsgMap[response.ErrorCodeConflict])
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("Failed to hash password: ", zap.Error(err))
		return nil, err
	}
	user.PasswordHash = string(passwordHash)

	// 这里使用 Unix 时间戳和用户名的哈希值来生成一个唯一的 UID
	usernameHash := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(userData.Username))
	user.UID = fmt.Sprintf("%d-%s-%d", time.Now().Unix(), usernameHash.String(), rand.Int()%100)

	if err := user.Create(a.DB); err != nil {
		zap.L().Error("Failed to create user: ", zap.Error(err))
		return nil, err
	}

	return &dto.UserBasicInfo{
		UID:      user.UID,
		Username: user.Username,
	}, nil
}

func (a UserServ) CheckUsernameExists(username string) (bool, error) {
	if _, err := a.GetByUsername(username); err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		zap.L().Error("Failed to check user exist: ", zap.Error(err))
		return false, err
	}
	return true, nil
}

func (a UserServ) GetByUsername(username string) (*dao.User, error) {
	user := dao.User{Username: username}
	if err := user.GetByUsername(a.DB); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		zap.L().Error("Failed to get user by username: ", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (a UserServ) UpdateUserProfile(userData *request.UpdateUserProfile) error {
	user := dao.User{
		UID:               userData.UID,
		PersonalSignature: userData.PersonalSignature,
	}
	err := user.Update(a.DB)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			zap.L().Error("User not found for update", zap.String("uid", userData.UID))
			return err
		}
		zap.L().Error("Failed to update user profile", zap.String("uid", userData.UID), zap.Error(err))
		return err
	}

	return nil
}

func (a UserServ) GetUserList(pageNum, pageSize int) ([]dto.UserProfile, error) {
	user := dao.User{}
	userListRaw, err := user.GetUserList(a.DB, pageNum, pageSize)
	if err != nil && err != gorm.ErrRecordNotFound {
		zap.L().Error("Failed to get user list: ", zap.Error(err))
		return nil, err
	}

	var userList []dto.UserProfile
	for _, u := range userListRaw {
		userList = append(userList, dto.UserProfile{
			Username:          u.Username,
			PersonalSignature: u.PersonalSignature,
			Level:             u.Level,
			TotalGames:        u.TotalGames,
			TotalWins:         u.TotalWins,
			TotalLosses:       u.TotalLosses,
		})
	}

	return userList, nil
}

func (a UserServ) SearchUserByUsername(username string) ([]dto.UserProfile, error) {
	if username == "" {
		return a.GetUserList(1, 50) // Default to first page with 50 items if no username is provided
	}

	user := dao.User{Username: username}
	userListRaw, err := user.SearchByUsername(a.DB)
	if err != nil && err != gorm.ErrRecordNotFound {
		zap.L().Error("Failed to search user by username: ", zap.String("username", username), zap.Error(err))
		return nil, err
	}

	var userList []dto.UserProfile
	for _, u := range userListRaw {
		userList = append(userList, dto.UserProfile{
			Username:          u.Username,
			PersonalSignature: u.PersonalSignature,
			Level:             u.Level,
			TotalGames:        u.TotalGames,
			TotalWins:         u.TotalWins,
			TotalLosses:       u.TotalLosses,
		})
	}

	return userList, nil
}
