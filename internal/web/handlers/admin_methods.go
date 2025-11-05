package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/service"
)

type admHandler struct {
	logger *zerolog.Logger
	adms   service.AdminService
}

func NewAdminHandler(logger *zerolog.Logger, adms service.AdminService) AdminHandler {
	return &admHandler{logger: logger, adms: adms}
}

func (a *admHandler) CreateUserWithRole() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		return
	}
}
func (a *admHandler) DeleteUser()           {}
func (a *admHandler) ChangePasswordToUser() {}
func (a *admHandler) ChangeRoleToUser()     {}
