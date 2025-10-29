package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ReportHandler interface {
	CreateReport(pCtx context.Context) gin.HandlerFunc
}
