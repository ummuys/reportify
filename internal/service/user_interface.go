package service

import "context"

type UserService interface {
	Create(pCtx context.Context, username, password string) error
	CheckPass(pCtx context.Context, username, password string) error
}
