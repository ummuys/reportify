package config

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
)

func parseStr(env string) (string, error) {
	if env == "" {
		return "", fmt.Errorf("empty")
	}

	return env, nil
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

func parseInt(env string, canBeZero bool) (int, error) {
	intEnv, err := strconv.Atoi(env)
	if err != nil {
		return 0, fmt.Errorf("")
	}
	if !canBeZero && intEnv == 0 {
		return 0, fmt.Errorf("")
	}

	return intEnv, nil
}
