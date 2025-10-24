package config

import (
	"errors"
	"strings"
	"time"
)

type TMConfig struct {
	AccessTokenLimit  time.Duration
	AccessSecret      string
	RefreshTokenLimit time.Duration
	RefreshSecret     string
}

func ParseTMConfig() (TMConfig, error) {
	var errs []string
	add := func(err error) {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	atl, err := parseInt("ACCESS_TOKEN_LIMIT", false)
	if err != nil {
		add(err)
	}

	ac, err := parseStr("ACCESS_SECRET")
	if err != nil {
		add(err)
	}

	rtl, err := parseInt("REFRESH_TOKEN_LIMIT", false)
	if err != nil {
		add(err)
	}

	rc, err := parseStr("REFRESH_SECRET")
	if err != nil {
		add(err)
	}

	if len(errs) > 0 {
		return TMConfig{}, errors.New(strings.Join(errs, ", "))
	}

	return TMConfig{
		AccessTokenLimit:  time.Duration(atl) * time.Second,
		AccessSecret:      ac,
		RefreshTokenLimit: time.Duration(rtl) * time.Second,
		RefreshSecret:     rc,
	}, nil
}
