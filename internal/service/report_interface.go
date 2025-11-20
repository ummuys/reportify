package service

import (
	"context"

	"github.com/ummuys/reportify/internal/dto"
)

type ReportService interface {
	CreateReport(pCtx context.Context, reportInfo dto.CreateReport) error
}
