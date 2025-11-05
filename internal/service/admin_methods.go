package service

import (
	"context"

	"github.com/rs/zerolog"
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

func (a *admService) CreateUser(pCtx context.Context, username, password, role string) error {
	a.logger.Debug().Str("evt", "call Create").Msg("")

	err := a.db.Exists(pCtx, username)
	if err != nil {
		return err
	}

	hashPass, err := a.ph.Hash(password)
	if err != nil {
		return err
	}

	if err := a.db.CreateUser(pCtx, username, hashPass, role); err != nil {
		return err
	}
	return nil
}
