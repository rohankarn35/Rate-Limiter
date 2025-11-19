package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

func TestTokenBucketConcurrentAccess(t *testing.T) {
	store := storage.NewMemoryStorage()
	tb := limiter.NewTokenBucketLimiter(store, 5, 5, 100*time.Millisecond, "concurrent")
	ctx := context.Background()

	var wg sync.WaitGroup
	var allowed int
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := tb.Allow(ctx, "shared")
			if err != nil {
				t.Errorf("allow failed: %v", err)
				return
			}
			if res.Allowed {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowed > 5 {
		t.Fatalf("expected at most 5 allowed calls, got %d", allowed)
	}
}
