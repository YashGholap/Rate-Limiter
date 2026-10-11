package limiter

import (
	"context"
	_ "embed"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

var _ Limiter = (*RedisLimiter)(nil)

//go:embed token_bucket.lua
var tokenBucketScript string

type RedisLimiter struct {
	client *goredis.Client
	config Config
}

func NewRedisLimiter(client *goredis.Client, config Config) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		config: config,
	}
}

func (rl *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := "rate_limit:" + key

	ttlSeconds := 1.0
	if rl.config.RefillRate > 0 {
		ttlSeconds = 2 * rl.config.Capacity / rl.config.RefillRate
		if ttlSeconds < 1 {
			ttlSeconds = 1
		}
	}

	result, err := rl.client.Eval(
		ctx,
		tokenBucketScript,
		[]string{redisKey},
		rl.config.Capacity,
		rl.config.RefillRate,
		ttlSeconds,
	).Result()

	if err != nil {
		return false, fmt.Errorf("rate limiter eval failed: %w", err)
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected Redis result type: %T", result)
	}

	return allowed == 1, nil
}
