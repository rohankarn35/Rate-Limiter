package limiter

import (
	"sync"
	"time"
)

// SlidingWindowCounter implements the Sliding Window Counter algorithm.
// It approximates the request count in a rolling window using a weighted average
// of the previous window's count and the current window's count.
type SlidingWindowCounter struct {
	limit           int           // Max requests per window
	windowSize      time.Duration // Size of the window
	currWindowStart time.Time     // Start time of the current window
	prevCount       int           // Count in the previous window
	currCount       int           // Count in the current window
	mutex           sync.Mutex
}

// NewSlidingWindowCounter creates a new SlidingWindowCounter.
// limit: maximum requests allowed in the window.
// windowSize: duration of the window (e.g., 1 minute).
func NewSlidingWindowCounter(limit int, windowSize time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		limit:           limit,
		windowSize:      windowSize,
		currWindowStart: time.Now(),
		prevCount:       0,
		currCount:       0,
	}
}

// Allow checks if a request is allowed based on the sliding window approximation.
func (sw *SlidingWindowCounter) Allow() bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	now := time.Now()

	// Check if we need to move the window
	if now.Sub(sw.currWindowStart) >= sw.windowSize {
		// Calculate how many windows have passed
		elapsed := now.Sub(sw.currWindowStart)
		windowsPassed := int64(elapsed / sw.windowSize)

		if windowsPassed == 1 {
			// Just moved to the next window
			sw.prevCount = sw.currCount
			sw.currCount = 0
			sw.currWindowStart = sw.currWindowStart.Add(sw.windowSize)
		} else {
			// Multiple windows passed, reset everything
			sw.prevCount = 0
			sw.currCount = 0
			sw.currWindowStart = now
		}
	}

	// Calculate weighted count
	timeIntoWindow := now.Sub(sw.currWindowStart).Seconds()
	windowSeconds := sw.windowSize.Seconds()
	weight := (windowSeconds - timeIntoWindow) / windowSeconds

	// Ensure weight is not negative (clock skew protection)
	if weight < 0 {
		weight = 0
	}

	estimatedCount := int(float64(sw.prevCount)*weight) + sw.currCount

	if estimatedCount < sw.limit {
		sw.currCount++
		return true
	}

	return false
}
