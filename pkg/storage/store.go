package storage

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound signals a missing key in the backing store.
	ErrNotFound = errors.New("storage: value not found")
)

// Storage is the minimal interface the limiter algorithms rely on. Implementations
// are expected to be safe for concurrent use.
type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
