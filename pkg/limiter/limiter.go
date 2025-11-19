package limiter

import (
	"context"
	"time"
)

// AlgorithmType enumerates the supported limiter algorithms.
type AlgorithmType string

const (
	AlgorithmTokenBucket   AlgorithmType = "token_bucket"
	AlgorithmLeakyBucket   AlgorithmType = "leaky_bucket"
	AlgorithmSlidingWindow AlgorithmType = "sliding_window"
	defaultStateTTL                      = 5 * time.Minute
)

// Result captures the outcome of a limiter check.
type Result struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
	Limit      int
	ResetAfter time.Duration
}

// Limiter is implemented by algorithm instances that can rate limit based on a key.
type Limiter interface {
	Allow(ctx context.Context, key string) (Result, error)
}
