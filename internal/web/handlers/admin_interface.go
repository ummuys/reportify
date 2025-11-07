package handlers

import (
	"github.com/gin-gonic/gin"
)

type AdminHandler interface {
	GetUsers() gin.HandlerFunc
	CreateUser() gin.HandlerFunc
	DeleteUser() gin.HandlerFunc
	ChangeUserPassword() gin.HandlerFunc
	ChangeUserRole() gin.HandlerFunc
}
