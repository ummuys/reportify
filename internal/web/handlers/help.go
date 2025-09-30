package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	models "sq/internal/models/request/post"

	"github.com/gin-gonic/gin"
)

func (sh *repHandler) createReportPDF(pCtx context.Context, g *gin.Context) {
	// Create a file
	f, err := os.CreateTemp("", "report-*.pdf")
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	defer os.Remove(f.Name())

	// Take a info from request
	sh.logger.Debug().Str("evt", "call CreateReport")
	var req models.CreateReport
	if err := g.ShouldBindJSON(&req); err != nil {
		_ = f.Close()
		sh.logger.Error().Err(fmt.Errorf("bad json: %v", err))
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad json"})
		return
	}

	// Take a data from Postgresql
	sh.logger.Info().Str("script", req.Sql).Msg("catch new script")
	err = sh.srv.CreateReport(context.Background(), req.Sql, f)
	if err != nil {
		sh.logger.Error().Err(err)
		_ = f.Close()
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	// Close the file
	if err := f.Close(); err != nil {
		sh.logger.Error().Err(err)
		g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	// Check the file
	st, err := os.Stat(f.Name())
	if err != nil || st.Size() == 0 {
		sh.logger.Error().Msg("empty report")
		g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "empty report"})
		return
	}

	// If all ok -> send file
	sh.logger.Info().Int64("file_size", st.Size()).Msg("PDF are successful created and sended")
	g.Header("Content-Type", "application/pdf")
	g.Header("Content-Disposition", "attachment; filename=report.pdf")
	g.Status(200)
	g.File(f.Name())
}
