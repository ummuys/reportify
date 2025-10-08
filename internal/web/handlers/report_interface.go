package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ReportHandler interface {
	CreateReport(pCtx context.Context) gin.HandlerFunc
	GetSchemas(pCtx context.Context) gin.HandlerFunc
	GetTables(pCtx context.Context) gin.HandlerFunc
	GetColumns(pCtx context.Context) gin.HandlerFunc
	GetHashQuerys(pCtx context.Context) gin.HandlerFunc
}
