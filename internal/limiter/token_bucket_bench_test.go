package limiter

import (
	"context"
	"testing"
)

func BenchmarkTokenBucket_Allow(b *testing.B) {
	limiter := New(Config{
		Capacity:   1000000,
		RefillRate: 1000000,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.Allow(context.Background(), "user:123")
	}
}
