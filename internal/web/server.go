package web

import (
	"context"
	"net"
	"net/http"
	"os"
	ra "sq/internal/web/handlers/auth"
	rh "sq/internal/web/handlers/report"
	"sq/internal/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func CreateServer(pCtx context.Context, repHand rh.ReportHandler,
	authHand ra.AuthHandler, logger *zerolog.Logger) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.Use(middleware.RequestLogger(logger))
	g.Use(gin.Recovery())
	g.Use(cors.Default())
	// g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// REPORT
	g.POST(CreateReportPath, repHand.CreateReport(pCtx))
	g.GET(GetSchemasPath, repHand.GetSchemas(pCtx))
	g.GET(GetTablesPath, repHand.GetTables(pCtx))
	g.GET(GetColumnsPath, repHand.GetColumns(pCtx))
	g.GET(GetHashQuerysPath, repHand.GetHashQuerys(pCtx))

	g.POST(AuthPath, authHand.Authorization(pCtx))

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
