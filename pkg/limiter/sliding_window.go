package limiter

import (
	"context"
	"math"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

type slidingWindowState struct {
	PrevCount       int       `json:"prev_count"`
	CurrCount       int       `json:"curr_count"`
	CurrWindowStart time.Time `json:"curr_window_start"`
}

// SlidingWindowLimiter approximates a moving window using previous window weights.
type SlidingWindowLimiter struct {
	store      storage.Storage
	limit      int
	windowSize time.Duration
	keyPrefix  string
	ttl        time.Duration
	now        func() time.Time
}

// NewSlidingWindowLimiter instantiates a limiter with the sliding window algorithm.
func NewSlidingWindowLimiter(store storage.Storage, limit int, windowSize time.Duration, keyPrefix string) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		store:      store,
		limit:      limit,
		windowSize: windowSize,
		keyPrefix:  keyPrefix,
		ttl:        2 * windowSize,
		now:        time.Now,
	}
}

// Allow applies the sliding window count per key.
func (sw *SlidingWindowLimiter) Allow(ctx context.Context, key string) (Result, error) {
	stateKey := sw.stateKey(key)
	state := slidingWindowState{
		CurrWindowStart: sw.now(),
	}

	loaded, err := loadState(ctx, sw.store, stateKey, &state)
	if err != nil {
		return Result{}, err
	}
	if !loaded {
		state.CurrWindowStart = sw.now()
	}

	now := sw.now()
	if diff := now.Sub(state.CurrWindowStart); diff >= sw.windowSize {
		windowsPassed := int(diff / sw.windowSize)
		if windowsPassed == 1 {
			state.PrevCount = state.CurrCount
		} else {
			state.PrevCount = 0
		}
		state.CurrCount = 0
		state.CurrWindowStart = state.CurrWindowStart.Add(time.Duration(windowsPassed) * sw.windowSize)
		if state.CurrWindowStart.Before(now.Add(-sw.windowSize)) || state.CurrWindowStart.After(now) {
			state.CurrWindowStart = now
		}
	}

	timeIntoWindow := now.Sub(state.CurrWindowStart).Seconds()
	windowSeconds := sw.windowSize.Seconds()
	weight := (windowSeconds - timeIntoWindow) / windowSeconds
	if weight < 0 {
		weight = 0
	}

	estimatedCount := int(math.Round(float64(state.PrevCount)*weight)) + state.CurrCount

	result := Result{
		Limit:     sw.limit,
		Remaining: int(math.Max(0, float64(sw.limit-estimatedCount))),
	}

	if estimatedCount < sw.limit {
		state.CurrCount++
		result.Allowed = true
		result.Remaining = int(math.Max(0, float64(sw.limit-(estimatedCount+1))))
	} else {
		result.Allowed = false
		result.RetryAfter = sw.windowSize - time.Duration(timeIntoWindow)
		if result.RetryAfter < 0 {
			result.RetryAfter = sw.windowSize
		}
		result.ResetAfter = result.RetryAfter
	}

	if err := saveState(ctx, sw.store, stateKey, &state, sw.ttl); err != nil {
		return Result{}, err
	}

	return result, nil
}

func (sw *SlidingWindowLimiter) stateKey(key string) string {
	if sw.keyPrefix == "" {
		return key
	}
	return sw.keyPrefix + ":" + key
}
