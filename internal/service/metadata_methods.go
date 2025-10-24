package service

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/ummuys/reportify/internal/cache"
	"github.com/ummuys/reportify/internal/models"
	"github.com/ummuys/reportify/internal/repository"
)

type mdService struct {
	logger *zerolog.Logger
	db     repository.MetadataDB // mocks.MockMDDB
	chc    cache.ReportCache     // mocks.MockRepCache
}

func NewMetadataService(logger *zerolog.Logger, db repository.MetadataDB, chc cache.ReportCache) MetadataService {
	return &mdService{logger: logger, db: db, chc: chc}
}

func (mds *mdService) GetSchemas(pCtx context.Context) (*models.ListSchemas, error) {
	mds.logger.Debug().Str("evt", "call GetSchemas")
	data, err := mds.db.GetSchemas(pCtx)

	if err != nil {
		return nil, err
	}

	var ls models.ListSchemas
	ls.Schemas = make([]models.Schema, 0, len(data))
	for name, comm := range data {
		ls.Schemas = append(ls.Schemas, models.Schema{Name: name, Comment: comm})
	}
	return &ls, nil
}

func (mds *mdService) GetTables(pCtx context.Context, schemaName string) (*models.ListTables, error) {

	mds.logger.Debug().Str("evt", "call GetTables")
	data, err := mds.db.GetTables(pCtx, schemaName)
	if err != nil {
		return nil, err
	}

	var lt models.ListTables
	lt.Tables = make([]models.Table, 0, len(data))
	for name, comm := range data {
		lt.Tables = append(lt.Tables, models.Table{Name: name, Comment: comm})
	}
	return &lt, nil
}

func (mds *mdService) GetColumns(pCtx context.Context, schemaName string, tableName string) (*models.ListColumns, error) {
	mds.logger.Debug().Str("evt", "call GetColumns")
	data, err := mds.db.GetColumns(pCtx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	var lc models.ListColumns
	lc.Columns = make([]models.Column, 0, len(data))
	for name, comm := range data {
		lc.Columns = append(lc.Columns, models.Column{Name: name, Comment: comm})
	}
	return &lc, nil
}

func (mds *mdService) GetQueries(pCtx context.Context, key string) ([]string, error) {
	mds.logger.Debug().Str("evt", "call GetQueries")
	return mds.chc.Get(pCtx, key)
}

func (mds *mdService) DeleteAllQueries(pCtx context.Context, key string) error {
	mds.logger.Debug().Str("evt", "call DeleteAllQueries")
	return mds.chc.DeleteAll(pCtx, key)
}

func (mds *mdService) DeleteQuery(pCtx context.Context, key string, value string) error {
	mds.logger.Debug().Str("evt", "call DeleteQuery")
	return mds.chc.Delete(pCtx, key, value)
}
