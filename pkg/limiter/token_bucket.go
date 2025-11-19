package limiter

import (
	"context"
	"math"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

type tokenBucketState struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
}

// TokenBucketLimiter is a distributed-aware token bucket limiter.
type TokenBucketLimiter struct {
	store          storage.Storage
	capacity       float64
	refillRate     float64
	refillInterval time.Duration
	ttl            time.Duration
	keyPrefix      string
	now            func() time.Time
}

// NewTokenBucketLimiter builds a token bucket limiter that persists state in Storage.
func NewTokenBucketLimiter(store storage.Storage, capacity, refillRate int, refillInterval time.Duration, keyPrefix string) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		store:          store,
		capacity:       float64(capacity),
		refillRate:     float64(refillRate),
		refillInterval: refillInterval,
		ttl:            refillInterval * 2,
		keyPrefix:      keyPrefix,
		now:            time.Now,
	}
}

// Allow calculates the bucket state for the provided key.
func (tb *TokenBucketLimiter) Allow(ctx context.Context, key string) (Result, error) {
	stateKey := tb.stateKey(key)
	state := tokenBucketState{
		Tokens:     tb.capacity,
		LastRefill: tb.now(),
	}

	if ok, err := tb.load(ctx, stateKey, &state); err != nil {
		return Result{}, err
	} else if !ok {
		state.Tokens = tb.capacity
		state.LastRefill = tb.now()
	}

	now := tb.now()
	elapsed := now.Sub(state.LastRefill)
	if elapsed > 0 {
		refills := float64(elapsed) / float64(tb.refillInterval)
		if refills > 0 {
			state.Tokens = math.Min(tb.capacity, state.Tokens+refills*tb.refillRate)
			state.LastRefill = now
		}
	}

	result := Result{
		Limit: int(tb.capacity),
	}

	if state.Tokens >= 1 {
		state.Tokens--
		result.Allowed = true
	} else {
		result.Allowed = false
		needed := 1 - state.Tokens
		secondsPerToken := tb.refillInterval.Seconds() / tb.refillRate
		retry := time.Duration(math.Ceil(needed*secondsPerToken)) * time.Second
		if retry < tb.refillInterval {
			result.RetryAfter = retry
		} else {
			result.RetryAfter = tb.refillInterval
		}
		result.ResetAfter = result.RetryAfter
	}

	result.Remaining = int(math.Max(0, state.Tokens))

	if err := saveState(ctx, tb.store, stateKey, &state, tb.ttl); err != nil {
		return Result{}, err
	}

	return result, nil
}

func (tb *TokenBucketLimiter) load(ctx context.Context, key string, dst *tokenBucketState) (bool, error) {
	return loadState(ctx, tb.store, key, dst)
}

func (tb *TokenBucketLimiter) stateKey(key string) string {
	if tb.keyPrefix == "" {
		return key
	}
	return tb.keyPrefix + ":" + key
}
