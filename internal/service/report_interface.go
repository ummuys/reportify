package service

import (
	"context"
	"os"
	"sq/internal/models"
)

type ReportService interface {
	CreateReport(pCtx context.Context, sql string, f *os.File) error
	GetSchemas(pCtx context.Context) (*models.ListSchemas, error)
	GetTables(pCtx context.Context, schemaName string) (*models.ListTables, error)
	GetColumns(pCtx context.Context, schemaName string, tableName string) (*models.ListColumns, error)
	GetHashQuerys(pCtx context.Context, key string) ([]string, error)
}
