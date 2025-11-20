package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/dto"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/service"
	"github.com/ummuys/reportify/internal/webdto"
)

type admHandler struct {
	logger *zerolog.Logger
	srv    service.AdminService
}

func NewAdminHandler(logger *zerolog.Logger, adms service.AdminService) AdminHandler {
	return &admHandler{logger: logger, srv: adms}
}

// err ok
func (a *admHandler) CreateUser() gin.HandlerFunc {
	return func(g *gin.Context) {
		a.logger.Debug().Str("evt", "call CreateUser").Msg("")
		ctx := g.Request.Context()

		var req webdto.CreateUserRequest
		if err := g.ShouldBindBodyWithJSON(&req); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		if err := a.srv.CreateUser(ctx, dto.CreateUser{
			Username: req.Username,
			Password: req.Password,
			Role:     req.Role,
		}); err != nil {
			switch {
			case errors.Is(err, errs.ErrDuplicate):
				g.Set("msg", errs.ErrUserNotFound.Error())
				g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrUserNotFound.Error()})
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternal.Error()})
			}
			return

		}

		g.Set("msg", "user created")
		g.JSON(http.StatusOK, webdto.EmptyResponse{Message: "user created"})
	}
}

// err ok
func (a *admHandler) DeleteUser() gin.HandlerFunc {
	return func(g *gin.Context) {
		a.logger.Debug().Str("evt", "call DeleteUser").Msg("")
		username := g.Param("username")
		if username == "" {
			g.Set("msg", errs.ErrEmptyUsername.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrEmptyUsername.Error()})
			return
		}
		ctx := g.Request.Context()

		if err := a.srv.DeleteUser(ctx, dto.DeleteUser{
			Username: username,
		}); err != nil {
			switch {
			case errors.Is(err, errs.ErrNotFound):
				g.Set("msg", errs.ErrUserNotFound.Error())
				g.AbortWithStatusJSON(http.StatusNotFound, webdto.EmptyResponse{Message: errs.ErrUserNotFound.Error()})
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			}
			return
		}

		g.Set("msg", "user deleted")
		g.JSON(http.StatusOK, webdto.EmptyResponse{Message: "user deleted"})
	}
}

func (a *admHandler) UpdateUser() gin.HandlerFunc {
	return func(g *gin.Context) {
		a.logger.Debug().Str("evt", "call UpdateUser").Msg("")
		ctx := g.Request.Context()

		var req webdto.UpdateUserRequest
		if err := g.ShouldBindBodyWithJSON(&req); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		if req.UserID <= 0 {
			g.Set("msg", "invalid UserID")
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		if err := a.srv.UpdateUser(ctx, dto.UpdateUser{
			UserID:   req.UserID,
			Username: req.Username,
			Password: req.Password,
			Role:     req.Role,
		}); err != nil {
			switch {
			case errors.Is(err, errs.ErrNotFound):
				g.Set("msg", errs.ErrUserNotFound.Error())
				g.AbortWithStatusJSON(http.StatusNotFound, webdto.EmptyResponse{Message: errs.ErrUserNotFound.Error()})
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			}
			return
		}

		g.Set("msg", "user info updated")
		g.JSON(http.StatusOK, webdto.EmptyResponse{Message: "user info updated"})
	}
}

func (a *admHandler) GetUsers() gin.HandlerFunc {
	return func(g *gin.Context) {
		a.logger.Debug().Str("evt", "call GetUsers").Msg("")
		ctx := g.Request.Context()
		users, err := a.srv.GetUsers(ctx)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}

		var resp webdto.GetUsersResponse
		resp.Users = make([]webdto.UserResponse, len(users))
		for i, u := range users {
			respU := webdto.UserResponse(u)
			resp.Users[i] = respU
		}

		g.Set("msg", "users info returned")
		g.JSON(http.StatusOK, users)
	}
}
