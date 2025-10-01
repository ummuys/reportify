package service

import (
	"context"
	"os"
	models "sq/internal/models/response/get"
)

type ReportService interface {
	CreateReport(pCtx context.Context, sql string, f *os.File) error
	GetSchemas(pCtx context.Context) (*models.ListSchemas, error)
	GetTables(pCtx context.Context, schemaName string) (*models.ListTables, error)
	GetColumns(pCtx context.Context, schemaName string, tableName string) (*models.ListColumns, error)
}
