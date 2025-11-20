package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ummuys/reportify/internal/dto"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/service"
	"github.com/ummuys/reportify/internal/validation"
	"github.com/ummuys/reportify/internal/webdto"

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
// @Param        request  body   webdto.RawReportParams   true  "SQL-запрос для отчета"
// @Success      200      {file}  file                 "Файл отчета (pdf/csv/docx) или JSON при format=json"
// @Header       200      {string}  Content-Disposition  "attachment; filename=\"report.<ext>\" (для pdf/csv/docx)"
// @Failure      400      {object} webdto.EmptyResponse "Некорректный запрос или формат"
// @Failure      401      {object} webdto.EmptyResponse "Неавторизован"
// @Router       /reports/{format} [post]
func (rh *repHandler) CreateReport() gin.HandlerFunc {
	return func(g *gin.Context) {
		ctx := g.Request.Context()
		rh.logger.Debug().Str("evt", "call CreateReport")

		format := g.Param("format")
		if !isFormat(format) {
			msg := fmt.Sprintf("bad format: %s", format)
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: msg})
			g.Set("msg", msg)
			return
		}

		var rawParam webdto.RawReportParams
		if err := g.ShouldBindJSON(&rawParam); err != nil {
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			g.Set("msg", err.Error())
			return
		}

		param, err := validation.RequestParams(rawParam, true)
		if err != nil {
			g.AbortWithStatusJSON(http.StatusBadRequest, webdto.EmptyResponse{Message: errs.ErrInvalidJSON.Error()})
			g.Set("msg", err.Error())
			return
		}

		tmpName := fmt.Sprintf("report-*.%s", format)
		f, err := os.CreateTemp("", tmpName)
		if err != nil {
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			g.Set("msg", err.Error())
			return
		}
		defer func() {
			if err := os.Remove(f.Name()); err != nil {
				rh.logger.Err(err).Msg("can't remove tmp file")
			}
		}()

		u := g.GetInt64("user_id")
		err = rh.srv.CreateReport(ctx, dto.CreateReport{
			UserID:       u,
			ReportParams: param,
			ReportFile:   f,
			Format:       format,
		})
		if err != nil {
			_ = f.Close()
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}

		if err := f.Close(); err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}

		st, err := os.Stat(f.Name())
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatusJSON(http.StatusInternalServerError, webdto.EmptyResponse{Message: errs.ErrInternalServer.Error()})
			return
		}

		rh.logger.Info().Int64("file_size", st.Size()).Msg("report created and sent")
		switch format {
		case "pdf":
			g.Header("Content-Type", "application/pdf")
		case "csv":
			g.Header("Content-Type", "text/csv")
		case "xlxs":
			g.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		case "chart":
			fallthrough
		case "json":
			g.Header("Content-Type", "application/json")
		case "docx":
			g.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		}

		g.Status(200)
		g.Set("msg", "report successful created")

		name := "report-%s." + format
		filename := fmt.Sprintf(name, time.Now().UTC().Format("20060102-150405"))
		g.FileAttachment(f.Name(), filename)
	}
}

func isFormat(format string) bool {
	switch format {
	case "pdf":
		fallthrough
	case "csv":
		fallthrough
	case "xlsx":
		fallthrough
	case "json":
		fallthrough
	case "chart":
		fallthrough
	case "docx":
		return true
	}
	return false
}
