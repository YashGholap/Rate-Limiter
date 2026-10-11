package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisURL          string
	HTTPAddr          string
	RateLimitCapacity float64
	RateLimitRefill   float64
}

func Load(envPath string) (Config, error) {
	if envPath == "" {
		return Config{}, fmt.Errorf("env file path must not be empty.")
	}

	if _, err := os.Stat(envPath); err != nil {
		return Config{}, fmt.Errorf("access env file %q: %w", envPath, err)
	}

	values, err := godotenv.Read(envPath)
	if err != nil {
		return Config{}, fmt.Errorf("read env file %q: %w", envPath, err)
	}

	redisURL := values["REDIS_URL"]
	if redisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}

	httpAddr := values["HTTP_ADDR"]
	if httpAddr == "" {
		httpAddr = ":8082"
	}

	capacity, err := parsePositiveFloat(values, "RATE_LIMIT_CAPACITY", 10)
	if err != nil {
		return Config{}, err
	}

	refillRate, err := parsePositiveFloat(values, "RATE_LIMIT_REFILL_RATE", 2)
	if err != nil {
		return Config{}, err
	}

	return Config{
		RedisURL:          redisURL,
		HTTPAddr:          httpAddr,
		RateLimitCapacity: capacity,
		RateLimitRefill:   refillRate,
	}, nil
}

func parsePositiveFloat(
	values map[string]string,
	key string,
	defaultValue float64,
) (float64, error) {
	raw := values[key]
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return value, nil
}
