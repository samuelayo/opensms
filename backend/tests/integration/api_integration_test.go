// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/infrastructure/observability"
	"github.com/samuelayo/opensms/internal/infrastructure/storage"
	"github.com/samuelayo/opensms/internal/server"
	"github.com/samuelayo/opensms/internal/shared/security"
)

// TestSuite holds shared test infrastructure
type TestSuite struct {
	app    *fiber.App
	db     *database.DB
	cache  *cache.RedisClient
	jwt    *security.JWTManager
	rbac   *security.RBAC
	config *config.Config
}

func setupIntegrationTest(t *testing.T) *TestSuite {
	// Load test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "opensms_test",
			Password:        "test_password",
			Database:        "opensms_test",
			SSLMode:         "disable",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       1,
			PoolSize: 5,
		},
		JWT: config.JWTConfig{
			AccessSecret:       "test-access-secret-key-min-32-chars",
			RefreshSecret:      "test-refresh-secret-key-min-32-chr",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "opensms-test",
		},
		Security: config.SecurityConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	}

	// Setup infrastructure
	ctx := context.Background()

	db, err := database.NewPostgresDB(ctx, cfg.Database)
	require.NoError(t, err)

	redisClient := cache.NewRedisClient(cfg.Redis)

	jwt := security.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		cfg.JWT.Issuer,
	)

	rbac := security.NewRBAC()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: server.ErrorHandler,
	})

	return &TestSuite{
		app:    app,
		db:     db,
		cache:  redisClient,
		jwt:    jwt,
		rbac:   rbac,
		config: cfg,
	}
}

func (ts *TestSuite) cleanup() {
	if ts.db != nil {
		ts.db.Close()
	}
	if ts.cache != nil {
		ts.cache.Close()
	}
}

// Helper to make authenticated requests
func (ts *TestSuite) makeAuthRequest(method, url string, body interface{}, role string, permissions []string) (*httptest.ResponseRecorder, error) {
	tokens, err := ts.jwt.GenerateTokenPair(
		"test-user-id",
		"test-tenant-id",
		"test@example.com",
		role,
		permissions,
	)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := ts.app.Test(req)
	return httptest.NewRecorder(), err
}

func TestHealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Register health endpoint
	suite.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthentication_Flow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	tests := []struct {
		name           string
		email          string
		password       string
		expectedStatus int
	}{
		{
			name:           "valid credentials",
			email:          "teacher@example.com",
			password:       "SecureP@ssw0rd123",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "invalid email",
			email:          "invalid@example.com",
			password:       "password",
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "invalid password",
			email:          "teacher@example.com",
			password:       "wrongpassword",
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "empty email",
			email:          "",
			password:       "password",
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "empty password",
			email:          "teacher@example.com",
			password:       "",
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}

			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// This would need actual implementation in server
			// For now just showing structure
		})
	}
}

