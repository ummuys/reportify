package handlers

import (
	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	UpdateAccessToken() gin.HandlerFunc
	Authorization() gin.HandlerFunc
}
