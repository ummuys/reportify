package repository

import "context"

type ReportDB interface {
	ExecQuery(pCtx context.Context, script string) ([]string, [][]any, error)
}
