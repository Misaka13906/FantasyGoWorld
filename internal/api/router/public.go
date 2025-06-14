package router

import (
	controller "github.com/Misaka13906/FantasyGoWorld/internal/api/controller/public"
	"github.com/gin-gonic/gin"
)

func routerPublic(e *gin.RouterGroup) {
	e.POST("/register", controller.Register)
	e.POST("/login", controller.Login)
}
