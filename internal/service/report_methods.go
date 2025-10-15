package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ummuys/reportify/internal/cache"
	"github.com/ummuys/reportify/internal/convert"
	"github.com/ummuys/reportify/internal/repository"

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

func (rs *repService) CreateReport(pCtx context.Context, user_id int64, sql string, f *os.File) error {
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

	if err := rs.chc.SetQuery(pCtx, strconv.FormatInt(user_id, 10), sql); err != nil {
		rs.logger.Error().Err(err).Str("script", sql).Msg("failed to save last query")
	}

	return nil
}
