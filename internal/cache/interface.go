package cache

import (
	"sq/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type repCache struct {
	logger *zerolog.Logger
	cli    *redis.Client
}

func NewReportCache(logger *zerolog.Logger) (ReportCache, error) {

	config, err := config.ParseRepCacheEnv()
	if err != nil {
		return nil, err
	}

	cli := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &repCache{
		cli:    cli,
		logger: logger,
	}, nil
}

func (rc *repCache) Set(key string, value string) error {
	return nil
}

func (rc *repCache) Get(key string) error {
	return nil
}
