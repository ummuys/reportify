package handlers

import (
	"context"
	"net/http"
	"sq/internal/models"
	"sq/internal/service"
	"strconv"

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

func (rh *repHandler) GetSchemas(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		rh.logger.Debug().Str("evt", "call GetSchemas")
		data, err := rh.srv.GetSchemas(pCtx)
		if err != nil {
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "schema names are returned")
		g.JSON(http.StatusOK, data)

	}
}

func (rh *repHandler) GetTables(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		rh.logger.Debug().Str("evt", "call GetTables")
		schema := g.Query("schema")
		if schema == "" {
			g.Set("msg", "schema name: "+schema)
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "schema name is required"})
			return
		}

		data, err := rh.srv.GetTables(pCtx, schema)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "table names are returned")
		g.JSON(http.StatusOK, data)

	}
}

func (rh *repHandler) GetColumns(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		rh.logger.Debug().Str("evt", "call GetColumns")
		schema := g.Query("schema")
		table := g.Query("table")
		if schema == "" || table == "" {
			g.Set("msg", "schema: "+schema+"; table: "+table)
			g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "schema and table names are required"})
			return
		}
		data, err := rh.srv.GetColumns(pCtx, schema, table)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "column names are returned")
		g.JSON(http.StatusOK, data)

	}
}

func (rh *repHandler) GetCacheQueries(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		user_id := g.GetInt64("user_id")
		queries, err := rh.srv.GetCacheQueries(pCtx, strconv.FormatInt(user_id, 10))
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "list Queries are returned")
		g.JSON(http.StatusOK, models.QueryList{Queries: queries})
	}
}
