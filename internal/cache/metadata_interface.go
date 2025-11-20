package cache

import "context"

type ReportCache interface {
	Init(pCtx context.Context, queries map[string][]byte) error
	Set(pCtx context.Context, key string, value []byte) error
	Get(pCtx context.Context, key string) ([][]byte, error)
	GetAll(pCtx context.Context) (map[string][]byte, error)
	Delete(pCtx context.Context, key string, value []byte) error
	DeleteAll(pCtx context.Context, key string) error
}
