package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
)

func main() {
	// tb := limiter.NewTokenBucket(10, 5, time.Second)
	// tb := limiter.NewLeakyBucket(10, 5)
	tb := limiter.NewSlidingWindowCounter(10, time.Second)
	var wg sync.WaitGroup

	totalRequests := 20
	wg.Add(totalRequests)

	for i := 1; i <= totalRequests; i++ {
		go func(id int) {
			defer wg.Done()
			if tb.Allow() {
				fmt.Printf("Request %d: ✅ Allowed\n", id)
			} else {
				fmt.Printf("Request %d: ❌ Rate Limited\n", id)
			}
		}(i)
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()

}
