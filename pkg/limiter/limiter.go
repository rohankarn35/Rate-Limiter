package limiter

import "time"

type Result struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
	Limit      int
	ResetAfter time.Duration
}

type Limiter interface {
	Allow(key string) Result
}
