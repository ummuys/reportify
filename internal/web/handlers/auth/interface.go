package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	UpdateRefreshToken(pCtx context.Context) gin.HandlerFunc
	UpdateAccessToken(pCtx context.Context) gin.HandlerFunc
	Authorization(pCtx context.Context) gin.HandlerFunc
}
