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

	result, err := rl.client.Eval(
		ctx,
		tokenBucketScript,
		[]string{redisKey},
		rl.config.Capacity,
		rl.config.RefillRate,
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
