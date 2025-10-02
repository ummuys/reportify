package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sq/internal/service"

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

func (sh *repHandler) CreateReport(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		format := g.Param("format")

		switch format {
		case "pdf":
			sh.createReportPDF(g.Request.Context(), g)
		default:
			sh.logger.Error().Str("method", g.Request.Method).
				Str("path", g.FullPath()).
				Str("ip", g.ClientIP()).
				Int("status", http.StatusBadRequest).
				Int("res_bytes", g.Writer.Size()).
				Str("user_agent", g.Request.UserAgent()).
				Str("format", format).
				Msg("bad format for report")
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "you need to choose a format"})
		}
	}
}

func (sh *repHandler) GetSchemas(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		data, err := sh.srv.GetSchemas(pCtx)
		if err != nil {
			sh.logger.Error().Str("method", g.Request.Method).
				Str("path", g.FullPath()).
				Str("ip", g.ClientIP()).
				Int("status", http.StatusInternalServerError).
				Int("res_bytes", g.Writer.Size()).
				Str("user_agent", g.Request.UserAgent()).Err(err)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sh.logger.Info().Str("method", g.Request.Method).
			Str("path", g.FullPath()).
			Str("ip", g.ClientIP()).
			Int("status", http.StatusOK).
			Int("res_bytes", g.Writer.Size()).
			Str("user_agent", g.Request.UserAgent()).
			Msg("list of schemas returned")
		g.JSON(http.StatusOK, data)

	}
}

func (sh *repHandler) GetTables(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {

		schema := g.Query("schema")
		if schema == "" {
			sh.logger.Error().Str("method", g.Request.Method).
				Str("path", g.FullPath()).
				Str("ip", g.ClientIP()).
				Int("status", http.StatusBadRequest).
				Int("res_bytes", g.Writer.Size()).
				Str("user_agent", g.Request.UserAgent()).
				Str("schema_name", schema).
				Msg("bad name for schema")
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema name is required"})
			return
		}

		data, err := sh.srv.GetTables(pCtx, schema)
		if err != nil {
			sh.logger.Error().Str("method", g.Request.Method).
				Str("path", g.FullPath()).
				Str("ip", g.ClientIP()).
				Int("status", http.StatusInternalServerError).
				Int("res_bytes", g.Writer.Size()).
				Str("user_agent", g.Request.UserAgent()).Err(err)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sh.logger.Info().Str("method", g.Request.Method).
			Str("path", g.FullPath()).
			Str("ip", g.ClientIP()).
			Int("status", http.StatusOK).
			Int("res_bytes", g.Writer.Size()).
			Str("user_agent", g.Request.UserAgent()).
			Msg("list of tables returned")

		g.JSON(http.StatusOK, data)

	}
}

func (sh *repHandler) GetColumns(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {

		schema := g.Query("schema")
		table := g.Query("table")
		if schema == "" || table == "" {
			sh.logger.Error().Str("method", g.Request.Method).
				Str("path", g.FullPath()).
				Str("ip", g.ClientIP()).
				Int("status", http.StatusBadRequest).
				Int("res_bytes", g.Writer.Size()).
				Str("user_agent", g.Request.UserAgent()).
				Str("schema_name", schema).
				Str("table_name", table).
				Msg("bad names for schema and table")
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema and table names is required"})
			return
		}
		data, err := sh.srv.GetColumns(pCtx, schema, table)
		fmt.Println(data)
		if err != nil {
			sh.logger.Error().Str("method", g.Request.Method).
				Str("path", g.FullPath()).
				Str("ip", g.ClientIP()).
				Int("status", http.StatusInternalServerError).
				Int("res_bytes", g.Writer.Size()).
				Str("user_agent", g.Request.UserAgent()).Err(err)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sh.logger.Info().Str("method", g.Request.Method).
			Str("path", g.FullPath()).
			Str("ip", g.ClientIP()).
			Int("status", http.StatusOK).
			Int("res_bytes", g.Writer.Size()).
			Str("user_agent", g.Request.UserAgent()).
			Msg("list of tables returned")
		g.JSON(http.StatusOK, data)

	}
}
