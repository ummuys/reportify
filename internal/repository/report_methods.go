package repository

import (
	"context"
	"time"

	"github.com/ummuys/reportify/internal/config"

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

	cfg, err := config.ParseReportDBEnv()
	if err != nil {
		return nil, err
	}

	conn, err := PoolFromConfig(ctx, cfg, "report")
	if err != nil {
		return nil, err
	}

	obj := &rDB{
		pool:   conn,
		logger: logger,
	}

	return obj, nil

}

func (r *rDB) CreateReport(pCtx context.Context, script string) ([]string, [][]any, error) {
	r.logger.Debug().Str("evt", "CreateReport").Str("Query", script).Msg("")
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
