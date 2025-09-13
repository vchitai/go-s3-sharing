package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vchitai/go-s3-sharing/internal/domain"
)

// RedisService implements CacheService for Redis
type RedisService struct {
	client *redis.Client
}

// NewRedisService creates a new Redis service
func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{
		client: client,
	}
}

// Set stores a key-value pair in Redis with expiration
func (r *RedisService) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key in Redis: %w", err)
	}
	return nil
}

// Get retrieves a value from Redis by key
func (r *RedisService) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key from Redis: %w", err)
	}
	return val, nil
}

// Delete removes a key from Redis
func (r *RedisService) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}
	return nil
}
