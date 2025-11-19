package middleware

import (
	"net/http"
	"strconv"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
)

// MetricsRecorder exposes the minimal metric hooks used by the middleware.
type MetricsRecorder interface {
	Observe(policy string, allowed bool)
}

// RateLimiter applies limiter.Manager checks to HTTP traffic.
func RateLimiter(manager *limiter.Manager, recorder MetricsRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if manager == nil {
				next.ServeHTTP(w, r)
				return
			}

			result, policyName, matched, err := manager.Allow(r.Context(), r)
			if err != nil {
				http.Error(w, "rate limiter error", http.StatusInternalServerError)
				return
			}

			if !matched {
				next.ServeHTTP(w, r)
				return
			}

			decorateHeaders(w, result, policyName)
			if recorder != nil {
				recorder.Observe(policyName, result.Allowed)
			}

			if !result.Allowed {
				if result.RetryAfter > 0 {
					w.Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				}
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func decorateHeaders(w http.ResponseWriter, result limiter.Result, policy string) {
	if result.Limit > 0 {
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
	}
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	if result.ResetAfter > 0 {
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(int64(result.ResetAfter.Seconds()), 10))
	}
	if policy != "" {
		w.Header().Set("X-RateLimit-Policy", policy)
	}
}
