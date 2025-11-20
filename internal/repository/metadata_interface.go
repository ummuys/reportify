package repository

import "context"

type MetadataDB interface {
	// REPORT
	GetSchemas(pCtx context.Context) (map[string]string, error)
	GetTables(pCtx context.Context, schemaName string) (map[string]string, error)
	GetColumns(pCtx context.Context, schemaName, tableName string) (map[string]string, error)

	// QUERIES
	SetCacheQueries(pCtx context.Context, cache map[string][]byte) error
	GetCacheQueries(pCtx context.Context) (map[string][]byte, error)
}
