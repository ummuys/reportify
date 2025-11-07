package di

import (
	"context"
	"os"

	"github.com/ummuys/reportify/internal/cache"
	"github.com/ummuys/reportify/internal/config"
	"github.com/ummuys/reportify/internal/convert"
	"github.com/ummuys/reportify/internal/logger"
	"github.com/ummuys/reportify/internal/repository"
	"github.com/ummuys/reportify/internal/secure"
	"github.com/ummuys/reportify/internal/service"
	"github.com/ummuys/reportify/internal/web/handlers"
)

func InitServices(repos Repositories, sec Secure, tools Tools) Services {
	repSrv := service.NewReportService(tools.Logger.SvcLog, repos.ReportDB, tools.ReportConvert, repos.ReportCache)
	userSrv := service.NewUserService(tools.Logger.SvcLog, repos.UserDB, sec.PasswordHasher)
	mdSrv := service.NewMetadataService(tools.Logger.SvcLog, repos.MetadataDB, repos.ReportCache)
	admSrv := service.NewAdminService(tools.Logger.SvcLog, repos.UserDB, sec.PasswordHasher)
	return Services{ReportService: repSrv, UserService: userSrv, MetadataService: mdSrv, AdminService: admSrv}
}

func InitTools() (Tools, error) {
	logger, err := logger.InitLogger(os.Getenv("LOGS_PATH"))
	if err != nil {
		return Tools{}, err
	}
	repConv := convert.NewReportConvert(logger.SvcLog, true)
	return Tools{ReportConvert: repConv, Logger: logger}, nil
}

func InitHandlers(tools Tools, serv Services, sec Secure) Handlers {
	repHand := handlers.NewReportHandler(tools.Logger.SrvLog, serv.ReportService)
	authHand := handlers.NewAuthHandler(tools.Logger.SrvLog, sec.TokenManager, serv.UserService)
	mdHand := handlers.NewMetadataHandler(tools.Logger.SrvLog, serv.MetadataService)
	admHand := handlers.NewAdminHandler(tools.Logger.SrvLog, serv.AdminService)
	return Handlers{ReportHandler: repHand, AuthHandler: authHand, MetadataHandler: mdHand, AdminHandler: admHand}
}

func InitRepositories(mainCtx context.Context, logger *config.Loggers) (Repositories, error) {
	repDB, err := repository.NewReportDB(mainCtx, logger.DbLog)
	if err != nil {
		return Repositories{}, err
	}
	repChc, err := cache.NewReportCache(mainCtx, logger.ChcLog)
	if err != nil {
		return Repositories{}, err
	}

	uDB, err := repository.NewUserDB(mainCtx, logger.DbLog)
	if err != nil {
		return Repositories{}, err
	}

	mdDB, err := repository.NewMetadataDB(mainCtx, logger.DbLog)
	if err != nil {
		return Repositories{}, err
	}

	return Repositories{ReportDB: repDB, ReportCache: repChc, UserDB: uDB, MetadataDB: mdDB}, nil
}

func InitSecure() (Secure, error) {
	tm, err := secure.NewTokenManager()
	if err != nil {
		return Secure{}, err
	}
	ph := secure.NewPasswordHasher()
	return Secure{PasswordHasher: ph, TokenManager: tm}, nil
}
