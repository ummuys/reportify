package repository

import (
	"context"
	"time"

	"github.com/ummuys/reportify/internal/config"

	"github.com/jackc/pgx/v5/pgtype"
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
		for i, v := range vals {
			row[i] = convertPGIntoGo(v)
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return headers, data, rows.Err()
}

func convertPGIntoGo(v any) any {
	switch val := v.(type) {
	case nil:
		return nil

	case pgtype.Numeric:
		if f, err := val.Float64Value(); err == nil {
			return f.Float64
		}
		return nil

	case pgtype.Int2:
		if val.Valid {
			return val.Int16
		}
		return nil

	case pgtype.Int4:
		if val.Valid {
			return val.Int32
		}
		return nil

	case pgtype.Int8:
		if val.Valid {
			return val.Int64
		}
		return nil

	case pgtype.Bool:
		if val.Valid {
			return val.Bool
		}
		return nil

	case pgtype.Text:
		if val.Valid {
			return val.String
		}
		return nil

	case pgtype.Timestamp:
		if val.Valid {
			return val.Time
		}
		return nil

	case pgtype.Date:
		if val.Valid {
			return val.Time
		}
		return nil

	default:
		return v
	}
}
