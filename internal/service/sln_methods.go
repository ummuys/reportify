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