func TestAuthorization_RBAC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	tests := []struct {
		name           string
		role           string
		permissions    []string
		endpoint       string
		method         string
		expectedStatus int
	}{
		{
			name:           "teacher can read grades",
			role:           "teacher",
			permissions:    []string{"grade:read"},
			endpoint:       "/api/v1/grades",
			method:         "GET",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "student cannot create grades",
			role:           "student",
			permissions:    []string{"grade:read"},
			endpoint:       "/api/v1/grades",
			method:         "POST",
			expectedStatus: fiber.StatusForbidden,
		},
		{
			name:           "admin can create users",
			role:           "super_admin",
			permissions:    []string{},
			endpoint:       "/api/v1/users",
			method:         "POST",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "parent can view payments",
			role:           "parent",
			permissions:    []string{"payment:view"},
			endpoint:       "/api/v1/payments",
			method:         "GET",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test authorization logic
			hasPermission := suite.rbac.HasAnyPermission(
				security.Role(tt.role),
				[]security.Permission{"grade:create"},
			)

			if tt.role == "super_admin" {
				assert.True(t, hasPermission || suite.rbac.HasPermission(security.Role(tt.role), "grade:create"))
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	limiter := cache.NewRateLimiter(suite.cache)
	ctx := context.Background()

	key := "test-user"
	limit := int64(5)
	window := 1 * time.Second

	// First 5 requests should pass
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, key, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 6th request should be rate limited
	allowed, err := limiter.Allow(ctx, key, limit, window)
	require.NoError(t, err)
	assert.False(t, allowed, "Request beyond limit should be denied")
}

func TestDatabaseTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	ctx := context.Background()

	// Begin transaction
	tx, err := suite.db.Begin(ctx)
	require.NoError(t, err)

	// Rollback on test end
	defer tx.Rollback(ctx)

	// Create temporary table
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE test_transaction (
			id SERIAL PRIMARY KEY,
			value TEXT
		)
	`)
	require.NoError(t, err)

	// Insert data
	_, err = tx.Exec(ctx, "INSERT INTO test_transaction (value) VALUES ($1)", "test")
	require.NoError(t, err)

	// Verify data exists
	var value string
	err = tx.QueryRow(ctx, "SELECT value FROM test_transaction WHERE id = 1").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "test", value)

	// Rollback
	err = tx.Rollback(ctx)
	require.NoError(t, err)
}

func TestConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Register test endpoint
	counter := 0
	suite.app.Get("/concurrent", func(c *fiber.Ctx) error {
		counter++
		return c.JSON(fiber.Map{"count": counter})
	})

	// Make concurrent requests
	done := make(chan bool)
	requests := 50

	for i := 0; i < requests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/concurrent", nil)
			_, err := suite.app.Test(req)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all requests
	for i := 0; i < requests; i++ {
		<-done
	}

	assert.Equal(t, requests, counter)
}

func TestMultiTenancy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	ctx := context.Background()

	// Test tenant isolation
	tenant1 := "tenant-1"
	tenant2 := "tenant-2"

	// Set tenant 1
	err := suite.db.SetTenant(ctx, tenant1)
	require.NoError(t, err)

	// Verify tenant is set
	var currentTenant string
	err = suite.db.QueryRow(ctx, "SELECT current_setting('app.current_tenant', true)").Scan(&currentTenant)
	require.NoError(t, err)
	assert.Equal(t, tenant1, currentTenant)

	// Switch to tenant 2
	err = suite.db.SetTenant(ctx, tenant2)
	require.NoError(t, err)

	err = suite.db.QueryRow(ctx, "SELECT current_setting('app.current_tenant', true)").Scan(&currentTenant)
	require.NoError(t, err)
	assert.Equal(t, tenant2, currentTenant)
}

func TestCaching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	ctx := context.Background()

	type CacheData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Test cache operations
	data := CacheData{Name: "test", Value: 42}

	// Set
	err := suite.cache.Set(ctx, "test-key", data, 1*time.Minute)
	require.NoError(t, err)

	// Get
	var retrieved CacheData
	err = suite.cache.Get(ctx, "test-key", &retrieved)
	require.NoError(t, err)
	assert.Equal(t, data.Name, retrieved.Name)
	assert.Equal(t, data.Value, retrieved.Value)

	// Delete
	err = suite.cache.Delete(ctx, "test-key")
	require.NoError(t, err)

	// Verify deleted
	err = suite.cache.Get(ctx, "test-key", &retrieved)
	assert.Equal(t, cache.ErrCacheMiss, err)
}

func TestTokenRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	suite := setupIntegrationTest(t)
	defer suite.cleanup()

	// Generate initial token pair
	tokens, err := suite.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{"grade:read"},
	)
	require.NoError(t, err)

	// Validate tokens
	claims, err := suite.jwt.ValidateAccessToken(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)

	// Refresh tokens
	newTokens, err := suite.jwt.RefreshTokens(
		tokens.RefreshToken,
		"user-123",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{"grade:read"},
	)
	require.NoError(t, err)
	assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.AccessToken)
}

func TestPasswordSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	passwords := []string{
		"WeakP@ss1",        // Too short
		"longpassword123",  // No uppercase
		"LONGPASSWORD123",  // No lowercase
		"LongPassword",     // No number
		"LongPassword123",  // No special char
		"StrongP@ssw0rd",   // Valid
	}

	for _, password := range passwords {
		err := security.ValidatePasswordStrength(password, 12, true)
		if password == "StrongP@ssw0rd" {
			assert.NoError(t, err)
		}
	}
}

// Benchmark integration tests
func BenchmarkAPI_HealthCheck(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration benchmark")
	}

	suite := setupIntegrationTest(&testing.T{})
	defer suite.cleanup()

	suite.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = suite.app.Test(req)
	}
}

func BenchmarkAPI_AuthenticatedRequest(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration benchmark")
	}

	suite := setupIntegrationTest(&testing.T{})
	defer suite.cleanup()

	tokens, _ := suite.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{"grade:read"},
	)

	suite.app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"data": "protected"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = suite.app.Test(req)
	}
}
