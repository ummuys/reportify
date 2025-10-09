package repository

import (
	"context"
	"fmt"
	"sq/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type rDB struct {
	logger *zerolog.Logger
	pool   *pgxpool.Pool
}

func NewReportDB(pCtx context.Context, logger *zerolog.Logger) (ReportDB, error) {

	ctx, cancel := context.WithTimeout(pCtx, time.Second*10)
	defer cancel()

	cfg, err := config.ParseRepDBeEnv()
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

	var pool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if err != nil {
		return nil, fmt.Errorf("can't connect to db: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db didn't pinged: %w", err)
	}

	obj := &rDB{
		pool:   pool,
		logger: logger,
	}

	return obj, nil

}

func (r *rDB) ExecQuery(pCtx context.Context, script string) ([]string, [][]any, error) {
	r.logger.Debug().Str("evt", "ExecQuery").Str("Query", script).Msg("")
	qCtx, cancel := context.WithTimeout(pCtx, time.Second*180)
	defer cancel()

	rows, err := r.pool.Query(qCtx, script)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	fds := rows.FieldDescriptions()
	headers := make([]string, len(fds))
	for i := range fds {
		headers[i] = string(fds[i].Name)
	}

	var data [][]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, nil, err
		}

		row := make([]any, len(vals))
		copy(row, vals)
		data = append(data, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return headers, data, nil
}

func (r *rDB) GetSchemas(pCtx context.Context) (map[string]string, error) {
	r.logger.Debug().Str("evt", "GetSchemas").Msg("")

	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	rows, err := r.pool.Query(ctx, qSchemaWithComment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackingRows(rows)
}

func (r *rDB) GetTables(pCtx context.Context, schemaName string) (map[string]string, error) {
	r.logger.Debug().Str("evt", "GetTables").Msg("")

	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	rows, err := r.pool.Query(ctx, qTablesWithComment, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackingRows(rows)

}

func (r *rDB) GetColumns(ctx context.Context, schemaName, tableName string) (map[string]string, error) {
	r.logger.Debug().Str("evt", "GetColumns").Msg("")

	qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(qctx, qColumnsWithComment, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackingRows(rows)
}

func unpackingRows(rows pgx.Rows) (map[string]string, error) {
	res := make(map[string]string)
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		res[key] = value
	}

	return res, rows.Err()
}
