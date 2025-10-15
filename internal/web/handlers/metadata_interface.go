package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

type MetadataHandler interface {
	GetSchemas(pCtx context.Context) gin.HandlerFunc
	GetTables(pCtx context.Context) gin.HandlerFunc
	GetColumns(pCtx context.Context) gin.HandlerFunc
	GetCacheQueries(pCtx context.Context) gin.HandlerFunc
}
