package service

import (
	"context"

	"github.com/ummuys/reportify/internal/dto"
)

type AdminService interface {
	CreateUser(pCtx context.Context, userInfo dto.CreateUser) error
	UpdateUser(pCtx context.Context, userInfo dto.UpdateUser) error
	DeleteUser(pCtx context.Context, userInfo dto.DeleteUser) error
	GetUsers(pCtx context.Context) ([]dto.GetUser, error)
}
