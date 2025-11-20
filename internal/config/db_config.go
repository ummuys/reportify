package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type DBConfig struct {
	Addr                  string
	MinConn               int32
	MaxConn               int32
	MaxConnLifetime       time.Duration
	MaxConnLifetimeJitter time.Duration
	MaxConnIdleTime       time.Duration
	HealthCheckPeriod     time.Duration
}

func parseDBEnv(prefix string) (DBConfig, error) {
	var errs []string
	add := func(err error) {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	addr, err := parseStr(prefix + "_ADDR")
	add(err)

	minConn, err := parseInt(prefix+"_MIN_CONN", true)
	add(err)

	maxConn, err := parseInt(prefix+"_MAX_CONN", false)
	add(err)

	mclt, err := parseInt(prefix+"_MAX_CONN_LIFETIME", false)
	add(err)

	mcltj, err := parseInt(prefix+"_MAX_CONN_LIFETIME_JITTER", true)
	add(err)

	mcit, err := parseInt(prefix+"_MAX_CONN_IDLE_TIME", true)
	add(err)

	hcp, err := parseInt(prefix+"_HEALTH_CHECK_PERIOD", false)
	add(err)

	if len(errs) > 0 {
		return DBConfig{}, errors.New(strings.Join(errs, ", "))
	}

	// #nosec G115
	if maxConn > 0 && minConn > 0 && maxConn < minConn {
		return DBConfig{}, fmt.Errorf("max_conn must be >= min_conn")
	}

	// #nosec G115
	return DBConfig{
		Addr:                  addr,
		MinConn:               int32(minConn),
		MaxConn:               int32(maxConn),
		MaxConnLifetime:       time.Duration(mclt) * time.Second,
		MaxConnLifetimeJitter: time.Duration(mcltj) * time.Second,
		MaxConnIdleTime:       time.Duration(mcit) * time.Second,
		HealthCheckPeriod:     time.Duration(hcp) * time.Second,
	}, nil
}

func ParseReportDBEnv() (DBConfig, error) { return parseDBEnv("DB_RP") }
func ParseUserDBEnv() (DBConfig, error)   { return parseDBEnv("DB_UD") }
