package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ummuys/reportify/internal/config"
)

func PoolFromConfig(ctx context.Context, config config.DBConfig, dbName string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(config.Addr)
	if err != nil {
		return nil, err
	}
	poolCfg.MinConns = config.MinConn
	poolCfg.MaxConns = config.MaxConn
	poolCfg.MaxConnLifetime = config.MaxConnLifetime
	poolCfg.MaxConnLifetimeJitter = config.MaxConnLifetimeJitter
	poolCfg.MaxConnIdleTime = config.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = config.HealthCheckPeriod

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
		return nil, fmt.Errorf("%s_db didn't pinged: %w", dbName, err)
	}
	return conn, nil
}
