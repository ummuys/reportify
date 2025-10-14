package handlers

import (
	"context"
	"errors"
	"net/http"
	"sq/internal/errs"
	"sq/internal/models"
	"sq/internal/secure"
	"sq/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type authHandler struct {
	logger *zerolog.Logger
	tm     secure.TokenManager
	u      service.UserService
}

func NewAuthHandler(logger *zerolog.Logger, tm secure.TokenManager, u service.UserService) AuthHandler {
	return &authHandler{logger: logger, tm: tm, u: u}
}

func (ah *authHandler) UpdateAccessToken(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		refreshToken, err := g.Cookie("refresh_token")
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, models.EmptyResponse{Message: err.Error()})
			return
		}

		claims, err := ah.tm.ValidateToken(refreshToken, false)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, models.EmptyResponse{Message: "bad refresh token"})
		}

		accessToken, err := ah.tm.GenerateAccessToken(claims["user_id"].(int64))
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusUnauthorized)
		}

		g.Set("msg", "access token is returned")
		g.JSON(http.StatusOK, gin.H{"access_token": accessToken})
	}
}

func (ah *authHandler) Authorization(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		var req models.Auth
		if err := g.ShouldBindJSON(&req); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "bad request"})
			return
		}

		id, err := ah.u.CheckPass(pCtx, req.Username, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrInvalidCredentials):
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusUnauthorized, models.EmptyResponse{Message: err.Error()})
				return
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: err.Error()})
			}
		}

		access, err := ah.tm.GenerateAccessToken(id)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		refresh, err := ah.tm.GenerateRefreshToken(id)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		g.Set("msg", "auth successful")
		g.SetCookie("refresh_token", refresh, 3600*144, "/", "", false, true)
		g.JSON(http.StatusOK, gin.H{"access_token": access})
	}
}
