package service

import "context"

type AdminService interface {
	CreateUser(pCtx context.Context, username, password, role string) error
	DeleteUser(pCtx context.Context, username string) error
	GetUsers(pCtx context.Context) ([][]string, error)
}
