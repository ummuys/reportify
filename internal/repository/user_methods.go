package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ummuys/reportify/internal/config"

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

func (u *uDB) GetUsers(pCtx context.Context) ([][]any, error) {
	u.logger.Debug().Str("evt", "call GetUsers").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()
	ctx.Done()

	rows, err := u.pool.Query(ctx, GetUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data [][]any
	for rows.Next() {
		var (
			id       int64
			username string
			role     string
		)
		err = rows.Scan(&id, &username, &role)
		if err != nil {
			return nil, err
		}
		data = append(data, []any{id, username, role})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func (u *uDB) CreateUser(pCtx context.Context, username string, hashPassword string, role string) (err error) {
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

	_, err = tx.Exec(ctx, NewUserStep1, username, hashPassword)
	if err != nil {
		return
	}

	_, err = tx.Exec(ctx, NewUserStep2, username, role)
	if err != nil {
		return
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return
}

func (u *uDB) UpdateUser(pCtx context.Context, userID int64, username string, hashPassword string, role string) (err error) {
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

	if username != "" {
		_, err = tx.Exec(ctx, UpdateUsername, userID, username)
		if err != nil {
			return
		}
	}

	if hashPassword != "" {
		_, err = tx.Exec(ctx, UpdateUserPassword, userID, hashPassword)
		if err != nil {
			return
		}
	}

	if role != "" {
		_, err = tx.Exec(ctx, UpdateUserRole, userID, role)
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

func (u *uDB) CheckCredentials(pCtx context.Context, username string) (int64, string, string, error) {
	u.logger.Debug().Str("evt", "call CheckCredentials").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	var (
		user_id int64
		role    string
		pass    string
	)
	err := u.pool.QueryRow(ctx, GetCredentials, username).Scan(&user_id, &pass, &role)
	return user_id, role, pass, err
}

func (u *uDB) DeleteUser(pCtx context.Context, username string) error {
	u.logger.Debug().Str("evt", "call DeleteUser").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	res, err := u.pool.Exec(ctx, DeleteUser, username)
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return err
}
