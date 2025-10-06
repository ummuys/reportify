package service

import (
	"context"
	"os"
	"sq/internal/cache"
	"sq/internal/convert"
	models "sq/internal/models/response/get"
	"sq/internal/repository"
)

type mockRepService struct {
	db   repository.ReportDB
	conv convert.ReportConvert
	chc  cache.ReportCache
}

func newMockReportService(db repository.ReportDB,
	conv convert.ReportConvert, chc cache.ReportCache) ReportService {
	return &mockRepService{db: db, conv: conv, chc: chc}
}

func (mrs *mockRepService) CreateReport(pCtx context.Context, sql string, f *os.File) error {
	return nil
}

func (mrs *mockRepService) GetSchemas(pCtx context.Context) (*models.ListSchemas, error) {
	return nil, nil
}

func (mrs *mockRepService) GetTables(pCtx context.Context, schemaName string) (*models.ListTables, error) {
	return nil, nil
}

func (mrs *mockRepService) GetColumns(pCtx context.Context, schemaName string, tableName string) (*models.ListColumns, error) {
	return nil, nil
}

func (mrs *mockRepService) GetHashQuerys(pCtx context.Context, key string) ([]string, error) {
	return nil, nil
}
