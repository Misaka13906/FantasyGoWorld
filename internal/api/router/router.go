package router

import (
	"github.com/Misaka13906/FantasyGoWorld/internal/api/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetUpRouter() *gin.Engine {
	e := gin.Default()
	e.Use(cors.Default())
	// e.Use(middleware.Logger())

	g := e.Group("/api/v1")

	routerPublic(g.Group(""))
	routerAdmin(g.Group("/admin", middleware.Auth()))

	authedUserGroup := g.Group("", middleware.Auth())
	{
		routerUser(authedUserGroup.Group("/user"))
		routerGame(authedUserGroup.Group("/game"))
	}

	return e
}
