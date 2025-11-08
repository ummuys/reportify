package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/models"
	"github.com/ummuys/reportify/internal/service"
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

		var req models.CreateUserRequest
		if err := g.ShouldBindBodyWithJSON(&req); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		if err := a.srv.CreateUser(ctx, req.Username, req.Password, req.Role); err != nil {
			switch {
			case errors.Is(err, errs.ErrDuplicate):
				g.Set("msg", errs.ErrUserNotFound.Error())
				g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: errs.ErrUserNotFound.Error()})
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: errs.ErrInternal.Error()})
			}
			return

		}

		g.Set("msg", "user created")
		g.JSON(http.StatusOK, models.EmptyResponse{Message: "user created"})
	}
}

// err ok
func (a *admHandler) DeleteUser() gin.HandlerFunc {
	return func(g *gin.Context) {
		a.logger.Debug().Str("evt", "call DeleteUser").Msg("")
		username := g.Param("username")
		if username == "" {
			g.Set("msg", errs.ErrEmptyUsername.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: errs.ErrEmptyUsername.Error()})
			return
		}
		ctx := g.Request.Context()

		if err := a.srv.DeleteUser(ctx, username); err != nil {
			switch {
			case errors.Is(err, errs.ErrNotFound):
				g.Set("msg", errs.ErrUserNotFound.Error())
				g.AbortWithStatusJSON(http.StatusNotFound, models.EmptyResponse{Message: errs.ErrUserNotFound.Error()})
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			}
			return
		}

		g.Set("msg", "user deleted")
		g.JSON(http.StatusOK, models.EmptyResponse{Message: "user deleted"})
	}
}

func (a *admHandler) UpdateUser() gin.HandlerFunc {
	return func(g *gin.Context) {
		a.logger.Debug().Str("evt", "call UpdateUser").Msg("")
		ctx := g.Request.Context()

		var req models.UpdateUserRequest
		if err := g.ShouldBindBodyWithJSON(&req); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		if req.UserID <= 0 {
			g.Set("msg", "invalid UserID")
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			return
		}

		if err := a.srv.UpdateUser(ctx, req.UserID, req.Username, req.Password, req.Role); err != nil {
			switch {
			case errors.Is(err, errs.ErrNotFound):
				g.Set("msg", errs.ErrUserNotFound.Error())
				g.AbortWithStatusJSON(http.StatusNotFound, models.EmptyResponse{Message: errs.ErrUserNotFound.Error()})
			default:
				g.Set("msg", err.Error())
				g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			}
			return
		}

		g.Set("msg", "user info updated")
		g.JSON(http.StatusOK, models.EmptyResponse{Message: "user info updated"})

	}
}

// err ok
func (a *admHandler) GetUsers() gin.HandlerFunc {
	return func(g *gin.Context) {
		a.logger.Debug().Str("evt", "call GetUsers").Msg("")
		ctx := g.Request.Context()
		data, err := a.srv.GetUsers(ctx)

		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}

		var users models.GetUsersResponse
		users.List = make([]models.UserResponse, len(data))
		for i := 0; i < len(data); i++ {
			users.List[i].UserID = data[i][0].(int64)
			users.List[i].Username = data[i][1].(string)
			users.List[i].Role = data[i][2].(string)
		}

		g.Set("msg", "users info returned")
		g.JSON(http.StatusOK, users)
	}
}
