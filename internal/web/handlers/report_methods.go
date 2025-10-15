package handlers

import (
	"context"
	"net/http"

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
// @Param        request  body   models.CreateReport   true  "SQL-запрос для отчета"
// @Success      200      {file}  file                 "Файл отчета (pdf/csv/docx) или JSON при format=json"
// @Header       200      {string}  Content-Disposition  "attachment; filename=\"report.<ext>\" (для pdf/csv/docx)"
// @Failure      400      {object} models.EmptyResponse "Некорректный запрос или формат"
// @Failure      401      {object} models.EmptyResponse "Неавторизован"
// @Router       /reports/{format} [post]
func (rh *repHandler) CreateReport(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		rh.logger.Debug().Str("evt", "call CreateReport")
		format := g.Param("format")
		switch format {
		case "pdf":
			rh.createReportPDF(g.Request.Context(), g)
		default:
			g.Set("msg", "bad format")
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "format is required"})
		}
	}
}
