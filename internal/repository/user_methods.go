package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ummuys/reportify/internal/config"
	"github.com/ummuys/reportify/internal/dto"

	"github.com/jackc/pgx/v5"
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

func (u *uDB) GetUsers(pCtx context.Context) ([]dto.GetUser, error) {
	u.logger.Debug().Str("evt", "call GetUsers").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()
	ctx.Done()

	rows, err := u.pool.Query(ctx, GetUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []dto.GetUser
	for rows.Next() {
		var u dto.GetUser
		err = rows.Scan(&u.UserID, &u.Username, &u.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (u *uDB) CreateUser(pCtx context.Context, userInfo dto.CreateUser) (err error) {
	u.logger.Debug().Str("evt", "call CreateUser").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	var tx pgx.Tx
	tx, err = u.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				u.logger.Error().Err(rbErr).Msg("rollback failed")
			}
		}
	}()

	_, err = tx.Exec(ctx, NewUserStep1, userInfo.Username, userInfo.Password)
	if err != nil {
		return
	}

	_, err = tx.Exec(ctx, NewUserStep2, userInfo.Username, userInfo.Role)
	if err != nil {
		return
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return
}

func (u *uDB) UpdateUser(pCtx context.Context, userInfo dto.UpdateUser) (err error) {
	u.logger.Debug().Str("evt", "call UpdateUser").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	var tx pgx.Tx
	tx, err = u.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				u.logger.Error().Err(rbErr).Msg("rollback failed")
			}
		}
	}()

	if userInfo.Username != "" {
		_, err = tx.Exec(ctx, UpdateUsername, userInfo.UserID, userInfo.Username)
		if err != nil {
			return
		}
	}

	if userInfo.Password != "" {
		_, err = tx.Exec(ctx, UpdateUserPassword, userInfo.UserID, userInfo.Password)
		if err != nil {
			return
		}
	}

	if userInfo.Role != "" {
		_, err = tx.Exec(ctx, UpdateUserRole, userInfo.UserID, userInfo.Role)
		if err != nil {
			return
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return
}

func (u *uDB) ValidateRole(pCtx context.Context, role string) error {
	u.logger.Debug().Str("evt", "call CheckRole").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	_, err := u.pool.Exec(ctx, CheckRole, role)
	return err
}

func (u *uDB) CheckCredentials(pCtx context.Context, username string) (dto.UserCredentials, error) {
	u.logger.Debug().Str("evt", "call CheckCredentials").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	var userInfo dto.UserCredentials
	err := u.pool.QueryRow(ctx, GetCredentials, username).Scan(&userInfo.UserID, &userInfo.Password, &userInfo.Role)
	return userInfo, err
}

func (u *uDB) DeleteUser(pCtx context.Context, userInfo dto.DeleteUser) error {
	u.logger.Debug().Str("evt", "call DeleteUser").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	res, err := u.pool.Exec(ctx, DeleteUser, userInfo.Username)
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return err
}
