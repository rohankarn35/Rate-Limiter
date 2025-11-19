package limiter

import (
	"sync"
	"time"
)

// LeakyBucket implements the leaky bucket algorithm as a meter.
// It allows requests if the "water level" in the bucket is below the capacity.
// Water leaks at a constant rate.
type LeakyBucket struct {
	capacity     float64
	remaining    float64 // Current water level (inverse of remaining space, but let's track 'water')
	leakRate     float64 // Leaks per second
	lastLeakTime time.Time
	mutex        sync.Mutex
}

// NewLeakyBucket creates a new LeakyBucket.
// capacity: maximum burst size.
// leakRate: requests per second that leak out of the bucket.
func NewLeakyBucket(capacity int, leakRate float64) *LeakyBucket {
	return &LeakyBucket{
		capacity:     float64(capacity),
		remaining:    0, // Start empty (0 water)
		leakRate:     leakRate,
		lastLeakTime: time.Now(),
	}
}

// Allow checks if a request is allowed.
// It lazily updates the bucket state based on the time elapsed.
func (lb *LeakyBucket) Allow() bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.lastLeakTime).Seconds()

	// Calculate leak
	leak := elapsed * lb.leakRate
	if leak > 0 {
		lb.remaining -= leak
		if lb.remaining < 0 {
			lb.remaining = 0
		}
		lb.lastLeakTime = now
	}

	// Check if we can add a request (water)
	if lb.remaining+1 <= lb.capacity {
		lb.remaining++
		return true
	}

	return false
}

// Stop is a no-op for the lazy implementation but kept for interface compatibility if needed.
func (lb *LeakyBucket) Stop() {
	// No background routines to stop
}
