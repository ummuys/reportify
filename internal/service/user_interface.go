package service

import "context"

type UserService interface {
	CheckCredentials(pCtx context.Context, username, password string) (int64, string, error)
}
