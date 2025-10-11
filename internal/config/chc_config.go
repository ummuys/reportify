package config

import (
	"fmt"
	"os"
	"strings"
)

type RepCacheConfig struct {
	Addr     string
	Password string
	DB       int
	Exp      int
}

func ParseRepCacheEnv() (RepCacheConfig, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid env for %s", env))
	}

	addr, err := parseStr(os.Getenv("CACHE_ADDR"))
	if err != nil {
		add("cache_addr")
	}

	pass, err := parseStr(os.Getenv("CACHE_PASSWORD"))
	if err != nil {
		add("cache_password")
	}

	db, err := parseInt(os.Getenv("CACHE_DB"), true)
	if err != nil {
		add("cache_db")
	}

	exp, err := parseInt(os.Getenv("CACHE_EXPIRE_TIME"), true)
	if err != nil {
		add("cache_db")
	}

	if len(sErr) > 0 {
		return RepCacheConfig{}, fmt.Errorf(strings.Join(sErr, ", "))
	}

	return RepCacheConfig{
		Addr:     addr,
		Password: pass,
		DB:       db,
		Exp:      exp,
	}, nil
}
