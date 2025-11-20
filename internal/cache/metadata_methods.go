package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ummuys/reportify/internal/config"

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

	config, err := config.ParseReportCacheEnv()
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

func (rc *repCache) Init(pCtx context.Context, queries map[string][]byte) error {
	rc.logger.Debug().Str("evt", "call WarmUp").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, 5*time.Second)
	defer cancel()

	for key, values := range queries {
		for _, val := range values {
			err := rc.cli.LPush(ctx, key, val).Err()
			if err != nil {
				return fmt.Errorf("can't add a query (Init): %v", err)
			}
		}
	}

	return nil
}

func (rc *repCache) Set(pCtx context.Context, key string, value []byte) error {
	rc.logger.Debug().Str("evt", "call Set").Msg("")

	ctx, cancel := context.WithTimeout(pCtx, time.Second)
	defer cancel()

	err := rc.cli.LPush(ctx, key, value).Err()
	if err != nil {
		return fmt.Errorf("can't add a query: %v", err)
	}
	return nil
}

// Make it into [][]byte, because need for api :(
func (rc *repCache) Get(pCtx context.Context, key string) ([][]byte, error) {
	rc.logger.Debug().Str("evt", "call Get").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second)
	defer cancel()

	value, err := rc.cli.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("can't get a value: %v", err)
	}

	bytes := make([][]byte, len(value))
	for i, v := range value {
		bytes[i] = []byte(v)
	}

	return bytes, nil
}

func (rc *repCache) GetAll(pCtx context.Context) (map[string][]byte, error) {
	rc.logger.Debug().Str("evt", "call GetAll").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*5)
	defer cancel()

	keys, err := rc.cli.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	m := make(map[string][]byte)

	for _, k := range keys {
		val, err := rc.cli.LRange(ctx, k, 0, -1).Result()
		if err != nil {
			return nil, fmt.Errorf("can't get a value: %v", err)
		}
		var queries []byte
		queries = append(queries, '[')
		for i, q := range val {
			queries = append(queries, []byte(q)...)
			if i+1 < len(val) {
				queries = append(queries, ',')
			}
		}
		queries = append(queries, ']')
		m[k] = queries
	}

	return m, nil
}

func (rc *repCache) Delete(pCtx context.Context, key string, value []byte) error {
	rc.logger.Debug().Str("evt", "call Delete").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*1)
	defer cancel()

	if _, err := rc.cli.LRem(ctx, key, 0, value).Result(); err != nil {
		return fmt.Errorf("can't delete a value: %v", err)
	}

	return nil
}

func (rc *repCache) DeleteAll(pCtx context.Context, key string) error {
	rc.logger.Debug().Str("evt", "call DeleteAll").Msg("")
	ctx, cancel := context.WithTimeout(pCtx, time.Second*1)
	defer cancel()

	if _, err := rc.cli.Del(ctx, key).Result(); err != nil {
		return fmt.Errorf("can't delete all values: %v", err)
	}

	return nil
}
