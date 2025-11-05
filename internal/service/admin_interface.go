package service

import "context"

type AdminService interface {
	CreateUser(pCtx context.Context, username, password, role string) error
}
