package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sq/internal/convert"
	"sq/internal/logger"
	"sq/internal/repository"
	"sq/internal/service"
	"sq/internal/web"
	"sq/internal/web/handlers"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {

	// CONTEXT
	mainCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// ENVIRONMENT AND CONFIGS
	err := godotenv.Load(".env.test")
	if err != nil {
		log.Fatal(fmt.Errorf("can't load a env: %v", err))
	}

	// LOGGER

	logger, err := logger.InitLogger(os.Getenv("LOGS_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	logger.AppLog.Debug().Str("evt", "loggers are successful setted").Msg("")

	repDB, err := repository.NewReportDB(mainCtx, logger.DbLog)
	if err != nil {
		logger.DbLog.Fatal().Err(err)
		log.Fatal(err)
	}

	repConv := convert.NewReportConvert(logger.CnvLog)

	// INTERFACE
	repSrv := service.NewReportService(logger.SrvLog, repDB, repConv)
	repHand := handlers.NewReportHandler(logger.SrvLog, repSrv)
	server := web.CreateServer(mainCtx, repHand)
	logger.AppLog.Debug().Msg("Init interfaces: Service, RSLAPI, Handler")

	errsCh := make(chan error, 2)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		<-mainCtx.Done()
		if err := server.Shutdown(mainCtx); err != nil {
			errsCh <- err
			_ = server.Close()
			logger.AppLog.Error().Err(fmt.Errorf("shutdown: %v", err))
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
