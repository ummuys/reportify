package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sq/internal/cache"
	"sq/internal/convert"
	"sq/internal/logger"
	"sq/internal/repository"
	"sq/internal/secure"
	"sq/internal/service"
	"sq/internal/web"
	ha "sq/internal/web/handlers/auth"
	hr "sq/internal/web/handlers/report"
	"sync"
	"syscall"
)

func main() {

	// CONTEXT
	mainCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// // ENVIRONMENT AND CONFIGS -- Не нужно для docker
	// err := godotenv.Load(".env.test")
	// if err != nil {
	// 	log.Fatal(fmt.Errorf("can't load a env: %v", err))
	// }

	// LOGGER
	logger, err := logger.InitLogger(os.Getenv("LOGS_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	logger.AppLog.Info().Str("msg", "loggers successfully set up").Msg("")

	// INTERFACE
	repDB, err := repository.NewReportDB(mainCtx, logger.DbLog)
	if err != nil {
		logger.DbLog.Fatal().Err(err).Msg("")
		return
	}
	repChc, err := cache.NewReportCache(mainCtx, logger.ChcLog)
	if err != nil {
		logger.DbLog.Fatal().Err(err).Msg("")
		return
	}
	tm := secure.NewTokenManager()
	repConv := convert.NewReportConvert(logger.SvcLog)
	repSrv := service.NewReportService(logger.SvcLog, repDB, repConv, repChc)
	repHand := hr.NewReportHandler(logger.SrvLog, repSrv)
	authHand := ha.NewAuthHandler(logger.SrvLog, tm)
	logger.AppLog.Info().Msg("Init interfaces: repService, repCache, repHandler, authHandler, repDatabase, repConv")

	server := web.CreateServer(mainCtx, repHand, authHand, logger.SrvLog)

	errsCh := make(chan error, 2)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		<-mainCtx.Done()
		if err := server.Shutdown(mainCtx); err != nil {
			errsCh <- err
			_ = server.Close()
			logger.AppLog.Error().Err(fmt.Errorf("server shutdown: %v", err))
		}
	})

	wg.Go(func() {
		if err := web.RunServer(server); err != nil {
			errsCh <- err
			logger.AppLog.Error().Err(fmt.Errorf("server: %v", err))
		}
	})

	<-mainCtx.Done()
	logger.AppLog.Info().Msg("turning down the server")
	wg.Wait()
	close(errsCh)
	var hadErr bool

	for err := range errsCh {
		if err != nil {
			hadErr = true
			logger.AppLog.Error().Err(err).Send()
		}
	}

	if hadErr {
		logger.AppLog.Error().Msg("fatal shutdown")
	} else {
		logger.AppLog.Info().Msg("shutdown successful")
	}
}
