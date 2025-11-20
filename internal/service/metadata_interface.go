package service

import (
	"context"

	"github.com/ummuys/reportify/internal/webdto"
)

type MetadataService interface {
	// DB
	GetSchemas(pCtx context.Context) (*webdto.ListSchemas, error)
	GetTables(pCtx context.Context, schemaName string) (*webdto.ListTables, error)
	GetColumns(pCtx context.Context, schemaName string, tableName string) (*webdto.ListColumns, error)

	// CACHE
	GetQueries(pCtx context.Context, key string) ([][]byte, error)
	DeleteAllQueries(pCtx context.Context, key string) error
	DeleteQuery(pCtx context.Context, key string, value webdto.ReportParams) error
}
