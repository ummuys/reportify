package repository

import (
	"context"
	"errors"
	"fmt"
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

	cfg, err := config.ParseUDDBeEnv()
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.Addr)
	if err != nil {
		return nil, err
	}
	poolCfg.MinConns = cfg.MinConn
	poolCfg.MaxConns = cfg.MaxConn
	poolCfg.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Second
	poolCfg.MaxConnLifetimeJitter = time.Duration(cfg.MaxConnLifetimeJitter) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Second
	poolCfg.HealthCheckPeriod = time.Duration(cfg.HealthCheckPeriod) * time.Second

	var conn *pgxpool.Pool
	for i := 0; i < 5; i++ {
		conn, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if err != nil {
		return nil, fmt.Errorf("can't connect to db: %w", err)
	}

	if err = conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db didn't pinged: %w", err)
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

func (u *uDB) SetCacheQueries(pCtx context.Context, cache map[string][]string) error {
	u.logger.Debug().Str("evt", "call SaveCacheQuerys").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()

	b := &pgx.Batch{}
	br := u.pool.SendBatch(ctx, b)
	defer br.Close()
	for range cache {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (u *uDB) GetCacheQueries(pCtx context.Context) (map[string][]string, error) {
	u.logger.Debug().Str("evt", "call SaveCacheQuerys").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()

	rows, err := u.pool.Query(ctx, qGetCacheQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string][]string)
	for rows.Next() {
		var (
			key   string
			value []string
		)
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		m[key] = value
	}
	return m, nil
}
