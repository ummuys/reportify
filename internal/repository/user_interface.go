package repository

import (
	"context"

	"github.com/ummuys/reportify/internal/dto"
)

type UserDB interface {
	GetUsers(pCtx context.Context) ([]dto.GetUser, error)
	CreateUser(pCtx context.Context, userInfo dto.CreateUser) (err error)
	UpdateUser(pCtx context.Context, userInfo dto.UpdateUser) (err error)
	DeleteUser(pCtx context.Context, userInfo dto.DeleteUser) error
	CheckCredentials(pCtx context.Context, username string) (dto.UserCredentials, error)
	ValidateRole(pCtx context.Context, role string) error
}
