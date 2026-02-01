package server

import (
	"io"
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
	"github.com/samuelayo/opensms/internal/shared/security"
)

func setupTestServer(t *testing.T) (*Server, *fiber.App) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	// Create mock dependencies
	jwt := security.NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	rbac := security.NewRBAC()

	srv := &Server{
		jwt:  jwt,
		rbac: rbac,
		deps: Dependencies{
			Logger: observability.NewLogger(),
		},
	}

	return srv, app
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	srv, app := setupTestServer(t)

	app.Get("/protected", srv.authMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Missing authorization header")
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	srv, app := setupTestServer(t)

	app.Get("/protected", srv.authMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer", "just-a-token"},
		{"wrong prefix", "Basic token123"},
		{"empty token", "Bearer "},
		{"no space", "Bearertoken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.header)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	srv, app := setupTestServer(t)

	app.Get("/protected", srv.authMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Invalid or expired token")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	srv, app := setupTestServer(t)

	// Generate valid token
	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{"grade:read"},
	)
	require.NoError(t, err)

	app.Get("/protected", srv.authMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		tenantID := c.Locals("tenant_id")
		email := c.Locals("email")
		role := c.Locals("role")

		return c.JSON(fiber.Map{
			"user_id":   userID,
			"tenant_id": tenantID,
			"email":     email,
			"role":      role,
		})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "user-123")
	assert.Contains(t, string(body), "tenant-456")
	assert.Contains(t, string(body), "test@example.com")
	assert.Contains(t, string(body), "teacher")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create JWT manager with immediate expiry
	jwt := security.NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		-1*time.Second, // Already expired
		7*24*time.Hour,
		"opensms-test",
	)

	srv := &Server{jwt: jwt}
	app := fiber.New()

	tokens, err := jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{},
	)
	require.NoError(t, err)

	app.Get("/protected", srv.authMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Wait a bit to ensure expiration
	time.Sleep(100 * time.Millisecond)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRequirePermission_HasPermission(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"teacher@example.com",
		"teacher",
		[]string{"grade:read", "grade:create"},
	)
	require.NoError(t, err)

	app.Get("/grades",
		srv.authMiddleware(),
		srv.requirePermission(security.PermGradeRead),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("GET", "/grades", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequirePermission_NoPermission(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"student@example.com",
		"student",
		[]string{"grade:read"}, // Student doesn't have user:delete
	)
	require.NoError(t, err)

	app.Delete("/users/:id",
		srv.authMiddleware(),
		srv.requirePermission(security.PermUserDelete),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("DELETE", "/users/123", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Insufficient permissions")
}

func TestRequirePermission_SuperAdmin(t *testing.T) {
	srv, app := setupTestServer(t)

	// Super admin should have all permissions
	tokens, err := srv.jwt.GenerateTokenPair(
		"admin-123",
		"tenant-456",
		"admin@example.com",
		"super_admin",
		[]string{}, // Empty permissions, but super_admin has all
	)
	require.NoError(t, err)

	app.Delete("/users/:id",
		srv.authMiddleware(),
		srv.requirePermission(security.PermUserDelete),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("DELETE", "/users/123", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireAnyPermission(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"teacher@example.com",
		"teacher",
		[]string{"grade:read"}, // Has one of the required permissions
	)
	require.NoError(t, err)

	app.Get("/grades",
		srv.authMiddleware(),
		srv.requireAnyPermission([]security.Permission{
			security.PermGradeRead,
			security.PermGradeCreate,
		}),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("GET", "/grades", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireAnyPermission_NoMatch(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"student@example.com",
		"student",
		[]string{"grade:read"}, // Doesn't have user permissions
	)
	require.NoError(t, err)

	app.Get("/users",
		srv.authMiddleware(),
		srv.requireAnyPermission([]security.Permission{
			security.PermUserCreate,
			security.PermUserDelete,
		}),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRequireRole(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"admin@example.com",
		"school_admin",
		[]string{},
	)
	require.NoError(t, err)

	app.Get("/admin",
		srv.authMiddleware(),
		srv.requireRole(security.RoleSchoolAdmin),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireRole_WrongRole(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"teacher@example.com",
		"teacher",
		[]string{},
	)
	require.NoError(t, err)

	app.Get("/admin",
		srv.authMiddleware(),
		srv.requireRole(security.RoleSchoolAdmin),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestTenantMiddleware_FromHeader(t *testing.T) {
	srv, app := setupTestServer(t)

	app.Get("/test", srv.tenantMiddleware(), func(c *fiber.Ctx) error {
		tenantID := c.Locals("tenant_id")
		return c.SendString(tenantID.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "tenant-789")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "tenant-789", string(body))
}

func TestTenantMiddleware_MissingTenant(t *testing.T) {
	srv, app := setupTestServer(t)

	app.Get("/test", srv.tenantMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No tenant header, and hostname is likely "example.com" in tests

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestMultipleMiddleware(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"teacher@example.com",
		"teacher",
		[]string{"grade:create"},
	)
	require.NoError(t, err)

	// Chain multiple middleware
	app.Post("/grades",
		srv.authMiddleware(),
		srv.requirePermission(security.PermGradeCreate),
		func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"user_id": c.Locals("user_id"),
				"role":    c.Locals("role"),
			})
		},
	)

	req := httptest.NewRequest("POST", "/grades", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMiddleware_ContextPropagation(t *testing.T) {
	srv, app := setupTestServer(t)

	tokens, err := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{},
	)
	require.NoError(t, err)

	app.Get("/test", srv.authMiddleware(), func(c *fiber.Ctx) error {
		// Verify all expected locals are set
		assert.NotNil(t, c.Locals("user_id"))
		assert.NotNil(t, c.Locals("tenant_id"))
		assert.NotNil(t, c.Locals("email"))
		assert.NotNil(t, c.Locals("role"))
		assert.NotNil(t, c.Locals("permissions"))

		// Verify tenant in context
		ctx := c.UserContext()
		tenantID, ok := database.GetTenant(ctx)
		assert.True(t, ok)
		assert.Equal(t, "tenant-456", tenantID)

		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// Benchmark middleware performance
func BenchmarkAuthMiddleware(b *testing.B) {
	srv, app := setupTestServer(&testing.T{})

	tokens, _ := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{"grade:read"},
	)

	app.Get("/test", srv.authMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = app.Test(req)
	}
}

func BenchmarkRequirePermission(b *testing.B) {
	srv, app := setupTestServer(&testing.T{})

	tokens, _ := srv.jwt.GenerateTokenPair(
		"user-123",
		"tenant-456",
		"teacher@example.com",
		"teacher",
		[]string{"grade:read"},
	)

	app.Get("/test",
		srv.authMiddleware(),
		srv.requirePermission(security.PermGradeRead),
		func(c *fiber.Ctx) error {
			return c.SendString("OK")
		},
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = app.Test(req)
	}
}
