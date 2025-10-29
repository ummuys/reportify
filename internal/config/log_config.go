package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

//---LOGS---

type LogLevels struct {
	AppLvl zerolog.Level
	SrvLvl zerolog.Level
	DbLvl  zerolog.Level
	ChcLvl zerolog.Level
	SvcLvl zerolog.Level
}

type Loggers struct {
	AppLog *zerolog.Logger // APP
	SrvLog *zerolog.Logger // SERVER
	DbLog  *zerolog.Logger // DATABASE
	SvcLog *zerolog.Logger // SERVICE
	ChcLog *zerolog.Logger // CACHE
}

//---LOGS---

func ParseLogLevels() (*LogLevels, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid level for %s", env))
	}

	appLvl, err := parseLevel("LOG_LEVEL_APP")
	if err != nil {
		add("log_level_app")
	}

	srvLvl, err := parseLevel("LOG_LEVEL_SERVER")
	if err != nil {
		add("log_level_server")
	}

	dbLvl, err := parseLevel("LOG_LEVEL_DATABASE")
	if err != nil {
		add("log_level_database")
	}

	svcLvl, err := parseLevel("LOG_LEVEL_SERVICE")
	if err != nil {
		add("log_level_service")
	}

	chcLvl, err := parseLevel("LOG_LEVEL_CACHE")
	if err != nil {
		add("log_level_cache")
	}

	if len(sErr) > 0 {
		msg := strings.Join(sErr, ", ")
		return nil, errors.New(msg)
	}

	return &LogLevels{
		AppLvl: appLvl,
		SrvLvl: srvLvl,
		DbLvl:  dbLvl,
		SvcLvl: svcLvl,
		ChcLvl: chcLvl,
	}, nil
}
