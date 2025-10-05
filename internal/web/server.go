package web

import (
	"context"
	"net"
	"net/http"
	"os"
	"sq/internal/middleware"
	"sq/internal/web/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func CreateServer(pCtx context.Context, repHand handlers.ReportHandler, logger *zerolog.Logger) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.Use(middleware.RequestLogger(logger))
	g.Use(gin.Recovery())
	g.Use(cors.Default())
	// g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	g.POST(CreateReportPath, repHand.CreateReport(pCtx))
	g.GET(GetSchemas, repHand.GetSchemas(pCtx))
	g.GET(GetTables, repHand.GetTables(pCtx))
	g.GET(GetColumns, repHand.GetColumns(pCtx))

	host := os.Getenv("SERVER_IP")
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "1337"
	}

	server := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: g,
	}

	return server
}

func RunServer(server *http.Server) error {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
