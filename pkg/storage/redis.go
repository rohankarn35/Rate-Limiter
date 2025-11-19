package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implements Storage using Redis.
type RedisStorage struct {
	client *redis.Client
}

// RedisConfig describes the connection settings.
type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

// NewRedisStorage returns a Storage backed by Redis.
func NewRedisStorage(cfg RedisConfig) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisStorage{client: client}
}

// Get fetches a value from Redis.
func (r *RedisStorage) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return val, err
}

// Set writes data with TTL.
func (r *RedisStorage) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key.
func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Close closes the Redis client connection.
func (r *RedisStorage) Close() error {
	return r.client.Close()
}
