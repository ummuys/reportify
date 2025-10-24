package service

import (
	"context"
	"os"

	"github.com/ummuys/reportify/internal/models"
)

type ReportService interface {
	CreateReport(pCtx context.Context, user_id int64, param models.ReportParams, f *os.File, format string) error
}
