package service

import (
	"context"
	"os"
)

type ReportService interface {
	CreateReport(pCtx context.Context, sql string, f *os.File) error
}
