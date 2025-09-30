package handlers

import (
	"context"
	"fmt"
	"net/http"
	models "sq/internal/models/post"
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
		sh.logger.Debug().Str("evt", "call CreateReport")
		var req models.CreateReport
		if err := g.ShouldBindJSON(&req); err != nil {
			sh.logger.Error().Err(fmt.Errorf("bad json: %v", err))
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad json"})
			return
		}
		sh.logger.Info().Str("script", req.Sql).Msg("catch new script")
		err := sh.srv.CreateReport(context.Background(), req.Sql)
		if err != nil {
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}

		g.JSON(http.StatusOK, gin.H{"msg": "ok"})
	}
}
