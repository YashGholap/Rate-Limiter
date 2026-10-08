package limiter

import (
	"context"
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type Config struct {
	Capacity   float64
	RefillRate float64
}

type TokenBucket struct {
	capacity   float64
	refillRate float64

	buckets map[string]*bucket

	mu    sync.Mutex
	clock Clock
}

type Clock func() time.Time

func New(config Config) *TokenBucket {
	return &TokenBucket{
		capacity:   config.Capacity,
		refillRate: config.RefillRate,
		buckets:    make(map[string]*bucket),
		clock:      time.Now,
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b, exists := tb.buckets[key]

	if !exists {
		b = &bucket{
			tokens:     tb.capacity,
			lastRefill: tb.clock(),
		}
		tb.buckets[key] = b
	}

	now := tb.clock()
	elapsed := now.Sub(b.lastRefill).Seconds()

	tokensToAdd := elapsed * tb.refillRate

	b.tokens += tokensToAdd

	if b.tokens > tb.capacity {
		b.tokens = tb.capacity
	}

	b.lastRefill = now

	if b.tokens < 1 {
		return false, nil
	}

	b.tokens--
	return true, nil
}
