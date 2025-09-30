package handlers

import (
	"context"
	"fmt"
	"net/http"
	models "sq/internal/models/response/get"
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
			sh.logger.Error().Msg("bad format")
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "you need to choose a format"})
		}

	}
}

func (sh *repHandler) GetSchemas(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		data, err := sh.srv.GetSchemas(pCtx)
		if err != nil {
			sh.logger.Error().Err(err)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sh.logger.Info().Msg("list of schemas name is returned")
		g.JSON(http.StatusOK, models.ListSchemas{Schemas: data})

	}
}

func (sh *repHandler) GetTables(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {

		schema := g.Query("schema")
		if schema == "" {
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema is required"})
			return
		}

		data, err := sh.srv.GetTables(pCtx, schema)
		if err != nil {
			sh.logger.Error().Err(err)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sh.logger.Info().Msg("list of schemas name is returned")
		g.JSON(http.StatusOK, models.ListTables{Tables: data})

	}
}

func (sh *repHandler) GetColumns(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {

		schema := g.Query("schema")
		table := g.Query("table")
		if schema == "" || table == "" {
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema and table are required"})
			return
		}
		data, err := sh.srv.GetColumns(pCtx, schema, table)
		fmt.Println(data)
		if err != nil {
			sh.logger.Error().Err(err)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sh.logger.Info().Msg("list of columns name is returned")
		g.JSON(http.StatusOK, models.ListColumns{Columns: data})

	}
}
