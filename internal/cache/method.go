package cache

import (
	"context"
	"fmt"
	"sq/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type repCache struct {
	logger *zerolog.Logger
	cli    *redis.Client
	expire time.Duration
}

func NewReportCache(pCtx context.Context, logger *zerolog.Logger) (ReportCache, error) {

	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	config, err := config.ParseRepCacheEnv()
	if err != nil {
		return nil, err
	}

	cli := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis didn't pinged: %v", err)
	}

	return &repCache{
		cli:    cli,
		logger: logger,
		expire: time.Duration(config.Exp) * time.Second,
	}, nil
}

func (rc *repCache) SetQuery(pCtx context.Context, key string, value any) error {
	rc.logger.Debug().Str("evt", "call Set").Msg("")

	ctx, cancel := context.WithTimeout(pCtx, time.Second)
	defer cancel()

	err := rc.cli.LPush(ctx, key, value).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *repCache) GetQuerys(pCtx context.Context, key string) ([]string, error) {
	rc.logger.Debug().Str("evt", "call Get").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second)
	defer cancel()

	value, err := rc.cli.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	return value, nil
}
