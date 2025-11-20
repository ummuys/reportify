package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ummuys/reportify/internal/webdto"
)

func RequestParams(rawParams webdto.RawReportParams, toCreate bool) (webdto.ReportParams, error) {
	var (
		params webdto.ReportParams
		err    error
	)
	if rawParams.ReportName == "" {
		return webdto.ReportParams{}, errors.New("invalid report name")
	}
	params.ReportName = rawParams.ReportName

	if rawParams.ReportComm == "" {
		return webdto.ReportParams{}, errors.New("invalid report commentary")
	}
	params.ReportComm = rawParams.ReportComm

	if toCreate {
		diff := time.Since(rawParams.CreatedAt)
		if diff < 0 {
			diff = -diff
		}
		if diff > 5*time.Minute {
			return webdto.ReportParams{}, errors.New("invalid created report time")
		}
	}
	params.CreatedAt = rawParams.CreatedAt

	if err = checkQuery(rawParams.Sql); err != nil {
		return webdto.ReportParams{}, err
	}
	params.Sql = rawParams.Sql

	if rawParams.CSVSep != "" {
		params.CSVSep, err = checkSepCSV(rawParams.CSVSep)
		if err != nil {
			return webdto.ReportParams{}, err
		}
	}

	return params, nil
}

func checkSepCSV(rawSep string) (rune, error) {
	r := []rune(rawSep)
	if len(r) != 1 {
		return 0, fmt.Errorf("csv_sep must be exactly 1 character, got %q", rawSep)
	}
	sep := r[0]
	switch sep {
	case '\r', '\n', '"':
		return 0, fmt.Errorf("csv_sep cannot be a control or quote character: %q", sep)
	}
	return sep, nil
}

// ЗАМЕЧАНИЯ
// 1) Сейчас функция не пропустит в любой позиции цифру 1. Надо подумать как бы это исправить, так как это не позволяет сделать, к примеру, where id = 1;
// 2) Очень сильная блокировка не дает писать гибкие запросы. Строгую проверку, к примеру, можно убрать у людей, которые имеют токен повышенной возможности
// 3) select * from b where id = 1 -- проходит, а не должна
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
