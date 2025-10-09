package service

import (
	"context"
	"fmt"
	"sq/internal/repository"
	"sq/internal/secure"

	"github.com/rs/zerolog"
)

type uSrv struct {
	logger *zerolog.Logger
	db     repository.UserDB
	ph     secure.PasswordHasher
}

func NewUserService(logger *zerolog.Logger, db repository.UserDB, ph secure.PasswordHasher) UserService {
	return &uSrv{logger: logger, db: db, ph: ph}
}

func (u *uSrv) Create(pCtx context.Context, username, password string) error {
	u.logger.Debug().Str("evt", "call Create").Msg("")
	if err := u.db.Exists(pCtx, username); err != nil {
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

func (u *uSrv) CheckPass(pCtx context.Context, username, password string) error {
	u.logger.Debug().Str("evt", "call CheckPass").Msg("")
	if err := u.db.Exists(pCtx, username); err != nil {
		return err
	}

	hashPass, err := u.ph.Hash(password)
	if err != nil {
		return err
	}

	dbHashPass, err := u.db.GetPassword(pCtx, username)
	if err != nil {
		return err
	}

	if dbHashPass != hashPass {
		return fmt.Errorf("incorrect pass")
	}
	return nil
}
