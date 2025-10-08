package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sq/internal/cache"
	"sq/internal/convert"
	"sq/internal/models"
	"sq/internal/repository"
	"strings"

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

// ЗАМЕЧАНИЯ
// 1) Сейчас функция не пропустит в любой позиции цифру 1. Надо подумать как бы это исправить, так как это не позволяет сделать, к примеру, where id = 1;
// 2) Очень сильная блокировка не дает писать гибкие запросы. Строгую проверку, к примеру, можно убрать у людей, которые имеют токен повышенной возможности
// 3) select * from zopa where id = 1 -- проходит, а не должна
func checkQuery(query string) error {
	bannedWords := []string{
		"drop", "truncate", "delete", "update", "insert", "alter", "create",
		"union", "into", "load_file", "pg_catalog", "information_schema",
		"exec", "xp_cmdshell", "benchmark", "sleep", "1",
	}

	q := strings.ToLower(query)
	if strings.Contains(query, "--") {
		return errors.New("SQL injection / dangerous pattern detected: --")
	}

	checkList := make([]string, 0, len(bannedWords))
	for _, w := range bannedWords {
		checkList = append(checkList, `\b`+regexp.QuoteMeta(w)+`\b`)
	}

	reg := regexp.MustCompile("(?i)" + strings.Join(checkList, "|"))

	if find := reg.FindString(q); find != "" {
		return fmt.Errorf("SQL injection / dangerous pattern detected: %s", find)
	}

	return nil
}

func (rs *repService) CreateReport(pCtx context.Context, sql string, f *os.File) error {
	rs.logger.Debug().Str("evt", "call CreateReport")

	if err := checkQuery(sql); err != nil {
		return err
	}

	headers, rows, err := rs.db.ExecQuery(pCtx, sql)
	if err != nil {
		return err
	}

	err = rs.conv.ToPDF(headers, rows, f)
	if err != nil {
		return err
	}

	if err := rs.chc.SetQuery(pCtx, "good key", sql); err != nil {
		rs.logger.Error().Err(err).Str("script", sql).Msg("failed to save last query")
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

func (rs *repService) GetHashQuerys(pCtx context.Context, key string) ([]string, error) {
	rs.logger.Debug().Str("evt", "call GetHashQuerys")
	return rs.chc.GetQuerys(pCtx, key)
}
