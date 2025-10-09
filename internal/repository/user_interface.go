package repository

import "context"

type UserDB interface {
	CreateUser(pCtx context.Context, username string, hashPassword string) error
	GetPassword(pCtx context.Context, username string) (string, error)
	Exists(pCtx context.Context, username string) error
}
