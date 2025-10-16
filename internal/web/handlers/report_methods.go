package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ummuys/reportify/internal/models"
	"github.com/ummuys/reportify/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type repHandler struct {
	logger *zerolog.Logger
	srv    service.ReportService
}

func NewReportHandler(logger *zerolog.Logger, srv service.ReportService) ReportHandler {
	return &repHandler{logger: logger, srv: srv}
}

// CreateReport godoc
// @Summary      Создать отчет
// @Description  Создает пользовательский отчет по SQL-запросу. Формат выбирается параметром {format}. \n\nПоддерживаемые форматы:\n- pdf → application/pdf (файл)\n- csv → text/csv; charset=utf-8 (файл)\n- docx → application/vnd.openxmlformats-officedocument.wordprocessingml.document (файл)\n- json → application/json (объект)\n\nВ ответе для файлов выставляется заголовок Content-Disposition: attachment.
// @Tags         report
// @Accept       json
// @Produce      application/pdf
// @Produce      text/csv
// @Produce      application/json
// @Produce      application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Security     BearerAuth
// @Param        format   path   string                true  "Формат отчета"  Enums(pdf,csv,docx,json)
// @Param        request  body   models.RawReportParams   true  "SQL-запрос для отчета"
// @Success      200      {file}  file                 "Файл отчета (pdf/csv/docx) или JSON при format=json"
// @Header       200      {string}  Content-Disposition  "attachment; filename=\"report.<ext>\" (для pdf/csv/docx)"
// @Failure      400      {object} models.EmptyResponse "Некорректный запрос или формат"
// @Failure      401      {object} models.EmptyResponse "Неавторизован"
// @Router       /reports/{format} [post]
func (rh *repHandler) CreateReport(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		rh.logger.Debug().Str("evt", "call CreateReport")

		format := g.Param("format")
		var rawParam models.RawReportParams
		if err := g.ShouldBindJSON(&rawParam); err != nil {
			rh.logger.Error().Err(fmt.Errorf("bad json: %v", err))
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "bad json"})
			g.Set("msg", err.Error())
			return
		}

		param, err := validParams(rawParam)
		if err != nil {
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "bad json"})
			g.Set("msg", err.Error())
			return
		}

		tmpName := "report-*." + format
		f, err := os.CreateTemp("", tmpName)
		if err != nil {
			g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: err.Error()})
			return
		}

		defer os.Remove(f.Name())
		u := g.GetInt64("user_id")
		err = rh.srv.CreateReport(pCtx, u, param, f, format)
		if err != nil {
			rh.logger.Error().Err(err)
			_ = f.Close()
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: err.Error()})
			return
		}

		if err := f.Close(); err != nil {
			rh.logger.Error().Err(err)
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: err.Error()})
			return
		}

		st, err := os.Stat(f.Name())
		if err != nil || st.Size() == 0 {
			rh.logger.Error().Msg("empty report")
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: "empty report"})
			return
		}

		rh.logger.Info().Int64("file_size", st.Size()).Msg("report created and sent")
		switch format {
		case "pdf":
			g.Header("Content-Type", "application/pdf")
		case "csv":
			g.Header("Content-Type", "text/csv")
		}
		g.Status(200)
		g.Set("msg", "report successful created")

		name := "report-%s." + format
		filename := fmt.Sprintf(name, time.Now().UTC().Format("20060102-150405"))
		g.FileAttachment(f.Name(), filename)

	}
}

func validParams(rawParams models.RawReportParams) (models.ReportParams, error) {
	var (
		params models.ReportParams
		err    error
	)
	if err = checkQuery(rawParams.Sql); err != nil {
		return models.ReportParams{}, err
	}
	params.Sql = rawParams.Sql

	if rawParams.CSVSep != "" {
		params.CSVSep, err = checkSepCSV(rawParams.CSVSep)
		if err != nil {
			return models.ReportParams{}, err
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
