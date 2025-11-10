package service

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/ummuys/reportify/internal/cache"
	"github.com/ummuys/reportify/internal/convert"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/models"
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

func (rs *repService) CreateReport(pCtx context.Context, user_id int64, params models.ReportParams, f *os.File, format string) error {
	rs.logger.Debug().Str("evt", "call CreateReport")

	headers, rows, err := rs.db.CreateReport(pCtx, params.Sql)
	if err != nil {
		return errs.ParsePgError(err)
	}

	switch format {
	case "pdf":
		err = rs.conv.ToPDF(headers, rows, f)
	case "csv":
		err = rs.conv.ToCSV(headers, rows, f, params.CSVSep)
	case "xlsx":
		err = rs.conv.ToXLSX(headers, rows, f)
	case "chart":
		fallthrough
	case "json":
		err = rs.conv.ToJSON(headers, rows, f)
	case "docx":
		err = rs.conv.ToDOCX(headers, rows, f)
	}

	if err != nil {
		return err
	}

	data := models.CacheValue(params)
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if format != "chart" {
		if err := rs.chc.Set(pCtx, strconv.FormatInt(user_id, 10), bytes); err != nil {
			rs.logger.Error().Err(err).Msg("failed to save last query")
		}
	}

	return nil
}
