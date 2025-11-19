package limiter

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

func loadState[T any](ctx context.Context, store storage.Storage, key string, dst *T) (bool, error) {
	data, err := store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return false, err
	}
	return true, nil
}

func saveState(ctx context.Context, store storage.Storage, key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = defaultStateTTL
	}
	return store.Set(ctx, key, bytes, ttl)
}
