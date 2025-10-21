package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ummuys/reportify/internal/config"
	"github.com/ummuys/reportify/internal/errs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type uDB struct {
	pool   *pgxpool.Pool
	logger *zerolog.Logger
}

func NewUserDB(pCtx context.Context, logger *zerolog.Logger) (UserDB, error) {

	ctx, cancel := context.WithTimeout(pCtx, time.Second*10)
	defer cancel()

	cfg, err := config.ParseUserDBEnv()
	if err != nil {
		return nil, err
	}

	conn, err := PoolFromConfig(ctx, cfg, "user")
	if err != nil {
		return nil, err
	}

	obj := &uDB{
		pool:   conn,
		logger: logger,
	}

	return obj, nil

}

func (u *uDB) CreateUser(pCtx context.Context, username string, hashPassword string) error {
	u.logger.Debug().Str("evt", "call Set").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	_, err := u.pool.Exec(ctx, NewUser, username, hashPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.ErrUsernameAlredyExists
		}
	}
	return err
}

func (u *uDB) GetPassword(pCtx context.Context, username string) (int64, string, error) {
	u.logger.Debug().Str("evt", "call Get").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	var (
		user_id int64
		pass    string
	)
	err := u.pool.QueryRow(ctx, GetPass, username).Scan(&user_id, &pass)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, pass, errs.ErrInvalidCredentials
		}
	}

	return user_id, pass, nil
}

func (u *uDB) Exists(pCtx context.Context, username string) error {
	u.logger.Debug().Str("evt", "call Exitsts").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	_, err := u.pool.Exec(ctx, CheckUset, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrInvalidCredentials
		}
	}
	return err
}
