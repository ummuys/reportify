package service

import (
	"context"

	"github.com/ummuys/reportify/internal/models"
)

type MetadataService interface {
	GetSchemas(pCtx context.Context) (*models.ListSchemas, error)
	GetTables(pCtx context.Context, schemaName string) (*models.ListTables, error)
	GetColumns(pCtx context.Context, schemaName string, tableName string) (*models.ListColumns, error)
	GetCacheQueries(pCtx context.Context, key string) ([]string, error)
}
