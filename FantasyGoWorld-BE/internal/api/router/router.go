package router

import (
	"fantasy-go-world-be/internal/api/controller"
	"fantasy-go-world-be/internal/api/middleware"
	"fantasy-go-world-be/pkg/response"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

	r.GET("/ping", func(c *gin.Context) {
		response.Success(c, map[string]string{"message": "pong"})
	})

	apiV1 := r.Group("/api/v1")
	{
		// 认证模块
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", controller.Register)
			auth.POST("/login", controller.Login)
			auth.POST("/refresh", controller.Refresh)
			auth.POST("/logout", middleware.JWTAuth(), controller.Logout)
		}

		// 用户模块
		user := apiV1.Group("/user")
		user.Use(middleware.JWTAuth())
		{
			user.GET("/list", controller.GetOnlineList)
			// 其他 user 逻辑占位
			user.GET("/me", func(c *gin.Context) {
				uid, _ := c.Get("uid")
				response.Success(c, gin.H{"uid": uid})
			})
		}

		// 房间模块
		room := apiV1.Group("/room")
		room.Use(middleware.JWTAuth())
		{
			room.POST("", controller.CreateRoom)
			room.GET("/list", controller.ListPublicRooms)
			room.GET("/:roomId", controller.GetRoomByID)
			room.DELETE("/:roomId", controller.CloseRoom)
		}

		// 健康检查
		apiV1.GET("/health", func(c *gin.Context) {
			response.Success(c, map[string]string{"status": "ok"})
		})
	}

	return r
}
