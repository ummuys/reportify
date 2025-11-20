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

func (u *uSrv) CheckCredentials(pCtx context.Context, username, password string) (int64, string, error) {
	u.logger.Debug().Str("evt", "call CheckCredentials").Msg("")

	userInfo, err := u.db.CheckCredentials(pCtx, username)
	if err != nil {
		return 0, "", errs.ParsePgError(err)
	}

	if !u.ph.CheckHash(password, userInfo.Password) {
		return 0, "", errs.ErrInvalidCredentials
	}

	return userInfo.UserID, userInfo.Role, nil
}
