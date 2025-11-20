package service

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/dto"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/repository"
	"github.com/ummuys/reportify/internal/secure"
)

type admService struct {
	logger *zerolog.Logger       //
	db     repository.UserDB     // mocks.MockUDB
	ph     secure.PasswordHasher // mocks.MockHasher
}

func NewAdminService(logger *zerolog.Logger, db repository.UserDB, ph secure.PasswordHasher) AdminService {
	return &admService{logger: logger, db: db, ph: ph}
}

func (a *admService) CreateUser(pCtx context.Context, userInfo dto.CreateUser) error {
	a.logger.Debug().Str("evt", "call CreateUser").Msg("")

	err := a.db.ValidateRole(pCtx, userInfo.Role)
	if err != nil {
		return errs.ParsePgError(err)
	}

	userInfo.Password, err = a.ph.Hash(userInfo.Password)
	if err != nil {
		return err
	}

	if err := a.db.CreateUser(pCtx, userInfo); err != nil {
		return errs.ParsePgError(err)
	}

	return nil
}

func (a *admService) UpdateUser(pCtx context.Context, userInfo dto.UpdateUser) error {
	a.logger.Debug().Str("evt", "call CreateUser").Msg("")

	var err error
	if userInfo.Password != "" {
		userInfo.Password, err = a.ph.Hash(userInfo.Password)
		if err != nil {
			return err
		}
	}

	if userInfo.Role != "" {
		err = a.db.ValidateRole(pCtx, userInfo.Role)
		if err != nil {
			return errs.ParsePgError(err)
		}
	}

	if err := a.db.UpdateUser(pCtx, userInfo); err != nil {
		return errs.ParsePgError(err)
	}

	return nil
}

func (a *admService) DeleteUser(pCtx context.Context, userInfo dto.DeleteUser) error {
	a.logger.Debug().Str("evt", "call DeleteUser").Msg("")

	if err := a.db.DeleteUser(pCtx, userInfo); err != nil {
		return errs.ParsePgError(err)
	}

	return nil
}

func (a *admService) GetUsers(pCtx context.Context) ([]dto.GetUser, error) {
	a.logger.Debug().Str("evt", "call GetUsers").Msg("")

	users, err := a.db.GetUsers(pCtx)
	if err != nil {
		return nil, errs.ParsePgError(err)
	}
	return users, nil
}
