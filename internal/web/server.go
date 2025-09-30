package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sq/internal/web/handlers"

	"github.com/gin-gonic/gin"
)

func CreateServer(pCtx context.Context, slnHand handlers.ReportHandler) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.Use(gin.Recovery())

	// g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	g.POST(CreateReportPath, slnHand.CreateReport(pCtx)) //    C

	host := os.Getenv("SERVER_IP")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "1337"
	}

	server := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: g,
	}

	fmt.Println(server.Addr)

	return server
}

func RunServer(server *http.Server) error {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
