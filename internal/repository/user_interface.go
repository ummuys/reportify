package repository

import "context"

type UserDB interface {
	CreateUser(pCtx context.Context, username string, hashPassword string) error
	GetPassword(pCtx context.Context, username string) (int64, string, error)
	Exists(pCtx context.Context, username string) error

	SetCacheQueries(pCtx context.Context, cache map[string][]string) error
	GetCacheQueries(pCtx context.Context) (map[string][]string, error)
}
