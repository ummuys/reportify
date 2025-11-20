package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RequestLogger(logger *zerolog.Logger) gin.HandlerFunc {
	return func(g *gin.Context) {
		// Latenсy нормально не отображается
		start := time.Now()
		g.Next()
		status := g.Writer.Status()
		latency := time.Since(start).Milliseconds()

		msg := g.GetString("msg")

		if msg == "" {
			msg = "lost message"
		}

		route := g.FullPath()
		if route == "" {
			route = "-"
		}

		resBytes := g.Writer.Size()
		if resBytes < 0 {
			resBytes = 0
		}

		evt := logger.With().
			Str("method", g.Request.Method).
			Str("route", route).
			Str("url_path", g.Request.URL.Path).
			Str("query", g.Request.URL.RawQuery).
			Str("ip", g.ClientIP()).
			Str("user_agent", g.Request.UserAgent()).
			Int("status", status).
			Int("res_bytes", resBytes).
			Int64("latency", latency).
			Logger()

		if len(g.Errors) > 0 {
			evt.Error().Msg("gin: " + g.Errors.String() + "; server: " + msg)
			return
		}

		switch {
		case status >= http.StatusInternalServerError:
			evt.Error().Msg(msg)
		case status >= http.StatusBadRequest:
			evt.Warn().Msg(msg)
		default:
			evt.Info().Msg(msg)
		}
	}
}
