package cache

import "context"

type ReportCache interface {
	SetQuery(pCtx context.Context, key string, value any) error
	GetQuerys(pCtx context.Context, key string) ([]string, error)
}
