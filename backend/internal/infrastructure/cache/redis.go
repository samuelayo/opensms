package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
)

// Cache interface defines caching operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Close() error
}

// RedisClient wraps redis.Client with additional functionality
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.PoolSize / 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	return &RedisClient{client: client}
}

// Get retrieves a value from cache and unmarshals it
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Set stores a value in cache with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes one or more keys from cache
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	return n > 0, nil
}

// Increment atomically increments a key's value
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache key: %w", err)
	}
	return val, nil
}

// Expire sets an expiration time on a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache expiration: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Health checks Redis health
func (r *RedisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// GetClient returns the underlying Redis client for advanced operations
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Custom errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)

// Rate limiting using Redis
type RateLimiter struct {
	client *RedisClient
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *RedisClient) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow checks if a request is allowed based on rate limit
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	now := time.Now().Unix()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now/int64(window.Seconds()))

	count, err := rl.client.Increment(ctx, windowKey)
	if err != nil {
		return false, err
	}

	// Set expiration on first increment
	if count == 1 {
		if err := rl.client.Expire(ctx, windowKey, window); err != nil {
			return false, err
		}
	}

	return count <= limit, nil
}
