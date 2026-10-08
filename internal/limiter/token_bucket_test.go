package limiter

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_AllowsRequestsWithinCapacity(t *testing.T) {
	limiter := New(Config{
		Capacity:   5,
		RefillRate: 1,
	})

	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(context.Background(), "user:123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allowed {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
}

func TestTokenBucket_RejectsRequestsWhenEmpty(t *testing.T) {
	limiter := New(Config{
		Capacity:   5,
		RefillRate: 0,
	})

	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(context.Background(), "user:123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allowed {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}

	allowed, err := limiter.Allow(context.Background(), "user:123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("6th request should have been rejected")
	}
}

func TestTokenBucket_RefillsTokens(t *testing.T) {
	currentTime := time.Now()

	limiter := New(Config{
		Capacity:   1,
		RefillRate: 10,
	})

	limiter.clock = func() time.Time {
		return currentTime
	}

	allowed, err := limiter.Allow(context.Background(), "user:123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !allowed {
		t.Fatal("first request should have been allowed")
	}

	allowed, err = limiter.Allow(context.Background(), "user:123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("second request should have been rejected")
	}

	currentTime = currentTime.Add(100 * time.Millisecond)

	allowed, err = limiter.Allow(context.Background(), "user:123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !allowed {
		t.Fatal("request should have been allowed after refill")
	}
}

func TestTokenBucket_IsolatesKeys(t *testing.T) {
	limiter := New(Config{
		Capacity:   2,
		RefillRate: 0,
	})

	// user:123 gets two tokens.
	for i := 0; i < 2; i++ {
		allowed, err := limiter.Allow(context.Background(), "user:123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allowed {
			t.Fatalf("user:123 request %d should have been allowed", i+1)
		}
	}

	// user:123 is now exhausted.
	allowed, err := limiter.Allow(context.Background(), "user:123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("user:123 request 3 should have been rejected")
	}

	// user:456 has its own independent bucket.
	for i := 0; i < 2; i++ {
		allowed, err := limiter.Allow(context.Background(), "user:456")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !allowed {
			t.Fatalf("user:456 request %d should have been allowed", i+1)
		}
	}

	// user:456 is now exhausted.
	allowed, err = limiter.Allow(context.Background(), "user:456")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("user:456 request 3 should have been rejected")
	}
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	limiter := New(Config{
		Capacity:   10,
		RefillRate: 0,
	})

	const goroutines = 100

	results := make(chan bool, goroutines)
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			allowed, err := limiter.Allow(
				context.Background(),
				"user:123",
			)

			results <- allowed
			errors <- err
		}()
	}

	allowedCount := 0

	for i := 0; i < goroutines; i++ {
		if err := <-errors; err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if <-results {
			allowedCount++
		}
	}

	if allowedCount != 10 {
		t.Fatalf(
			"expected 10 allowed requests, got %d",
			allowedCount,
		)
	}
}
