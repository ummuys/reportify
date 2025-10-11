package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"sq/internal/config"
	"sq/internal/di"
	"sq/internal/errs"
	"sq/internal/web"
	"sync"
	"syscall"
)

// // ENVIRONMENT AND CONFIGS -- Не нужно для docker
// err := godotenv.Load(".env.test")
// if err != nil {
// 	log.Fatal(fmt.Errorf("can't load a env: %v", err))
// }

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

	sec := di.InitSecure()
	srv := di.InitServices(repos, sec, tools)
	hand := di.InitHandlers(tools, srv, sec)
	tools.Logger.AppLog.Info().Msg("Init all interfaces: tools, repos, service, secure and handlers")

	// CREATE BASIC USER

	appConf, err := config.ParseAppConfig()
	fmt.Println(appConf)
	if err != nil {
		tools.Logger.AppLog.Fatal().Err(err).Msg("can't load app config")
	}
	pass, _ := sec.PasswordHasher.Hash(appConf.Password)
	err = repos.UserDB.CreateUser(mainCtx, appConf.Username, pass)
	if err != nil && !errors.Is(err, errs.ErrUsernameAlredyExists) {
		tools.Logger.AppLog.Error().Err(err).Msg("can't init basic user")
	} else {
		tools.Logger.AppLog.Info().Msg("basic user successful init")
	}

	server := web.CreateServer(mainCtx, tools, repos, srv, sec, hand)

	errsCh := make(chan error, 2)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		<-mainCtx.Done()
		if err := server.Shutdown(mainCtx); err != nil {
			errsCh <- err
			_ = server.Close()
			tools.Logger.AppLog.Error().Err(fmt.Errorf("server shutdown: %v", err))
		}
	})

	wg.Go(func() {
		if err := web.RunServer(server); err != nil {
			errsCh <- err
			tools.Logger.AppLog.Error().Err(fmt.Errorf("server: %v", err))
		}
	})

	<-mainCtx.Done()
	tools.Logger.AppLog.Info().Msg("turning down the server")
	wg.Wait()
	close(errsCh)
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
