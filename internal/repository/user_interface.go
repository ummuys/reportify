package repository

import "context"

type UserDB interface {
	CreateUser(pCtx context.Context, username string, hashPassword string, role string) error
	CheckCredentials(pCtx context.Context, username string) (int64, string, string, error)
	Exists(pCtx context.Context, username string) error
}
