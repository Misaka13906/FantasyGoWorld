package controller

import (
	"net/http"
	"strconv"

	"github.com/Misaka13906/FantasyGoWorld/internal/api/model/request"
	"github.com/Misaka13906/FantasyGoWorld/internal/api/model/response"
	service "github.com/Misaka13906/FantasyGoWorld/internal/service/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetUserInfo(c *gin.Context) {
	uid := c.Param("uid")

	userInfo, err := service.User.GetUserInfo(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, response.ErrorCodeRecordNotFound)
			return
		}
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}

	response.SuccessWithData(c, userInfo)
}

func UpdateUserProfile(c *gin.Context) {
	var userData request.UpdateUserProfile
	if err := c.ShouldBindJSON(&userData); err != nil {
		response.ErrorWithHttpCode(c, http.StatusBadRequest)
		return
	}
	if userData.UID != c.GetString("uid") {
		response.ErrorWithHttpCode(c, http.StatusForbidden)
		return
	}

	err := service.User.UpdateUserProfile(&userData)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, response.ErrorCodeRecordNotFound)
			return
		}
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}

	response.Success(c)
}

func GetUserList(c *gin.Context) {
	pageNumRaw := c.Query("pageNum")
	pageNum, err := strconv.Atoi(pageNumRaw)
	if err != nil || pageNum <= 0 {
		zap.L().Error("Invalid page number: ", zap.String("pageNum", pageNumRaw), zap.Error(err))
		response.ErrorWithHttpCode(c, http.StatusBadRequest)
		return
	}
	pageSizeRaw := c.Query("pageSize")
	pageSize, err := strconv.Atoi(pageSizeRaw)
	if err != nil || pageSize <= 0 {
		zap.L().Error("Invalid page size: ", zap.String("pageSize", pageSizeRaw), zap.Error(err))
		response.ErrorWithHttpCode(c, http.StatusBadRequest)
		return
	}

	users, err := service.User.GetUserList(pageNum, pageSize)
	if err != nil {
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		response.Error(c, response.ErrorCodeListNoRecords)
		return
	}

	response.SuccessWithData(c, users)
}

func SearchUserByUsername(c *gin.Context) {
	username := c.Query("username")

	users, err := service.User.SearchUserByUsername(username)
	if err != nil {
		response.ErrorWithHttpCode(c, http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		response.Error(c, response.ErrorCodeListNoRecords)
		return
	}

	response.SuccessWithData(c, users)
}
