package handlers

import (
	"context"
	"net/http"
	models "sq/internal/models/request"
	"sq/internal/secure"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type authHandler struct {
	logger *zerolog.Logger
	tm     secure.TokenManager
}

func NewAuthHandler(logger *zerolog.Logger, tm secure.TokenManager) AuthHandler {
	return &authHandler{logger: logger, tm: tm}
}

func (ah *authHandler) UpdateRefreshToken(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		refreshToken, err := g.Cookie("refresh_token")
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": err.Error()})
			return
		}

		claims, err := ah.tm.ValidateToken(refreshToken, false)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "bad refresh token"})
		}

		nRefTok, err := ah.tm.GenerateRefreshToken(claims["username"].(string))
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusUnauthorized)
		}

		g.SetCookie("refresh_token", nRefTok, 3600*144, "/", "app", false, false)
	}
}

func (ah *authHandler) UpdateAccessToken(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		refreshToken, err := g.Cookie("refresh_token")
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": err.Error()})
			return
		}

		claims, err := ah.tm.ValidateToken(refreshToken, false)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "bad refresh token"})
		}

		accessToken, err := ah.tm.GenerateAccessToken(claims["username"].(string))
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusUnauthorized)
		}

		g.JSON(http.StatusOK, gin.H{"access_token": accessToken})
	}
}

func (ah *authHandler) Authorization(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		var req models.Auth
		if err := g.ShouldBindJSON(&req); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
			return
		}

		access, err := ah.tm.GenerateAccessToken(req.Username)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
		}

		refresh, err := ah.tm.GenerateAccessToken(req.Username)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
		}

		g.JSON(http.StatusOK, gin.H{"access": access, "refresh": refresh})
	}
}
