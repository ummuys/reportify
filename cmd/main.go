package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ummuys/reportify/internal/config"
	"github.com/ummuys/reportify/internal/di"
	"github.com/ummuys/reportify/internal/dto"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/web"
)

func main() {
	mainCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// INIT INTERFACES AND TOOLS
	tools, err := di.InitTools()
	if err != nil {
		log.Fatal(err)
	}
	log := tools.Logger.AppLog
	log.Info().Msg("starting Reportify")

	repos, err := di.InitRepositories(mainCtx, tools.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("init repositories failed")
	}

	sec, err := di.InitSecure()
	if err != nil {
		log.Fatal().Err(err).Msg("init secure failed")
	}
	srv := di.InitServices(repos, sec, tools)
	hand := di.InitHandlers(tools, srv, sec)
	log.Info().Msg("initialized all components")

	// INIT CACHE AND DEFAULT USER
	appConf, err := config.ParseAppConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("load app config failed")
	}

	if err := srv.AdminService.CreateUser(mainCtx, dto.CreateUser{
		Username: appConf.Username,
		Password: appConf.Password,
		Role:     "admin",
	}); err == nil || errors.Is(err, errs.ErrDuplicate) {
		log.Info().Msg("default admin user initialized")
	} else {
		log.Error().Err(err).Msg("failed to init admin user")
	}

	cacheQueries, err := repos.MetadataDB.GetCacheQueries(mainCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("load cache queries failed")
	}
	if err := repos.ReportCache.Init(mainCtx, cacheQueries); err != nil {
		log.Fatal().Err(err).Msg("cache warm-up failed")
	}
	tools.Logger.DbLog.Info().Msg("cache warm-up complete")

	// START SERVER
	server := web.CreateServer(tools, repos, srv, sec, hand)
	errsCh := make(chan error, 4)
	srvOff := make(chan struct{})

	var wg sync.WaitGroup

	wg.Go(func() {
		<-mainCtx.Done()
		defer func() { _ = server.Close() }()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			errsCh <- fmt.Errorf("server shutdown: %w", err)
		}
		close(srvOff)
	})

	wg.Go(func() {
		if err := web.RunServer(server); err != nil {
			errsCh <- fmt.Errorf("server start failed: %w", err)
		}
	})

	wg.Go(func() {
		<-srvOff
		cache, err := repos.ReportCache.GetAll(context.Background())
		if err != nil {
			errsCh <- fmt.Errorf("read cache: %w", err)
			return
		}
		if err := repos.MetadataDB.SetCacheQueries(context.Background(), cache); err != nil {
			errsCh <- fmt.Errorf("save cache: %w", err)
		}
	})

	wg.Wait()
	close(errsCh)

	var hadErr bool
	for err := range errsCh {
		if err != nil {
			hadErr = true
			log.Error().Err(err).Send()
		}
	}

	if hadErr {
		log.Error().Msg("graceful shutdown completed with errors")
	} else {
		log.Info().Msg("shutdown successful")
	}
}
