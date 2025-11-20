package handlers

import (
	"github.com/gin-gonic/gin"
)

type AdminHandler interface {
	GetUsers() gin.HandlerFunc
	UpdateUser() gin.HandlerFunc
	CreateUser() gin.HandlerFunc
	DeleteUser() gin.HandlerFunc
}
