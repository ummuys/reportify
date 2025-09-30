package repository

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type rDB struct {
	logger *zerolog.Logger
	conn   *pgx.Conn
	busy   atomic.Bool
}

func returnConnect(pCtx context.Context, r *rDB, dataConn string) error {

	if r.conn != nil {
		_ = r.conn.Close(pCtx)
		r.conn = nil
	}

	newConn, err := pgx.Connect(pCtx, dataConn)
	if err != nil {
		return err
	}
	r.conn = newConn
	return nil
}

func doNotDie(pCtx context.Context, r *rDB) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pCtx.Done():
			return
		case <-ticker.C:
			if r.busy.Load() {
				r.logger.Debug().Msg("skip ping")
				continue
			}

			if r.conn == nil {
				r.logger.Warn().Msg("no connection, will try to connect")
			} else {
				if err := r.conn.Ping(pCtx); err == nil {
					r.logger.Debug().Msg("ping ok")
					continue
				} else {
					r.logger.Error().Err(err).Msg("ping failed")
				}
			}

			dataConn := os.Getenv("DB_LINK")
			if dataConn == "" {
				r.logger.Error().Msg("DB_LINK is empty")
				continue
			}

			for {
				tryCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				err := returnConnect(tryCtx, r, dataConn)
				cancel()
				if err == nil {
					r.logger.Info().Msg("reconnected to DB")
					break
				}
				r.logger.Error().Err(err).Msg("reconnect failed, will retry")
				time.Sleep(time.Second * 3)
			}

		}
	}
}

func NewReportDB(pCtx context.Context, logger *zerolog.Logger) (ReportDB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var (
		conn *pgx.Conn
		err  error
	)

	for i := 0; i < 5; i++ {
		conn, err = pgx.Connect(ctx, os.Getenv("DB_LINK"))
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

	go doNotDie(pCtx, obj)

	return obj, nil

}

func (r *rDB) ExecQuery(pCtx context.Context, script string) ([]string, [][]any, error) {
	r.logger.Debug().Str("evt", "ExecQuery").Str("Query", script).Msg("")
	qCtx, cancel := context.WithTimeout(pCtx, time.Second*180)
	defer cancel()

	r.busy.Store(true)
	defer func() { r.busy.Store(false) }()
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
