package service

import (
	"context"

	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/repository"
	"github.com/ummuys/reportify/internal/secure"

	"github.com/rs/zerolog"
)

type uSrv struct {
	logger *zerolog.Logger       //
	db     repository.UserDB     // mocks.MockUDB
	ph     secure.PasswordHasher // mocks.MockHasher
}

func NewUserService(logger *zerolog.Logger, db repository.UserDB, ph secure.PasswordHasher) UserService {
	return &uSrv{logger: logger, db: db, ph: ph}
}

func (u *uSrv) Create(pCtx context.Context, username, password string) error {
	u.logger.Debug().Str("evt", "call Create").Msg("")

	err := u.db.Exists(pCtx, username)
	if err != nil {
		return err
	}

	hashPass, err := u.ph.Hash(password)
	if err != nil {
		return err
	}

	if err := u.db.CreateUser(pCtx, username, hashPass); err != nil {
		return err
	}

	return nil
}

func (u *uSrv) CheckPass(pCtx context.Context, username, password string) (int64, error) {
	u.logger.Debug().Str("evt", "call CheckPass").Msg("")

	err := u.db.Exists(pCtx, username)
	if err != nil {
		return 0, err
	}

	user_id, hashPass, err := u.db.GetPassword(pCtx, username)
	if err != nil {
		return 0, err
	}

	if !u.ph.ChechHash(password, hashPass) {
		return 0, errs.ErrInvalidCredentials
	}

	return user_id, nil
}
