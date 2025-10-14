package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sq/internal/models"

	"github.com/gin-gonic/gin"
)

func (sh *repHandler) createReportPDF(pCtx context.Context, g *gin.Context) {
	f, err := os.CreateTemp("", "report-*.pdf")
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	defer os.Remove(f.Name())

	sh.logger.Debug().Str("evt", "call CreateReport")
	var req models.CreateReport
	if err := g.ShouldBindJSON(&req); err != nil {
		_ = f.Close()
		sh.logger.Error().Err(fmt.Errorf("bad json: %v", err))
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bad json"})
		return
	}

	u := g.GetInt64("user_id")

	err = sh.srv.CreateReport(pCtx, u, req.Sql, f)
	if err != nil {
		sh.logger.Error().Err(err)
		_ = f.Close()
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if err := f.Close(); err != nil {
		sh.logger.Error().Err(err)
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	st, err := os.Stat(f.Name())
	if err != nil || st.Size() == 0 {
		sh.logger.Error().Msg("empty report")
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "empty report"})
		return
	}

	sh.logger.Info().Int64("file_size", st.Size()).Msg("pdf created and sent")
	g.Header("Content-Type", "application/pdf")
	g.Header("Content-Disposition", "attachment; filename=report.pdf")
	g.Status(200)
	g.Set("msg", "report successful created")
	g.File(f.Name())
}
