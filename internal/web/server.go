package web

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ummuys/reportify/internal/di"
	"github.com/ummuys/reportify/internal/web/middleware"

	// _ "github.com/ummuys/reportify/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func CreateServer(tools di.Tools, repos di.Repositories, srv di.Services, sec di.Secure, hand di.Handlers) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// MAIN
	api := g.Group("/api/v1")
	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:8088"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "UPDATE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	api.Use(middleware.RequestLogger(tools.Logger.SrvLog))
	api.Use(gin.Recovery())

	adminAccess := []string{"admin"}
	basicAccess := []string{"user", "admin"}

	// ADMIN
	adm := api.Group("")
	adm.Use(middleware.Auth(sec.TokenManager, adminAccess))
	adm.GET(GetUsersPath, hand.AdminHandler.GetUsers())
	adm.POST(CreateUserPath, hand.AdminHandler.CreateUser())
	adm.DELETE(DeleteUserPath, hand.AdminHandler.DeleteUser())
	adm.PATCH(UpdateUserPath, hand.AdminHandler.UpdateUser())

	// REPORT
	rep := api.Group("")
	rep.Use(middleware.Auth(sec.TokenManager, basicAccess))
	rep.POST(CreateReportPath, hand.ReportHandler.CreateReport())

	// METADATA
	md := api.Group("")
	md.Use(middleware.Auth(sec.TokenManager, basicAccess))
	md.GET(GetSchemasPath, hand.MetadataHandler.GetSchemas())
	md.GET(GetTablesPath, hand.MetadataHandler.GetTables())
	md.GET(GetColumnsPath, hand.MetadataHandler.GetColumns())
	md.GET(GetAllQueriesPath, hand.MetadataHandler.GetQueries())
	md.DELETE(DeleteAllQueriesPath, hand.MetadataHandler.DeleteAllQueries())
	md.DELETE(DeleteQueryPath, hand.MetadataHandler.DeleteQuery())

	// SECURE
	auth := api.Group("")
	auth.POST(AuthPath, hand.AuthHandler.Authorization())
	auth.GET(GetAccessTokenPath, hand.AuthHandler.UpdateAccessToken())

	host := os.Getenv("SERVER_IP")
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8008"
	}

	server := &http.Server{
		Addr:              net.JoinHostPort(host, port),
		Handler:           g,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server
}

func RunServer(server *http.Server) error {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
