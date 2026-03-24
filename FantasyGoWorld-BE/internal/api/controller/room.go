package controller

import (
	"fantasy-go-world-be/internal/biz"
	"fantasy-go-world-be/pkg/e"
	"fantasy-go-world-be/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type createRoomRequest struct {
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	Password    string `json:"password"`
}

// CreateRoom 创建房间
// @router /api/v1/room [post]
func CreateRoom(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, e.RequestFieldError, nil)
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		response.Error(c, http.StatusUnauthorized, e.Unauthorized, nil)
		return
	}

	room, err := biz.CreateRoom(uid.(uint), req.Description, req.Password, req.IsPublic)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, e.DatabaseError, nil)
		return
	}

	response.Success(c, room)
}

// ListPublicRooms 获取公开房间列表
// @router /api/v1/room/list [get]
func ListPublicRooms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	rooms, total, err := biz.ListPublicRooms(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, e.DatabaseError, nil)
		return
	}

	response.Success(c, gin.H{
		"items":     rooms,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetRoomByID 根据ID查询房间
// @router /api/v1/room/:roomId [get]
func GetRoomByID(c *gin.Context) {
	roomIdStr := c.Param("roomId")
	roomId, err := strconv.ParseUint(roomIdStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, e.RequestFieldError, nil)
		return
	}

	room, err := biz.GetRoomByID(uint(roomId))
	if err != nil {
		response.Error(c, http.StatusNotFound, e.NotFound, nil)
		return
	}

	response.Success(c, room)
}

// CloseRoom 关闭房间 (仅房主)
// @router /api/v1/room/:roomId [delete]
func CloseRoom(c *gin.Context) {
	roomIdStr := c.Param("roomId")
	roomId, err := strconv.ParseUint(roomIdStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, e.RequestFieldError, nil)
		return
	}

	uid, exists := c.Get("uid")
	if !exists {
		response.Error(c, http.StatusUnauthorized, e.Unauthorized, nil)
		return
	}

	if err := biz.CloseRoom(uint(roomId), uid.(uint)); err != nil {
		if err.Error() == "forbidden: not the owner" {
			response.Error(c, http.StatusForbidden, e.NotRoomOwner, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, e.DatabaseError, nil)
		return
	}

	response.Success(c, gin.H{})
}
