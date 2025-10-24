package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

type MetadataHandler interface {
	GetSchemas(pCtx context.Context) gin.HandlerFunc
	GetTables(pCtx context.Context) gin.HandlerFunc
	GetColumns(pCtx context.Context) gin.HandlerFunc
	GetQueries(pCtx context.Context) gin.HandlerFunc
	DeleteAllQueries(pCtx context.Context) gin.HandlerFunc
	DeleteQuery(pCtx context.Context) gin.HandlerFunc
}
