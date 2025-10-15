package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type DBConfig struct {
	Addr                  string
	MinConn               int32
	MaxConn               int32
	MaxConnLifetime       int // SECOND
	MaxConnLifetimeJitter int // SECOND
	MaxConnIdleTime       int // SECOND
	HealthCheckPeriod     int // SECOND
}

func ParseRepDBeEnv() (DBConfig, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid env for %s", env))
	}

	addr, err := parseStr(os.Getenv("DB_RP_ADDR"))
	if err != nil {
		add("db_rp_addr")
	}

	minConn, err := parseInt(os.Getenv("DB_RP_MIN_CONN"), true)
	if err != nil {
		add("db_rp_min_conn")
	}

	maxConn, err := parseInt(os.Getenv("DB_RP_MAX_CONN"), false)
	if err != nil {
		add("db_rp_max_conn")
	}

	mclt, err := parseInt(os.Getenv("DB_RP_MAX_CONN_LIFETIME"), false)
	if err != nil {
		add("db_rp_max_conn_lifetime")
	}

	mcltj, err := parseInt(os.Getenv("DB_RP_MAX_CONN_LIFETIME_JITTER"), true)
	if err != nil {
		add("db_rp_max_conn_lifetime_jitter")
	}

	mcit, err := parseInt(os.Getenv("DB_RP_MAX_CONN_IDLE_TIME"), true)
	if err != nil {
		add("db_rp_max_conn_idle_time")
	}

	hcp, err := parseInt(os.Getenv("DB_RP_HEALTH_CHECK_PERIOD"), false)
	if err != nil {
		add("db_rp_max_conn_idle_time")
	}

	if len(sErr) > 0 {
		msg := strings.Join(sErr, ", ")
		return DBConfig{}, errors.New(msg)
	}

	return DBConfig{
		Addr:                  addr,
		MinConn:               int32(minConn),
		MaxConn:               int32(maxConn),
		MaxConnLifetime:       mclt,
		MaxConnLifetimeJitter: mcltj,
		MaxConnIdleTime:       mcit,
		HealthCheckPeriod:     hcp,
	}, nil
}

func ParseUDDBeEnv() (DBConfig, error) {
	var sErr []string

	add := func(env string) {
		sErr = append(sErr, fmt.Sprintf("invalid env for %s", env))
	}

	addr, err := parseStr(os.Getenv("DB_UD_ADDR"))
	if err != nil {
		add("db_ud_addr")
	}

	minConn, err := parseInt(os.Getenv("DB_UD_MIN_CONN"), true)
	if err != nil {
		add("db_ud_min_conn")
	}

	maxConn, err := parseInt(os.Getenv("DB_UD_MAX_CONN"), false)
	if err != nil {
		add("db_ud_max_conn")
	}

	mclt, err := parseInt(os.Getenv("DB_UD_MAX_CONN_LIFETIME"), false)
	if err != nil {
		add("db_ud_max_conn_lifetime")
	}

	mcltj, err := parseInt(os.Getenv("DB_UD_MAX_CONN_LIFETIME_JITTER"), true)
	if err != nil {
		add("db_ud_max_conn_lifetime_jitter")
	}

	mcit, err := parseInt(os.Getenv("DB_UD_MAX_CONN_IDLE_TIME"), true)
	if err != nil {
		add("db_ud_max_conn_idle_time")
	}

	hcp, err := parseInt(os.Getenv("DB_UD_HEALTH_CHECK_PERIOD"), false)
	if err != nil {
		add("db_ud_max_conn_idle_time")
	}

	if len(sErr) > 0 {
		msg := strings.Join(sErr, ", ")
		return DBConfig{}, errors.New(msg)
	}

	return DBConfig{
		Addr:                  addr,
		MinConn:               int32(minConn),
		MaxConn:               int32(maxConn),
		MaxConnLifetime:       mclt,
		MaxConnLifetimeJitter: mcltj,
		MaxConnIdleTime:       mcit,
		HealthCheckPeriod:     hcp,
	}, nil
}
