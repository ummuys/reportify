package handlers

import (
	"errors"
	"net/http"

	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/secure"
	"github.com/ummuys/reportify/internal/service"
	"github.com/ummuys/reportify/internal/webdto"

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
// @Success      200     {object} webdto.NewAccessToken  "Новый access-токен"
// @Failure      401     {object} webdto.EmptyResponse   "Отсутствует/некорректный refresh-токен"
// @Failure      500     {object} webdto.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /secure/access [get]
func (ah *authHandler) UpdateAccessToken() gin.HandlerFunc {
	return func(g *gin.Context) {
		_ = g.Request.Context()
		refreshToken, err := g.Cookie("refresh_token")
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, webdto.EmptyResponse{Message: errs.ErrBadRefreshToken.Error()})
			return
		}

		claims, err := ah.tm.ValidateRefreshToken(refreshToken)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusUnauthorized, webdto.EmptyResponse{Message: errs.ErrBadRefreshToken.Error()})
			return
		}

		userID := claims.UserID
		role := claims.Role

		access, err := ah.tm.GenerateAccessToken(userID, role)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInvalidCredentials.Error()})
			return
		}

		g.Set("msg", "access token is updated")
		g.JSON(http.StatusOK, webdto.NewAccessToken{AccessToken: access})
	}
}

// Authorization godoc
// @Summary      Авторизация пользователя
// @Description  Проверяет логин и пароль, выставляет refresh_token в Cookie и возвращает access-токен в ответе.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body   webdto.Auth  true  "Учетные данные пользователя"
// @Success      200      {object}  webdto.NewAccessToken  "Успешная авторизация, access-токен в теле ответа"
// @Failure      400      {object}  webdto.EmptyResponse   "Неверный формат запроса (bad request)"
// @Failure      401      {object}  webdto.EmptyResponse   "Неверные учетные данные"
// @Failure      500      {object}  webdto.EmptyResponse   "Внутренняя ошибка сервера"
// @Router       /secure/auth [post]
func (ah *authHandler) Authorization() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		var req webdto.Auth
		if err := g.ShouldBindJSON(&req); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		id, role, err := ah.u.CheckCredentials(ctx, req.Username, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrNotFound):
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusUnauthorized, webdto.EmptyResponse{Message: errs.ErrInvalidCredentials.Error()})
			case errors.Is(err, errs.ErrInvalidCredentials):
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusUnauthorized, webdto.EmptyResponse{Message: errs.ErrInvalidCredentials.Error()})
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternal.Error()})
			}
			return
		}

		access, err := ah.tm.GenerateAccessToken(id, role)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternal.Error()})
			return
		}

		refresh, err := ah.tm.GenerateRefreshToken(id, role)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternal.Error()})
			return
		}

		g.Set("msg", "auth successful")
		g.SetCookie("refresh_token", refresh, 3600*144, "/", "", false, true)
		g.JSON(http.StatusOK, webdto.NewAccessToken{AccessToken: access})
	}
}
