package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/models"
	"github.com/ummuys/reportify/internal/secure"
	"github.com/ummuys/reportify/internal/service"

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

// UpdateAccessToken godoc
// @Summary      Обновить access-токен по refresh-токену
// @Description  Читает refresh_token из Cookie и выдает новый access-токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Cookie  header  string  true  "Cookie: refresh_token=<REFRESH_TOKEN>"
// @Success      200     {object} models.NewAccessToken  "Новый access-токен"
// @Failure      401     {object} models.EmptyResponse   "Отсутствует/некорректный refresh-токен"
// @Failure      500     {object} models.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /secure/access [get]
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
			return
		}

		userID := int64(claims["user_id"].(float64))

		access, err := ah.tm.GenerateAccessToken(userID)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		g.Set("msg", "access token is updated")
		g.JSON(http.StatusOK, models.NewAccessToken{AccessToken: access})
	}
}

// Authorization godoc
// @Summary      Авторизация пользователя
// @Description  Проверяет логин и пароль, выставляет refresh_token в Cookie и возвращает access-токен в ответе.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body   models.Auth  true  "Учетные данные пользователя"
// @Success      200      {object}  models.NewAccessToken  "Успешная авторизация, access-токен в теле ответа"
// @Failure      400      {object}  models.EmptyResponse   "Неверный формат запроса (bad request)"
// @Failure      401      {object}  models.EmptyResponse   "Неверные учетные данные"
// @Failure      500      {object}  models.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /secure/auth [post]
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
		fmt.Println(refresh)
		g.JSON(http.StatusOK, models.NewAccessToken{AccessToken: access})
	}
}
