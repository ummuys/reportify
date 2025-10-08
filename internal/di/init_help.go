package di

import (
	"context"
	"os"
	"sq/internal/cache"
	"sq/internal/config"
	"sq/internal/convert"
	"sq/internal/logger"
	"sq/internal/repository"
	"sq/internal/secure"
	"sq/internal/service"
	"sq/internal/web/handlers"
)

func InitServices(repos Repositorys, tools Tools) Services {
	repSrv := service.NewReportService(tools.Logger.SrvLog, repos.ReportDB, tools.ReportConvert, repos.ReportCache)
	return Services{ReportService: repSrv}
}

func InitTools() (Tools, error) {
	logger, err := logger.InitLogger(os.Getenv("LOGS_PATH"))
	if err != nil {
		return Tools{}, err
	}
	repConv := convert.NewReportConvert(logger.SvcLog)
	return Tools{ReportConvert: repConv, Logger: logger}, nil
}

func InitHandlers(tools Tools, serv Services, sec Secure) Handlers {
	repHand := handlers.NewReportHandler(tools.Logger.SrvLog, serv.ReportService)
	authHand := handlers.NewAuthHandler(tools.Logger.SrvLog, sec.TokenManager)
	return Handlers{ReportHandler: repHand, AuthHandler: authHand}
}

func InitRepositorys(mainCtx context.Context, logger *config.Loggers) (Repositorys, error) {
	repDB, err := repository.NewReportDB(mainCtx, logger.DbLog)
	if err != nil {
		return Repositorys{}, err
	}
	repChc, err := cache.NewReportCache(mainCtx, logger.ChcLog)
	if err != nil {
		return Repositorys{}, err
	}
	return Repositorys{ReportDB: repDB, ReportCache: repChc}, nil
}

func InitSecure() Secure {
	tm := secure.NewTokenManager()
	ph := secure.NewPasswordHasher()
	return Secure{PasswordHasher: ph, TokenManager: tm}
}
