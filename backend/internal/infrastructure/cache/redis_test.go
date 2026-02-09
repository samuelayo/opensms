package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
)

func setupTestRedis(t *testing.T) *RedisClient {
	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       1, // Use DB 1 for testing
		PoolSize: 5,
	}

	client := NewRedisClient(cfg)
	return client
}

func TestNewRedisClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	assert.NotNil(t, client)

	// Test health
	err := client.Health(context.Background())
	assert.NoError(t, err)
}

func TestRedisClient_SetGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Test Set and Get
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := TestData{Name: "test", Value: 42}
	err := client.Set(ctx, "test-key", data, 1*time.Minute)
	require.NoError(t, err)

	var retrieved TestData
	err = client.Get(ctx, "test-key", &retrieved)
	require.NoError(t, err)
	assert.Equal(t, data.Name, retrieved.Name)
	assert.Equal(t, data.Value, retrieved.Value)
}

func TestRedisClient_GetNonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	var result string
	err := client.Get(ctx, "non-existent-key", &result)
	assert.Error(t, err)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestRedisClient_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Set a value
	err := client.Set(ctx, "delete-test", "value", 1*time.Minute)
	require.NoError(t, err)

	// Delete it
	err = client.Delete(ctx, "delete-test")
	require.NoError(t, err)

	// Verify it's gone
	var result string
	err = client.Get(ctx, "delete-test", &result)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestRedisClient_DeleteMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Set multiple values
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := client.Set(ctx, key, "value", 1*time.Minute)
		require.NoError(t, err)
	}

	// Delete all
	err := client.Delete(ctx, keys...)
	require.NoError(t, err)

	// Verify all are gone
	for _, key := range keys {
		var result string
		err := client.Get(ctx, key, &result)
		assert.Equal(t, ErrCacheMiss, err)
	}
}

func TestRedisClient_Exists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Key doesn't exist
	exists, err := client.Exists(ctx, "test-exists")
	require.NoError(t, err)
	assert.False(t, exists)

	// Set key
	err = client.Set(ctx, "test-exists", "value", 1*time.Minute)
	require.NoError(t, err)

	// Now it exists
	exists, err = client.Exists(ctx, "test-exists")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisClient_Increment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test-counter"

	// First increment
	val, err := client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Second increment
	val, err = client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(2), val)

	// Multiple increments
	for i := 0; i < 10; i++ {
		client.Increment(ctx, key)
	}

	val, err = client.Increment(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(13), val)
}

func TestRedisClient_Expire(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Set key without expiration
	redisClient := client.GetClient()
	err := redisClient.Set(ctx, "expire-test", "value", 0).Err()
	require.NoError(t, err)

	// Set expiration
	err = client.Expire(ctx, "expire-test", 100*time.Millisecond)
	require.NoError(t, err)

	// Key should exist
	exists, err := client.Exists(ctx, "expire-test")
	require.NoError(t, err)
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Key should be gone
	exists, err = client.Exists(ctx, "expire-test")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisClient_SetWithExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Set with short expiration
	err := client.Set(ctx, "ttl-test", "value", 100*time.Millisecond)
	require.NoError(t, err)

	// Should exist
	var result string
	err = client.Get(ctx, "ttl-test", &result)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be gone
	err = client.Get(ctx, "ttl-test", &result)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestRedisClient_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	done := make(chan bool)
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		go func(idx int) {
			key := "concurrent-test"
			err := client.Increment(ctx, key)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < iterations; i++ {
		<-done
	}

	// Verify final count
	redisClient := client.GetClient()
	val, err := redisClient.Get(ctx, "concurrent-test").Int64()
	require.NoError(t, err)
	assert.Equal(t, int64(iterations), val)
}

func TestRedisClient_ComplexDataStructures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Test with nested structures
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		ZipCode string `json:"zip_code"`
	}

	type User struct {
		ID        string            `json:"id"`
		Name      string            `json:"name"`
		Email     string            `json:"email"`
		Age       int               `json:"age"`
		Active    bool              `json:"active"`
		Tags      []string          `json:"tags"`
		Metadata  map[string]string `json:"metadata"`
		Address   Address           `json:"address"`
		CreatedAt time.Time         `json:"created_at"`
	}

	user := User{
		ID:     "123",
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    30,
		Active: true,
		Tags:   []string{"admin", "premium"},
		Metadata: map[string]string{
			"role":       "teacher",
			"department": "math",
		},
		Address: Address{
			Street:  "123 Main St",
			City:    "Boston",
			ZipCode: "02101",
		},
		CreatedAt: time.Now(),
	}

	// Store complex object
	err := client.Set(ctx, "user:123", user, 5*time.Minute)
	require.NoError(t, err)

	// Retrieve and verify
	var retrieved User
	err = client.Get(ctx, "user:123", &retrieved)
	require.NoError(t, err)

	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Name, retrieved.Name)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.Age, retrieved.Age)
	assert.Equal(t, user.Active, retrieved.Active)
	assert.Equal(t, user.Tags, retrieved.Tags)
	assert.Equal(t, user.Metadata, retrieved.Metadata)
	assert.Equal(t, user.Address.City, retrieved.Address.City)
}

func TestRateLimiter_Allow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	limiter := NewRateLimiter(client)
	ctx := context.Background()

	key := "test-user"
	limit := int64(5)
	window := 1 * time.Second

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, key, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	allowed, err := limiter.Allow(ctx, key, limit, window)
	require.NoError(t, err)
	assert.False(t, allowed, "Request beyond limit should be denied")

	// Wait for window to reset
	time.Sleep(window + 100*time.Millisecond)

	// Should be allowed again
	allowed, err = limiter.Allow(ctx, key, limit, window)
	require.NoError(t, err)
	assert.True(t, allowed, "Request after window reset should be allowed")
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := setupTestRedis(t)
	defer client.Close()

	limiter := NewRateLimiter(client)
	ctx := context.Background()

	limit := int64(3)
	window := 1 * time.Second

	// User 1 uses their limit
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, "user1", limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// User 1 is rate limited
	allowed, err := limiter.Allow(ctx, "user1", limit, window)
	require.NoError(t, err)
	assert.False(t, allowed)

	// User 2 should still be able to make requests
	allowed, err = limiter.Allow(ctx, "user2", limit, window)
	require.NoError(t, err)
	assert.True(t, allowed, "Different user should have separate limit")
}

// Benchmark tests
func BenchmarkRedisClient_Set(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	client := setupTestRedis(&testing.T{})
	defer client.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = client.Set(ctx, "bench-key", "value", 1*time.Minute)
	}
}

func BenchmarkRedisClient_Get(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	client := setupTestRedis(&testing.T{})
	defer client.Close()

	ctx := context.Background()
	_ = client.Set(ctx, "bench-key", "value", 1*time.Minute)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		_ = client.Get(ctx, "bench-key", &result)
	}
}

func BenchmarkRedisClient_Increment(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	client := setupTestRedis(&testing.T{})
	defer client.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = client.Increment(ctx, "bench-counter")
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	client := setupTestRedis(&testing.T{})
	defer client.Close()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = limiter.Allow(ctx, "bench-user", 1000, 1*time.Second)
	}
}
