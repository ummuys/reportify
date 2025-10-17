package repository

import "context"

type ReportDB interface {
	CreateReport(pCtx context.Context, script string) ([]string, [][]any, error)
	GetSchemas(pCtx context.Context) (map[string]string, error)
	GetTables(pCtx context.Context, schemaName string) (map[string]string, error)
	GetColumns(pCtx context.Context, schemaName, tableName string) (map[string]string, error)
}
