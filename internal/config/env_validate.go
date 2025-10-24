package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

func parseStr(path string) (string, error) {
	env := os.Getenv(path)
	if env == "" {
		return "", fmt.Errorf("%s is empty", path)
	}

	return env, nil
}

func parseLevel(path string) (zerolog.Level, error) {
	env := os.Getenv(path)
	if env == "" {
		return 0, fmt.Errorf("%s is empty", path)
	}

	level, err := zerolog.ParseLevel(env)
	if err != nil {
		return 0, err
	}

	return level, nil
}

func parseInt(path string, canBeZero bool) (int, error) {
	env := os.Getenv(path)
	intEnv, err := strconv.Atoi(env)
	if err != nil {
		return 0, fmt.Errorf("%s is empty", path)
	}
	if !canBeZero && intEnv == 0 {
		return 0, fmt.Errorf("%s is invalid", path)
	}

	return intEnv, nil
}
