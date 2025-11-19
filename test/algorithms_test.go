package tests

import (
	"testing"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
)

func TestLeakyBucket(t *testing.T) {
	// Capacity 10, Leak Rate 10 per second (1 token every 100ms)
	lb := limiter.NewLeakyBucket(10, 10)

	// 1. Fill the bucket
	for i := 0; i < 10; i++ {
		if !lb.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 2. Overflow
	if lb.Allow() {
		t.Error("Request 11 should be denied (bucket full)")
	}

	// 3. Wait for leak (550ms should leak 5.5 -> 5 tokens)
	// Remaining was 10. Leaks 5.5. Becomes 4.5.
	// Can we add 1? 4.5 + 1 = 5.5 <= 10. Yes.
	// Can we add 5?
	// 4.5 + 1 = 5.5
	// ...
	// 8.5 + 1 = 9.5
	// 9.5 + 1 = 10.5 -> No.
	// So we can add 5 requests.
	time.Sleep(550 * time.Millisecond)

	// 4. Should allow 5 requests
	for i := 0; i < 5; i++ {
		if !lb.Allow() {
			t.Errorf("Request %d after leak should be allowed", i+1)
		}
	}

	// 5. Should deny next
	if lb.Allow() {
		t.Error("Request after refill limit should be denied")
	}
}

func TestSlidingWindowCounter(t *testing.T) {
	// Limit 10, Window 1s
	windowSize := 1000 * time.Millisecond
	sw := limiter.NewSlidingWindowCounter(10, windowSize)

	// 1. Fill the window
	for i := 0; i < 10; i++ {
		if !sw.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 2. Exceed limit
	if sw.Allow() {
		t.Error("Request 11 should be denied")
	}

	// 3. Move to next window (partial overlap)
	// Wait 1450ms (1.45 * window).
	// We want to be safely in the "weight ~ 0.55" zone to avoid int truncation flipping 5 to 4.
	// Window size 1000ms.
	// T = 1450ms.
	// Current Window Start = 1000ms.
	// Time into window = 450ms.
	// Weight = (1000 - 450) / 1000 = 0.55.
	// Estimated = 10 * 0.55 = 5.5 -> 5.
	// Allowed = 10 - 5 = 5.
	// If we oversleep up to 1499ms, weight > 0.5, estimated is still 5.
	time.Sleep(1450 * time.Millisecond)

	// 4. Should allow 5 requests
	for i := 0; i < 5; i++ {
		if !sw.Allow() {
			t.Errorf("Request %d (new window) should be allowed", i+1)
		}
	}

	// 5. Should deny next
	if sw.Allow() {
		t.Error("Request (new window overflow) should be denied")
	}
}

func TestSlidingWindowReset(t *testing.T) {
	// Limit 2, Window 100ms
	sw := limiter.NewSlidingWindowCounter(2, 100*time.Millisecond)

	sw.Allow()
	sw.Allow()

	// Wait for 2.5 windows to pass completely
	time.Sleep(250 * time.Millisecond)

	// Should be fully reset
	if !sw.Allow() {
		t.Error("Should allow after full reset")
	}
	if !sw.Allow() {
		t.Error("Should allow second request after full reset")
	}
}
