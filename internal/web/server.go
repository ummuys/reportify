package web

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/ummuys/reportify/internal/di"
	"github.com/ummuys/reportify/internal/web/middleware"

	_ "github.com/ummuys/reportify/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func CreateServer(pCtx context.Context, tools di.Tools, repos di.Repositorys, srv di.Services, sec di.Secure, hand di.Handlers) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// MAIN
	api := g.Group("/api/v1")
	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:8088"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	api.Use(middleware.RequestLogger(tools.Logger.SrvLog)) // -- Логгирование любого запроса
	api.Use(gin.Recovery())

	// REPORT
	rep := api.Group("")
	rep.Use(middleware.Auth(sec.TokenManager)) // -- Проверка токена каждый раз, когда выполняется запрос
	rep.POST(CreateReportPath, hand.ReportHandler.CreateReport(pCtx))

	// METADATA
	md := api.Group("")
	md.Use(middleware.Auth(sec.TokenManager))
	md.GET(GetSchemasPath, hand.MetadataHandler.GetSchemas(pCtx))
	md.GET(GetTablesPath, hand.MetadataHandler.GetTables(pCtx))
	md.GET(GetColumnsPath, hand.MetadataHandler.GetColumns(pCtx))
	md.GET(GetAllQueriesPath, hand.MetadataHandler.GetQueries(pCtx))
	md.DELETE(DeleteAllQueriesPath, hand.MetadataHandler.DeleteAllQueries(pCtx))

	// SECURE
	auth := api.Group("")
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
