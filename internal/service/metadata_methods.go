package service

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/cache"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/repository"
	"github.com/ummuys/reportify/internal/webdto"
)

type mdService struct {
	logger *zerolog.Logger
	db     repository.MetadataDB // mocks.MockMDDB
	chc    cache.ReportCache     // mocks.MockRepCache
}

func NewMetadataService(logger *zerolog.Logger, db repository.MetadataDB, chc cache.ReportCache) MetadataService {
	return &mdService{logger: logger, db: db, chc: chc}
}

func (mds *mdService) GetSchemas(pCtx context.Context) (*webdto.ListSchemas, error) {
	mds.logger.Debug().Str("evt", "call GetSchemas")
	data, err := mds.db.GetSchemas(pCtx)
	if err != nil {
		return nil, errs.ParsePgError(err)
	}

	var ls webdto.ListSchemas
	ls.Schemas = make([]webdto.Schema, 0, len(data))
	for name, comm := range data {
		ls.Schemas = append(ls.Schemas, webdto.Schema{Name: name, Comment: comm})
	}
	sort.Slice(ls.Schemas, func(i, j int) bool {
		return ls.Schemas[i].Name < ls.Schemas[j].Name
	})
	return &ls, nil
}

func (mds *mdService) GetTables(pCtx context.Context, schemaName string) (*webdto.ListTables, error) {
	mds.logger.Debug().Str("evt", "call GetTables")
	data, err := mds.db.GetTables(pCtx, schemaName)
	if err != nil {
		return nil, errs.ParsePgError(err)
	}

	var lt webdto.ListTables
	lt.Tables = make([]webdto.Table, 0, len(data))
	for name, comm := range data {
		lt.Tables = append(lt.Tables, webdto.Table{Name: name, Comment: comm})
	}
	sort.Slice(lt.Tables, func(i, j int) bool {
		return lt.Tables[i].Name < lt.Tables[j].Name
	})
	return &lt, nil
}

func (mds *mdService) GetColumns(pCtx context.Context, schemaName string, tableName string) (*webdto.ListColumns, error) {
	mds.logger.Debug().Str("evt", "call GetColumns")
	data, err := mds.db.GetColumns(pCtx, schemaName, tableName)
	if err != nil {
		return nil, errs.ParsePgError(err)
	}
	var lc webdto.ListColumns
	lc.Columns = make([]webdto.Column, 0, len(data))
	for name, comm := range data {
		lc.Columns = append(lc.Columns, webdto.Column{Name: name, Comment: comm})
	}
	sort.Slice(lc.Columns, func(i, j int) bool {
		return lc.Columns[i].Name < lc.Columns[j].Name
	})
	return &lc, nil
}

func (mds *mdService) GetQueries(pCtx context.Context, key string) ([][]byte, error) {
	mds.logger.Debug().Str("evt", "call GetQueries")
	return mds.chc.Get(pCtx, key)
}

func (mds *mdService) DeleteAllQueries(pCtx context.Context, key string) error {
	mds.logger.Debug().Str("evt", "call DeleteAllQueries")
	return mds.chc.DeleteAll(pCtx, key)
}

func (mds *mdService) DeleteQuery(pCtx context.Context, key string, value webdto.ReportParams) error {
	mds.logger.Debug().Str("evt", "call DeleteQuery")
	data := webdto.CacheValue(value)
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return mds.chc.Delete(pCtx, key, bytes)
}
