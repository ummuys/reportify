package handlers

import (
	"context"
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
			g.Set("msg", "bad format")
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "format is required"})
		}
	}
}

func (sh *repHandler) GetSchemas(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		data, err := sh.srv.GetSchemas(pCtx)
		if err != nil {
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "schema names are returned")
		g.JSON(http.StatusOK, data)

	}
}

func (sh *repHandler) GetTables(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {

		schema := g.Query("schema")
		if schema == "" {
			g.Set("msg", "schema name: "+schema)
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema name is required"})
			return
		}

		data, err := sh.srv.GetTables(pCtx, schema)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "table names are returned")
		g.JSON(http.StatusOK, data)

	}
}

func (sh *repHandler) GetColumns(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {

		schema := g.Query("schema")
		table := g.Query("table")
		if schema == "" || table == "" {
			g.Set("msg", "schema: "+schema+"; table: "+table)
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema and table names are required"})
			return
		}
		data, err := sh.srv.GetColumns(pCtx, schema, table)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "column names are returned")
		g.JSON(http.StatusOK, data)

	}
}
