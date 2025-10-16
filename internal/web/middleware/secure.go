package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ummuys/reportify/internal/secure"
)

func Auth(tm secure.TokenManager) gin.HandlerFunc {
	return func(g *gin.Context) {
		authHeader := g.GetHeader("Authorization")
		if authHeader == "" {
			g.Set("msg", "empty token")
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "you need to auth"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tm.ValidateToken(tokenStr, true)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user_id := int64(claims["user_id"].(float64))
		g.Set("user_id", user_id)
		g.Next()
	}
}
