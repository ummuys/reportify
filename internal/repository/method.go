package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type rDB struct {
	logger *zerolog.Logger
	conn   *pgxpool.Pool
	dsn    string
}

func NewReportDB(pCtx context.Context, logger *zerolog.Logger) (ReportDB, error) {

	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(os.Getenv("DB_LINK"))
	if err != nil {
		return nil, err
	}
	cfg.MinConns = 2
	cfg.MaxConns = 16
	cfg.MaxConnLifetime = 45 * time.Minute
	cfg.MaxConnLifetimeJitter = 5 * time.Minute
	cfg.MaxConnIdleTime = 2 * time.Minute

	var conn *pgxpool.Pool
	for i := 0; i < 5; i++ {
		conn, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		return nil, fmt.Errorf("can't connect to db: %w", err)
	}

	if err = conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db didn't pinged: %w", err)
	}

	obj := &rDB{
		conn:   conn,
		logger: logger,
	}

	return obj, nil

}

func (r *rDB) ExecQuery(pCtx context.Context, script string) ([]string, [][]any, error) {
	r.logger.Debug().Str("evt", "ExecQuery").Str("Query", script).Msg("")
	qCtx, cancel := context.WithTimeout(pCtx, time.Second*180)
	defer cancel()

	if err := r.conn.Ping(pCtx); err != nil {
		return nil, nil, fmt.Errorf("db didn't pinged: %w", err)
	}

	rows, err := r.conn.Query(qCtx, script)
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

	if err := r.conn.Ping(pCtx); err != nil {
		return nil, fmt.Errorf("db didn't pinged: %w", err)
	}

	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	rows, err := r.conn.Query(ctx, qSchemaWithComment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackingRows(rows)
}

func (r *rDB) GetTables(pCtx context.Context, schemaName string) (map[string]string, error) {
	r.logger.Debug().Str("evt", "GetTables").Msg("")

	if err := r.conn.Ping(pCtx); err != nil {
		return nil, fmt.Errorf("db didn't pinged: %w", err)
	}

	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	rows, err := r.conn.Query(ctx, qTablesWithComment, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackingRows(rows)

}

func (r *rDB) GetColumns(ctx context.Context, schemaName, tableName string) (map[string]string, error) {

	if err := r.conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db didn't pinged: %w", err)
	}

	qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.conn.Query(qctx, qColumnsWithComment, schemaName, tableName)
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
