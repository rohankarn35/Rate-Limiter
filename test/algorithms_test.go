package tests

import (
	"context"
	"testing"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

func TestTokenBucketLimiter(t *testing.T) {
	store := storage.NewMemoryStorage()
	tb := limiter.NewTokenBucketLimiter(store, 2, 1, 200*time.Millisecond, "test")
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		res, err := tb.Allow(ctx, "client-1")
		if err != nil {
			t.Fatalf("allow failed: %v", err)
		}
		if !res.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	res, err := tb.Allow(ctx, "client-1")
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if res.Allowed {
		t.Fatal("third request should be blocked")
	}

	time.Sleep(250 * time.Millisecond)
	res, err = tb.Allow(ctx, "client-1")
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if !res.Allowed {
		t.Fatal("token bucket should allow after refill")
	}
}

func TestLeakyBucketLimiter(t *testing.T) {
	store := storage.NewMemoryStorage()
	lb := limiter.NewLeakyBucketLimiter(store, 3, 10, "lb-test")
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		res, err := lb.Allow(ctx, "client")
		if err != nil {
			t.Fatalf("allow failed: %v", err)
		}
		if !res.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	res, err := lb.Allow(ctx, "client")
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if res.Allowed {
		t.Fatal("fourth request should be denied")
	}

	time.Sleep(150 * time.Millisecond)
	res, err = lb.Allow(ctx, "client")
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if !res.Allowed {
		t.Fatal("request should pass after leak")
	}
}

func TestSlidingWindowLimiter(t *testing.T) {
	store := storage.NewMemoryStorage()
	sw := limiter.NewSlidingWindowLimiter(store, 2, 200*time.Millisecond, "sw")
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		res, err := sw.Allow(ctx, "user")
		if err != nil {
			t.Fatalf("allow failed: %v", err)
		}
		if !res.Allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	res, err := sw.Allow(ctx, "user")
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if res.Allowed {
		t.Fatal("third request must be denied")
	}

	time.Sleep(250 * time.Millisecond)
	res, err = sw.Allow(ctx, "user")
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if !res.Allowed {
		t.Fatal("window should reset")
	}
}
