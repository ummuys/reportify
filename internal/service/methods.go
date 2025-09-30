package service

import (
	"context"
	"os"
	"sq/internal/convert"
	"sq/internal/repository"

	"github.com/rs/zerolog"
)

type repService struct {
	logger *zerolog.Logger
	db     repository.ReportDB
	conv   convert.ReportConvert
}

func NewReportService(logger *zerolog.Logger, db repository.ReportDB, conv convert.ReportConvert) ReportService {
	return &repService{logger: logger, db: db, conv: conv}
}

func (rs *repService) CreateReport(pCtx context.Context, sql string, f *os.File) error {
	rs.logger.Debug().Str("evt", "call CreateReport")

	headers, rows, err := rs.db.ExecQuery(pCtx, sql)
	if err != nil {
		return err
	}

	err = rs.conv.ToPDF(headers, rows, f)
	if err != nil {
		return err
	}

	return nil
}

func (rs *repService) GetSchemas(pCtx context.Context) ([]string, error) {
	rs.logger.Debug().Str("evt", "call GetSchemas")
	data, err := rs.db.GetSchemas(pCtx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (rs *repService) GetTables(pCtx context.Context, schemaName string) ([]string, error) {
	rs.logger.Debug().Str("evt", "call GetTables")
	data, err := rs.db.GetTables(pCtx, schemaName)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (rs *repService) GetColumns(pCtx context.Context, schemaName string, tableName string) ([]string, error) {
	rs.logger.Debug().Str("evt", "call GetTables")
	data, err := rs.db.GetColumns(pCtx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	return data, nil
}
