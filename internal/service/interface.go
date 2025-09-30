package service

import (
	"context"
	"os"
)

type ReportService interface {
	CreateReport(pCtx context.Context, sql string, f *os.File) error
	GetSchemas(pCtx context.Context) ([]string, error)
	GetTables(pCtx context.Context, schemaName string) ([]string, error)
	GetColumns(pCtx context.Context, schemaName string, tableName string) ([]string, error)
}
