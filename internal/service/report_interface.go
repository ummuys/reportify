package service

import (
	"context"
	"os"
)

type ReportService interface {
	CreateReport(pCtx context.Context, user_id int64, sql string, f *os.File) error
}
