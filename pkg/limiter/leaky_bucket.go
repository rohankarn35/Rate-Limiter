package limiter

import (
	"context"
	"math"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

type leakyBucketState struct {
	WaterLevel float64   `json:"water_level"`
	LastLeak   time.Time `json:"last_leak"`
}

// LeakyBucketLimiter enforces the leaky bucket algorithm.
type LeakyBucketLimiter struct {
	store     storage.Storage
	capacity  float64
	leakRate  float64
	keyPrefix string
	ttl       time.Duration
	now       func() time.Time
}

// NewLeakyBucketLimiter returns a limiter that leaks requests over time.
func NewLeakyBucketLimiter(store storage.Storage, capacity int, leakRate float64, keyPrefix string) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		store:     store,
		capacity:  float64(capacity),
		leakRate:  leakRate,
		keyPrefix: keyPrefix,
		ttl:       defaultStateTTL,
		now:       time.Now,
	}
}

// Allow enforces the leaky bucket rules per key.
func (lb *LeakyBucketLimiter) Allow(ctx context.Context, key string) (Result, error) {
	stateKey := lb.stateKey(key)
	state := leakyBucketState{
		LastLeak: lb.now(),
	}
	loaded, err := loadState(ctx, lb.store, stateKey, &state)
	if err != nil {
		return Result{}, err
	}
	if !loaded {
		state.LastLeak = lb.now()
	}

	now := lb.now()
	elapsed := now.Sub(state.LastLeak).Seconds()
	if elapsed > 0 {
		leaked := elapsed * lb.leakRate
		state.WaterLevel = math.Max(0, state.WaterLevel-leaked)
		state.LastLeak = now
	}

	result := Result{
		Limit: int(lb.capacity),
	}

	if state.WaterLevel+1 <= lb.capacity {
		state.WaterLevel++
		result.Allowed = true
	} else {
		result.Allowed = false
		result.RetryAfter = time.Duration(math.Ceil((state.WaterLevel+1-lb.capacity)/lb.leakRate)) * time.Second
		result.ResetAfter = result.RetryAfter
	}

	result.Remaining = int(math.Max(0, lb.capacity-state.WaterLevel))

	if err := saveState(ctx, lb.store, stateKey, &state, lb.ttl); err != nil {
		return Result{}, err
	}

	return result, nil
}

func (lb *LeakyBucketLimiter) stateKey(key string) string {
	if lb.keyPrefix == "" {
		return key
	}
	return lb.keyPrefix + ":" + key
}
