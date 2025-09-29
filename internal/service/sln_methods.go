package service

import (
	"context"

	"github.com/rs/zerolog"
)

type repService struct {
	logger *zerolog.Logger
}

func NewReportService(logger *zerolog.Logger) RepService {
	return &repService{logger: logger}
}

func (rs *repService) CreateReport(pCtx context.Context, sql string) {
	rs.logger.Debug().Str("evt", "call CreateReport")
}
