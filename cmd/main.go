package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ummuys/reportify/internal/config"
	"github.com/ummuys/reportify/internal/di"
	"github.com/ummuys/reportify/internal/errs"
	"github.com/ummuys/reportify/internal/web"
)

// @title           github.com/ummuys/reportify
// @version         1.0
// @description     API для отчетов
// @host            localhost:1337
// @BasePath       	/
// @schemes         http

func main() {

	// CONTEXT
	mainCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// INTERFACES
	tools, err := di.InitTools()
	if err != nil {
		log.Fatal(err)
	}
	tools.Logger.AppLog.Info().Msg("Start the app")

	repos, err := di.InitRepositorys(mainCtx, tools.Logger)
	if err != nil {
		tools.Logger.AppLog.Fatal().Err(err).Msg("")
	}

	sec, err := di.InitSecure()
	if err != nil {
		tools.Logger.AppLog.Fatal().Err(err).Msg("")
	}
	srv := di.InitServices(repos, sec, tools)
	hand := di.InitHandlers(tools, srv, sec)
	tools.Logger.AppLog.Info().Msg("Init all interfaces: tools, repos, service, secure and handlers")

	// CREATE BASIC USER
	appConf, err := config.ParseAppConfig()
	if err != nil {
		tools.Logger.AppLog.Fatal().Err(err).Msg("can't load app config")
	}
	pass, err := sec.PasswordHasher.Hash(appConf.Password)
	if err != nil {
		tools.Logger.AppLog.Fatal().Err(err).Msg("can't hash the password")
	} else {
		err = repos.UserDB.CreateUser(mainCtx, appConf.Username, pass)
		if err != nil && !errors.Is(err, errs.ErrUsernameAlredyExists) {
			tools.Logger.AppLog.Error().Err(err).Msg("can't init basic user")
		} else {
			tools.Logger.AppLog.Info().Msg("basic user successfull init")
		}
	}

	//Warn Up (Cache)
	m, err := repos.MetadataDB.GetCacheQueries(mainCtx)
	if err != nil {
		tools.Logger.AppLog.Fatal().Err(err).Msg("can't get cache querys")
	}
	err = repos.ReportCache.Init(mainCtx, m)
	if err != nil {
		tools.Logger.AppLog.Fatal().Err(err).Msg("can't set querys in cache")
	}
	tools.Logger.DbLog.Info().Msg("cache successfully warmed up")

	server := web.CreateServer(mainCtx, tools, repos, srv, sec, hand)
	errsCh := make(chan error, 4)
	srvOff := make(chan struct{})

	// Аккуратно выключаем сервер после ctx.Done
	wg := sync.WaitGroup{}
	wg.Go(func() {
		<-mainCtx.Done()
		defer server.Close()
		if err := server.Shutdown(mainCtx); err != nil {
			errsCh <- err
			tools.Logger.AppLog.Error().Err(fmt.Errorf("server shutdown: %v", err))
		}
		srvOff <- struct{}{}
	})

	// Включаем сервер
	wg.Go(func() {
		if err := web.RunServer(server); err != nil {
			errsCh <- err
			tools.Logger.AppLog.Error().Err(fmt.Errorf("server: %v", err))
		}
	})

	// Сохраняем кэш
	wg.Go(func() {
		<-srvOff
		cache, err := repos.ReportCache.GetAll(context.Background())
		if err != nil {
			errsCh <- err
		}

		err = repos.MetadataDB.SetCacheQueries(context.Background(), cache)
		if err != nil {
			errsCh <- err
		}
	})

	<-mainCtx.Done()
	tools.Logger.AppLog.Info().Msg("turning down the server")
	wg.Wait()
	close(errsCh)
	close(srvOff)
	var hadErr bool

	for err := range errsCh {
		if err != nil {
			hadErr = true
			tools.Logger.AppLog.Error().Err(err).Send()
		}
	}

	if hadErr {
		tools.Logger.AppLog.Error().Msg("fatal shutdown")
	} else {
		tools.Logger.AppLog.Info().Msg("shutdown successful")
	}
}
