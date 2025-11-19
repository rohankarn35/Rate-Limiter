package storage

import (
	"context"
	"sync"
	"time"
)

// memoryEntry represents a memoized value with an optional expiration.
type memoryEntry struct {
	value   []byte
	expires time.Time
}

// MemoryStorage is an in-memory implementation that satisfies Storage.
type MemoryStorage struct {
	mu    sync.RWMutex
	store map[string]memoryEntry
}

// NewMemoryStorage creates a new MemoryStorage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: make(map[string]memoryEntry),
	}
}

// Get returns the value for a key if present and not expired.
func (m *MemoryStorage) Get(_ context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	entry, ok := m.store[key]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	if !entry.expires.IsZero() && time.Now().After(entry.expires) {
		m.mu.Lock()
		delete(m.store, key)
		m.mu.Unlock()
		return nil, ErrNotFound
	}
	return entry.value, nil
}

// Set saves the value with the provided TTL.
func (m *MemoryStorage) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	var expires time.Time
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}

	m.mu.Lock()
	m.store[key] = memoryEntry{
		value:   append([]byte(nil), value...),
		expires: expires,
	}
	m.mu.Unlock()
	return nil
}

// Delete removes a key from the storage.
func (m *MemoryStorage) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.store, key)
	m.mu.Unlock()
	return nil
}
