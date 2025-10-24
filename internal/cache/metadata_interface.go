package cache

import "context"

type ReportCache interface {
	Init(pCtx context.Context, queries map[string][]string) error
	Set(pCtx context.Context, key string, value any) error
	Get(pCtx context.Context, key string) ([]string, error)
	GetAll(pCtx context.Context) (map[string][]string, error)
	Delete(pCtx context.Context, key string, value any) error
	DeleteAll(pCtx context.Context, key string) error
}
