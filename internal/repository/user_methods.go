package repository

import (
	"context"
	"fmt"
	"sq/internal/config"
	"time"

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
	return err
}

func (u *uDB) GetPassword(pCtx context.Context, username string) (string, error) {
	u.logger.Debug().Str("evt", "call Get").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	var pass string
	err := u.pool.QueryRow(ctx, GetPass, username).Scan(&pass)
	return pass, err
}

func (u *uDB) Exists(pCtx context.Context, username string) error {
	u.logger.Debug().Str("evt", "call Exitsts").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	_, err := u.pool.Exec(ctx, CheckUset, username)
	return err
}
