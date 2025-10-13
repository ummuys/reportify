package handlers

import (
	"context"
	"net/http"
	"sq/internal/models"
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

func (rh *repHandler) CreateReport(pCtx context.Context) gin.HandlerFunc {
	return func(g *gin.Context) {
		rh.logger.Debug().Str("evt", "call CreateReport")
		format := g.Param("format")
		switch format {
		case "pdf":
			rh.createReportPDF(g.Request.Context(), g)
		default:
			g.Set("msg", "bad format")
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "format is required"})
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
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema name is required"})
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
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "schema and table names are required"})
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
		username := g.GetString("username")
		queries, err := rh.srv.GetCacheQueries(pCtx, username)
		if err != nil {
			g.Set("msg", err.Error())
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.Set("msg", "list Queries are returned")
		g.JSON(http.StatusOK, models.QueryList{Queries: queries})
	}
}
