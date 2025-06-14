package router

import (
	controller "github.com/Misaka13906/FantasyGoWorld/internal/api/controller/user"
	"github.com/gin-gonic/gin"
)

func routerUser(user *gin.RouterGroup) {
	user.GET("/:uid", controller.GetUserInfo)
	user.PUT("", controller.UpdateUserProfile)
	user.GET("/list", controller.GetUserList)
	user.GET("/search", controller.SearchUserByUsername)
}

func routerGame(game *gin.RouterGroup) {
	game.GET("/:gameId", controller.GetGameInfo)
	// game.POST("/:gameId/update", controller.UpdateGame)
}
