package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

type RepHandler interface {
	CreateReport(pCtx context.Context) gin.HandlerFunc
}
