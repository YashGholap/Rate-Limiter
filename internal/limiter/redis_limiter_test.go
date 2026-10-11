package limiter

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func newTestRedisClient() *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})
}

func TestRedisLimiter_AllowsRequestsWithinCapacity(t *testing.T) {
	client := newTestRedisClient()
	defer client.Close()

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis is not available: %v", err)
	}

	key := "test:capacity:" + time.Now().Format("20060102150405.000000000")
	defer client.Del(ctx, "rate_limit:"+key)

	limiter := NewRedisLimiter(
		client,
		Config{
			Capacity:   5,
			RefillRate: 0,
		},
	)

	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, key)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allowed {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}

	allowed, err := limiter.Allow(ctx, key)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("6th request should have been rejected")
	}
}

func TestRedisLimiter_IsolatesKeys(t *testing.T) {
	client := newTestRedisClient()
	defer client.Close()

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis is not available: %v", err)
	}

	keyA := "test:isolation:a:" + time.Now().Format("20060102150405.000000000")
	keyB := "test:isolation:b:" + time.Now().Format("20060102150405.000000000")

	defer client.Del(
		ctx,
		"rate_limit:"+keyA,
		"rate_limit:"+keyB,
	)

	limiter := NewRedisLimiter(
		client,
		Config{
			Capacity:   2,
			RefillRate: 0,
		},
	)

	for i := 0; i < 2; i++ {
		allowed, err := limiter.Allow(ctx, keyA)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allowed {
			t.Fatalf("key A request %d should have been allowed", i+1)
		}
	}

	allowed, err := limiter.Allow(ctx, keyA)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("key A should have been rate limited")
	}

	allowed, err = limiter.Allow(ctx, keyB)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !allowed {
		t.Fatal("key B should have its own bucket")
	}

}

func TestRedisLimiter_ExpiresInactiveBucket(t *testing.T) {
	client := newTestRedisClient()
	defer client.Close()

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis is not available: %v", err)
	}

	key := "test:expiry:" + time.Now().Format("20060102150405.000000000")
	redisKey := "rate_limit:" + key
	defer client.Del(ctx, redisKey)

	limiter := NewRedisLimiter(
		client,
		Config{
			Capacity:   1,
			RefillRate: 0,
		},
	)

	// The first request creates the bucket.
	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("first request should have been allowed")
	}

	// Confirm the key exists and has an expiry.
	exists, err := client.Exists(ctx, redisKey).Result()
	if err != nil {
		t.Fatalf("failed to check key existence: %v", err)
	}
	if exists != 1 {
		t.Fatal("expected Redis bucket to exist")
	}

	ttl, err := client.PTTL(ctx, redisKey).Result()
	if err != nil {
		t.Fatalf("failed to get TTL: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("expected a positive TTL, got %v", ttl)
	}

	// Wait for Redis to expire the key, polling rather than
	// assuming it disappears at an exact instant.
	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		exists, err = client.Exists(ctx, redisKey).Result()
		if err != nil {
			t.Fatalf("failed to check key existence: %v", err)
		}

		if exists == 0 {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("bucket key did not expire within 3 seconds")
}
