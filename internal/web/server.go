package web

import (
	"context"
	"net"
	"net/http"
	"os"
	"sq/internal/di"
	"sq/internal/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CreateServer(pCtx context.Context, tools di.Tools, repos di.Repositorys, srv di.Services, sec di.Secure, hand di.Handlers) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()

	g.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:8088"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	g.Use(middleware.RequestLogger(tools.Logger.SrvLog)) // -- Логгирование любого запроса
	g.Use(gin.Recovery())

	// REPORT
	rep := g.Group("")
	rep.Use(middleware.Auth(sec.TokenManager)) // -- Проверка токена каждый раз, когда выполняется запрос
	rep.POST(CreateReportPath, hand.ReportHandler.CreateReport(pCtx))
	rep.GET(GetSchemasPath, hand.ReportHandler.GetSchemas(pCtx))
	rep.GET(GetTablesPath, hand.ReportHandler.GetTables(pCtx))
	rep.GET(GetColumnsPath, hand.ReportHandler.GetColumns(pCtx))
	rep.GET(GetCacheQuerysPath, hand.ReportHandler.GetCacheQueries(pCtx))

	// SECURE
	auth := g.Group("")
	auth.POST(AuthPath, hand.AuthHandler.Authorization(pCtx))
	auth.GET(GetAccessTokenPath, hand.AuthHandler.UpdateAccessToken(pCtx))

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
