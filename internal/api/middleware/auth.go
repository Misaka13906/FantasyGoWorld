package middleware

import (
	"net/http"
	"strings"

	"github.com/Misaka13906/FantasyGoWorld/internal/api/model/response"
	"github.com/Misaka13906/FantasyGoWorld/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Resp{
				Code: http.StatusUnauthorized,
				Data: nil,
				Msg:  "Authorization header is required",
			})
			return
		}

		authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Resp{
				Code: http.StatusUnauthorized,
				Data: nil,
				Msg:  "Invalid token",
			})
			return
		}

		c.Set("uid", claims.UID)

		c.Next()
	}
}
