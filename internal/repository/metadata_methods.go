package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/config"
)

type mdDB struct {
	logger *zerolog.Logger
	rPool  *pgxpool.Pool
	uPool  *pgxpool.Pool
}

func NewMetadataDB(pCtx context.Context, logger *zerolog.Logger) (MetadataDB, error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Second*10)
	defer cancel()

	rCfg, err := config.ParseReportDBEnv()
	if err != nil {
		return nil, err
	}

	rConn, err := PoolFromConfig(ctx, rCfg, "metadata")
	if err != nil {
		return nil, err
	}

	uCfg, err := config.ParseUserDBEnv()
	if err != nil {
		return nil, err
	}

	uConn, err := PoolFromConfig(ctx, uCfg, "metadata")
	if err != nil {
		return nil, err
	}

	obj := &mdDB{
		rPool:  rConn,
		uPool:  uConn,
		logger: logger,
	}

	return obj, nil
}

func (m *mdDB) GetSchemas(pCtx context.Context) (map[string]string, error) {
	m.logger.Debug().Str("evt", "GetSchemas").Msg("")

	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	rows, err := m.rPool.Query(ctx, qSchemaWithComment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackingRows(rows)
}

func (m *mdDB) GetTables(pCtx context.Context, schemaName string) (map[string]string, error) {
	m.logger.Debug().Str("evt", "GetTables").Msg("")

	ctx, cancel := context.WithTimeout(pCtx, time.Second*2)
	defer cancel()

	rows, err := m.rPool.Query(ctx, qTablesWithComment, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return unpackingRows(rows)
}

func (m *mdDB) GetColumns(pCtx context.Context, schemaName, tableName string) (map[string]string, error) {
	m.logger.Debug().Str("evt", "call GetColumns").Msg("")

	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()

	rows, err := m.rPool.Query(ctx, qColumnsWithComment, schemaName, tableName)
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

func (m *mdDB) SetCacheQueries(pCtx context.Context, cache map[string][]byte) (err error) {
	m.logger.Debug().Str("evt", "call SetCacheQuerys").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, 30*time.Second)
	defer cancel()

	var rows pgx.Rows
	allID := `select user_id from identity.users`
	rows, err = m.uPool.Query(ctx, allID)
	if err != nil {
		return err
	}
	defer rows.Close()

	usersID := make([]string, 0, 64)
	for rows.Next() {
		var uid string
		if err = rows.Scan(&uid); err != nil {
			return err
		}
		usersID = append(usersID, uid)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	var tx pgx.Tx
	tx, err = m.uPool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				m.logger.Error().Err(rbErr).Msg("rollback failed")
			}
		}
	}()

	b := &pgx.Batch{}

	for _, uid := range usersID {
		val, ok := cache[uid]
		if !ok {
			b.Queue(qSetCacheQuery, uid, []byte("[]"))
		} else {
			b.Queue(qSetCacheQuery, uid, val)
		}
	}

	br := tx.SendBatch(ctx, b)

	for range usersID {
		if _, err = br.Exec(); err != nil {
			_ = br.Close()
			return err
		}
	}

	if err = br.Close(); err != nil {
		return
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return nil
}

func (m *mdDB) GetCacheQueries(pCtx context.Context) (map[string][]byte, error) {
	m.logger.Debug().Str("evt", "call SaveCacheQuerys").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()

	rows, err := m.uPool.Query(ctx, qGetCacheQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quer := make(map[string][]byte)
	for rows.Next() {
		var (
			key   string
			value []byte
		)
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		quer[key] = value
	}
	return quer, nil
}
