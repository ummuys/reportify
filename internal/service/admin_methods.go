package service

import (
	"context"

	"github.com/rs/zerolog"
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

// TO FIX: MAKE BETTER ERR CHECKER
func (a *admService) CreateUser(pCtx context.Context, username, password, role string) error {
	a.logger.Debug().Str("evt", "call CreateUser").Msg("")

	err := a.db.CheckRole(pCtx, role)
	if err != nil {
		return errs.ParsePgError(err)
	}

	hashPass, err := a.ph.Hash(password)
	if err != nil {
		return err
	}

	if err := a.db.CreateUser(pCtx, username, hashPass, role); err != nil {
		return errs.ParsePgError(err)
	}

	return nil
}

// TO FIX: MAKE BETTER ERR CHECKER
func (a *admService) DeleteUser(pCtx context.Context, username string) error {
	a.logger.Debug().Str("evt", "call DeleteUser").Msg("")

	if err := a.db.DeleteUser(pCtx, username); err != nil {
		return errs.ParsePgError(err)
	}

	return nil
}

// TO FIX: MAKE BETTER ERR CHECKER
func (a *admService) GetUsers(pCtx context.Context) ([][]string, error) {
	a.logger.Debug().Str("evt", "call GetUsers").Msg("")

	list, err := a.db.GetUsers(pCtx)
	if err != nil {
		return nil, errs.ParsePgError(err)
	}
	return list, nil
}
