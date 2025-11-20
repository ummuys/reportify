package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/ummuys/reportify/internal/cache"
	"github.com/ummuys/reportify/internal/convert"
	"github.com/ummuys/reportify/internal/dto"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/repository"
	"github.com/ummuys/reportify/internal/webdto"

	"github.com/rs/zerolog"
)

type repService struct {
	logger *zerolog.Logger
	db     repository.ReportDB
	conv   convert.ReportConvert
	chc    cache.ReportCache
}

func NewReportService(logger *zerolog.Logger, db repository.ReportDB,
	conv convert.ReportConvert, chc cache.ReportCache,
) ReportService {
	return &repService{logger: logger, db: db, conv: conv, chc: chc}
}

func (rs *repService) CreateReport(pCtx context.Context, reportInfo dto.CreateReport) error {
	rs.logger.Debug().Str("evt", "call CreateReport")

	headers, rows, err := rs.db.CreateReport(pCtx, reportInfo.ReportParams.Sql)
	if err != nil {
		return errs.ParsePgError(err)
	}

	switch reportInfo.Format {
	case "pdf":
		err = rs.conv.ToPDF(headers, rows, reportInfo.ReportFile)
	case "csv":
		err = rs.conv.ToCSV(headers, rows, reportInfo.ReportFile, reportInfo.ReportParams.CSVSep)
	case "xlsx":
		err = rs.conv.ToXLSX(headers, rows, reportInfo.ReportFile)
	case "chart":
		fallthrough
	case "json":
		err = rs.conv.ToJSON(headers, rows, reportInfo.ReportFile)
	case "docx":
		err = rs.conv.ToDOCX(headers, rows, reportInfo.ReportFile)
	}

	if err != nil {
		return err
	}

	data := webdto.CacheValue(reportInfo.ReportParams)
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if reportInfo.Format != "chart" {
		if err := rs.chc.Set(pCtx, strconv.FormatInt(reportInfo.UserID, 10), bytes); err != nil {
			rs.logger.Error().Err(err).Msg("failed to save last query")
		}
	}

	return nil
}
