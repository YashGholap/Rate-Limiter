package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type Config struct {
	Addr string
}

func NewClient(config Config) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr: config.Addr,
	})
}

func Ping(ctx context.Context, client *goredis.Client) error {
	return client.Ping(ctx).Err()
}
