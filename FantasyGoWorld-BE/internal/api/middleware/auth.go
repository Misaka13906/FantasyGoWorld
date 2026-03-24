package middleware

import (
	"fantasy-go-world-be/internal/config"
	"fantasy-go-world-be/pkg/e"
	"fantasy-go-world-be/pkg/jwtauth"
	"fantasy-go-world-be/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth 鉴权中间件
// 严格对齐 api-spec.md 附录 B
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1. 优先从 Authorization Header 读取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader[7:]
		}

		// 2. Fallback: 从 Cookie 读取 (WebSocket 或某些前端场景)
		if token == "" {
			token, _ = c.Cookie("access_token")
		}

		if token == "" {
			response.Error(c, http.StatusUnauthorized, e.Unauthorized, nil)
			c.Abort()
			return
		}

		// 3. 解析并校验
		cfg := config.GetConfig()
		claims, err := jwtauth.ParseToken(token, cfg.JWT.Secret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, e.TokenExpired, nil)
			c.Abort()
			return
		}

		// 将 UID 注入上下文
		c.Set("uid", claims.UserID)
		c.Next()
	}
}
