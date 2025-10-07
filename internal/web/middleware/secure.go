package middleware

import (
	"net/http"
	"sq/internal/secure"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth(tm secure.TokenManager) gin.HandlerFunc {
	return func(g *gin.Context) {
		authHeader := g.GetHeader("Authorization")
		if authHeader == "" {
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "you need to auth"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if _, err := tm.ValidateToken(tokenStr, true); err != nil {
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		g.Next()
	}
}
