package config

import (
	"errors"
	"fmt"
	"strings"
)

type RepCacheConfig struct {
	Addr     string
	Password string
	DB       int
	Exp      int
}

func ParseReportCacheEnv() (RepCacheConfig, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid env for %s", env))
	}

	addr, err := parseStr("CACHE_ADDR")
	if err != nil {
		add("cache_addr")
	}

	pass, err := parseStr("CACHE_PASSWORD")
	if err != nil {
		add("cache_password")
	}

	db, err := parseInt("CACHE_DB", true)
	if err != nil {
		add("cache_db")
	}

	exp, err := parseInt("CACHE_EXPIRE_TIME", true)
	if err != nil {
		add("cache_db")
	}

	if len(sErr) > 0 {
		msg := strings.Join(sErr, ", ")
		return RepCacheConfig{}, errors.New(msg)
	}

	return RepCacheConfig{
		Addr:     addr,
		Password: pass,
		DB:       db,
		Exp:      exp,
	}, nil
}
