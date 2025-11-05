package handlers

import (
	"github.com/gin-gonic/gin"
)

type AdminHandler interface {
	CreateUserWithRole() gin.HandlerFunc
	DeleteUser()
	ChangePasswordToUser()
	ChangeRoleToUser()
}
