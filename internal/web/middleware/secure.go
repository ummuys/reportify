package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/secure"
)

func Auth(tm secure.TokenManager, access []string) gin.HandlerFunc {
	return func(g *gin.Context) {
		authHeader := g.GetHeader("Authorization")
		if authHeader == "" {
			g.Set("msg", "empty token")
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": errs.ErrUserUnauthorized.Error()})
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			g.Set("msg", "invalid token format")
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errs.ErrBadAccessToken.Error()})
			return
		}

		tokenStr := parts[1]
		claims, err := tm.ValidateAccessToken(tokenStr)
		if err != nil {
			g.Set("msg", err.Error())
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errs.ErrBadAccessToken.Error()})
			return
		}

		user_id := claims.UserID
		role := claims.Role
		forbidden := true
		for _, acc := range access {
			if acc == role {
				forbidden = false
				break
			}
		}
		if forbidden {
			g.AbortWithStatus(http.StatusForbidden)
			return
		}
		g.Set("user_id", user_id)
		g.Next()
	}
}
