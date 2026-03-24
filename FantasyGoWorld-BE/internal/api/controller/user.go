package controller

import (
	"fantasy-go-world-be/internal/biz"
	"fantasy-go-world-be/pkg/e"
	"fantasy-go-world-be/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetOnlineList 获取在线用户列表
// @router /api/v1/user/list [get]
func GetOnlineList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := biz.GetOnlineList(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, e.DatabaseError, nil)
		return
	}

	response.Success(c, gin.H{
		"items":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
