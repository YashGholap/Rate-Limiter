package main

import (
	"context"
	"fmt"
	"log"

	"github.com/YashGholap/rate-limiter/internal/limiter"
	"github.com/YashGholap/rate-limiter/internal/redis"
)

func main() {

	client := redis.NewClient(redis.Config{
		Addr: "localhost:6379",
	})

	defer client.Close()

	ctx := context.Background()

	if err := redis.Ping(ctx, client); err != nil {
		log.Fatal(err)
	}

	rateLimiter := limiter.NewRedisLimiter(
		client,
		limiter.Config{
			Capacity:   5,
			RefillRate: 1,
		},
	)

	fmt.Println("Rate Limiter starting...")
	fmt.Println("Redis connection successful")

	for i := 1; i <= 7; i++ {
		allowed, err := rateLimiter.Allow(ctx, "user:123")

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Request %d: allowed=%v\n", i, allowed)
	}
}
