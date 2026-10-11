package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/YashGholap/rate-limiter/internal/config"
	"github.com/YashGholap/rate-limiter/internal/httpapi"
	"github.com/YashGholap/rate-limiter/internal/limiter"
	"github.com/YashGholap/rate-limiter/internal/redis"
	redisclient "github.com/YashGholap/rate-limiter/internal/redis"
)

func main() {

	envPath := flag.String("env", ".env", "path to the environment file")
	flag.Parse()

	cfg, err := config.Load(*envPath)
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	client, err := redis.NewClient(redis.Config{
		URL: cfg.RedisURL,
	})
	if err != nil {
		log.Fatalf("Redis configuration error: %v", err)
	}

	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	err = redisclient.Ping(ctx, client)
	cancel()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	rateLimiter := limiter.NewRedisLimiter(
		client,
		limiter.Config{
			Capacity:   cfg.RateLimitCapacity,
			RefillRate: cfg.RateLimitRefill,
		},
	)

	handler := httpapi.NewHandler(rateLimiter)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/resource", handler.Resource)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Rate limiter API listening on %s", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}

}
