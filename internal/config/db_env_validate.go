package config

import (
	"fmt"
	"os"
	"strings"
)

func ParseRepDBeEnv() (*RepDBConfig, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid env for %s", env))
	}

	addr, err := parseStr(os.Getenv("DB_ADDR"))
	if err != nil {
		add("db_addr")
	}

	minConn, err := parseInt(os.Getenv("DB_MIN_CONN"), true)
	if err != nil {
		add("db_min_conn")
	}

	maxConn, err := parseInt(os.Getenv("DB_MAX_CONN"), false)
	if err != nil {
		add("db_max_conn")
	}

	mclt, err := parseInt(os.Getenv("DB_MAX_CONN_LIFETIME"), false)
	if err != nil {
		add("db_max_conn_lifetime")
	}

	mcltj, err := parseInt(os.Getenv("DB_MAX_CONN_LIFETIME_JITTER"), true)
	if err != nil {
		add("db_max_conn_lifetime_jitter")
	}

	mcit, err := parseInt(os.Getenv("DB_MAX_CONN_IDLE_TIME"), true)
	if err != nil {
		add("db_max_conn_idle_time")
	}

	hcp, err := parseInt(os.Getenv("DB_HEALTH_CHECK_PERIOD"), false)
	if err != nil {
		add("db_max_conn_idle_time")
	}

	if len(sErr) > 0 {
		return nil, fmt.Errorf(strings.Join(sErr, ", "))
	}

	return &RepDBConfig{
		Addr:                  addr,
		MinConn:               int32(minConn),
		MaxConn:               int32(maxConn),
		MaxConnLifetime:       mclt,
		MaxConnLifetimeJitter: mcltj,
		MaxConnIdleTime:       mcit,
		HealthCheckPeriod:     hcp,
	}, nil
}
