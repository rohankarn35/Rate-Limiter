package tests

import (
	"testing"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
)

func TestTokenBucketAllow(t *testing.T) {
	tb := limiter.NewTokenBucket(2, 1, time.Second)
	if !tb.Allow() {
		t.Error("expected first request to be allowed")
	}
	if !tb.Allow() {
		t.Error("expected second request to be allowed")
	}
	if tb.Allow() {
		t.Error("expected third request to be rate limited")
	}
	tb.Stop()
}
