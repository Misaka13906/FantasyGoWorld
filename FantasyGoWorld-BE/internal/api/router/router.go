package router

import (
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
		// More routes to come
		apiV1.GET("/health", func(c *gin.Context) {
			response.Success(c, map[string]string{"status": "ok"})
		})
	}

	return r
}
