package repository

import (
	"context"
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

func (u *uDB) GetUsers(pCtx context.Context) ([][]string, error) {
	u.logger.Debug().Str("evt", "call GetUsers").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()
	ctx.Done()

	rows, err := u.pool.Query(ctx, GetUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data [][]string
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make([]string, 2)
		row[0], row[1] = vals[0].(string), vals[1].(string)
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func (u *uDB) CreateUser(pCtx context.Context, username string, hashPassword string, role string) error {
	u.logger.Debug().Str("evt", "call CreateUser").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	_, err := u.pool.Exec(ctx, NewUserStep1, username, hashPassword)
	if err != nil {
		return err
	}
	_, err = u.pool.Exec(ctx, NewUserStep2, username, role)
	return err
}

func (u *uDB) CheckRole(pCtx context.Context, role string) error {
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

func (u *uDB) CheckUser(pCtx context.Context, username string) error {
	u.logger.Debug().Str("evt", "call CheckUser").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	_, err := u.pool.Exec(ctx, CheckUser, username)
	return err
}
