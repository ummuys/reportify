package repository

import "context"

type ReportDB interface {
	ExecQuery(pCtx context.Context, script string) ([]string, [][]any, error)
	GetSchemas(pCtx context.Context) ([]string, error)
	GetTables(pCtx context.Context, schemaName string) ([]string, error)
	GetColumns(pCtx context.Context, schemaName, tableName string) ([]string, error)
}
