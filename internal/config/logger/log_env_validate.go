package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func ParseLogLevels() (*LogLevels, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid level for %s", env))
	}

	appLvl, err := parseLevel(os.Getenv("LOG_LEVEL_APP"))
	if err != nil {
		add("app")
	}

	srvLvl, err := parseLevel(os.Getenv("LOG_LEVEL_SERVER"))
	if err != nil {
		add("server")
	}

	dbLvl, err := parseLevel(os.Getenv("LOG_LEVEL_DATABASE"))
	if err != nil {
		add("database")
	}

	if len(sErr) > 0 {
		return nil, fmt.Errorf(strings.Join(sErr, ", "))
	}

	return &LogLevels{
		AppLvl: appLvl,
		SrvLvl: srvLvl,
		DbLvl:  dbLvl,
	}, nil
}

func parseLevel(levelStr string) (zerolog.Level, error) {
	if levelStr == "" {
		return 0, fmt.Errorf("empty")
	}

	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return 0, err
	}

	return level, nil
}
