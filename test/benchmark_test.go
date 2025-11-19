package tests

import (
	"context"
	"testing"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

func BenchmarkTokenBucket(b *testing.B) {
	store := storage.NewMemoryStorage()
	tb := limiter.NewTokenBucketLimiter(store, 100, 100, time.Millisecond*10, "bench")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tb.Allow(ctx, "bench-client")
	}
}
