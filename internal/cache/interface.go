package cache

import "context"

type ReportCache interface {
	Set(pCtx context.Context, key string, value any) error
	Get(pCtx context.Context, key string) (string, error)
}
