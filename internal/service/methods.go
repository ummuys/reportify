package service

import (
	"context"
	"os"
	"sq/internal/cache"
	"sq/internal/convert"
	models "sq/internal/models/response/get"
	"sq/internal/repository"

	"github.com/rs/zerolog"
)

type repService struct {
	logger *zerolog.Logger
	db     repository.ReportDB
	conv   convert.ReportConvert
	chc    cache.ReportCache
}

func NewReportService(logger *zerolog.Logger, db repository.ReportDB,
	conv convert.ReportConvert, chc cache.ReportCache) ReportService {
	return &repService{logger: logger, db: db, conv: conv, chc: chc}
}

func (rs *repService) CreateReport(pCtx context.Context, sql string, f *os.File) error {
	rs.logger.Debug().Str("evt", "call CreateReport")

	headers, rows, err := rs.db.ExecQuery(pCtx, sql)
	if err != nil {
		return err
	}

	err = rs.conv.ToPDF(headers, rows, f)
	if err != nil {
		return err
	}

	return nil
}

func (rs *repService) GetSchemas(pCtx context.Context) (*models.ListSchemas, error) {
	rs.logger.Debug().Str("evt", "call GetSchemas")
	data, err := rs.db.GetSchemas(pCtx)

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

func (rs *repService) GetTables(pCtx context.Context, schemaName string) (*models.ListTables, error) {

	rs.logger.Debug().Str("evt", "call GetTables")
	data, err := rs.db.GetTables(pCtx, schemaName)
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

func (rs *repService) GetColumns(pCtx context.Context, schemaName string, tableName string) (*models.ListColumns, error) {
	rs.logger.Debug().Str("evt", "call GetTables")
	data, err := rs.db.GetColumns(pCtx, schemaName, tableName)
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
