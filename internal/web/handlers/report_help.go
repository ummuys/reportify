package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ummuys/reportify/internal/models"

	"github.com/gin-gonic/gin"
)

func (sh *repHandler) createReportPDF(pCtx context.Context, g *gin.Context) {
	f, err := os.CreateTemp("", "report-*.pdf")
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: err.Error()})
		return
	}
	defer os.Remove(f.Name())

	sh.logger.Debug().Str("evt", "call CreateReport")
	var req models.CreateReport
	if err := g.ShouldBindJSON(&req); err != nil {
		_ = f.Close()
		sh.logger.Error().Err(fmt.Errorf("bad json: %v", err))
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: "bad json"})
		return
	}

	u := g.GetInt64("user_id")

	err = sh.srv.CreateReport(pCtx, u, req.Sql, f)
	if err != nil {
		sh.logger.Error().Err(err)
		_ = f.Close()
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusBadRequest, models.EmptyResponse{Message: err.Error()})
		return
	}

	if err := f.Close(); err != nil {
		sh.logger.Error().Err(err)
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: err.Error()})
		return
	}

	st, err := os.Stat(f.Name())
	if err != nil || st.Size() == 0 {
		sh.logger.Error().Msg("empty report")
		g.Set("msg", err.Error())
		g.AbortWithStatusJSON(http.StatusInternalServerError, models.EmptyResponse{Message: "empty report"})
		return
	}

	sh.logger.Info().Int64("file_size", st.Size()).Msg("pdf created and sent")
	g.Header("Content-Type", "application/pdf")
	g.Status(200)
	g.Set("msg", "report successful created")
	filename := fmt.Sprintf("report-%s.pdf", time.Now().UTC().Format("20060102-150405"))
	g.FileAttachment(f.Name(), filename)
}
