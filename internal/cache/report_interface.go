package cache

import "context"

type ReportCache interface {
	WarmUp(pCtx context.Context, queries map[string][]string) error
	SetQuery(pCtx context.Context, key string, value any) error
	GetQueries(pCtx context.Context, key string) ([]string, error)
	GetCacheQueries(pCtx context.Context) (map[string][]string, error)
}
