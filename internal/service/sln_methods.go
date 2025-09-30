package service

import (
	"context"
	"fmt"
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

func (rs *repService) CreateReport(pCtx context.Context, sql string) error {
	rs.logger.Debug().Str("evt", "call CreateReport")
	headers, data, err := rs.db.ExecQuery(pCtx, sql)
	if err != nil {
		return err
	}
	fmt.Println(headers)
	for _, values := range data {
		fmt.Println(values...)
	}
	return nil
}
