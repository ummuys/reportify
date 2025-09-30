package repository

import (
	"context"
	"fmt"
	"math/rand"
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
	dsn    string
}

func healthy(pCtx context.Context, r *rDB) bool {
	if r.conn == nil {
		return false
	}
	pingCtx, cancel := context.WithTimeout(pCtx, 2*time.Second)
	defer cancel()
	return r.conn.Ping(pingCtx) == nil
}

func reconnect(pCtx context.Context, r *rDB) error {
	if r.conn != nil {
		_ = r.conn.Close(pCtx)
		r.conn = nil
	}

	newConn, err := pgx.Connect(pCtx, r.dsn)
	if err != nil {
		return err
	}
	r.conn = newConn
	return nil
}

func doNotDie(pCtx context.Context, r *rDB) {
	const (
		tick         = 10 * time.Second
		backoffStart = 1 * time.Second
		backoffMax   = 30 * time.Second
	)

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	backoff := backoffStart

	for {
		select {
		case <-pCtx.Done():
			return

		case <-ticker.C:
			if r.busy.Load() {
				r.logger.Debug().Msg("skip ping: busy")
				continue
			}

			if healthy(pCtx, r) {
				backoff = backoffStart
				continue
			}
			r.logger.Warn().Msg("db unhealthy or no connection, trying to reconnect")

			for {
				if err := reconnect(pCtx, r); err == nil {
					r.logger.Info().Msg("reconnected to DB")
					backoff = backoffStart
					break
				} else {
					r.logger.Error().Err(err).Msg("reconnect failed")
				}

				jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
				wait := backoff + jitter

				select {
				case <-pCtx.Done():
					return
				case <-time.After(wait):
				}

				if backoff < backoffMax {
					backoff *= 2
					if backoff > backoffMax {
						backoff = backoffMax
					}
				}
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
	dsn := os.Getenv("DB_LINK")

	for i := 0; i < 5; i++ {
		conn, err = pgx.Connect(ctx, dsn)
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
