package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type AppConfig struct {
	Username string
	Password string
}

func ParseAppConfig() (AppConfig, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid env for %s", env))
	}

	un, err := parseStr(os.Getenv("BASE_USERNAME"))
	if err != nil {
		add("base_username")
	}

	pw, err := parseStr(os.Getenv("BASE_PASSWORD"))
	if err != nil {
		add("base_password")
	}

	if len(sErr) > 0 {
		msg := strings.Join(sErr, ", ")
		return AppConfig{}, errors.New(msg)
	}

	return AppConfig{Username: un, Password: pw}, nil
}
