package service

import "context"

type ReportService interface {
	CreateReport(pCtx context.Context, sql string) error
}
