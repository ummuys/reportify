package repository

import "context"

type ReportDB interface {
	CreateReport(pCtx context.Context, script string) ([]string, [][]any, error)
}
