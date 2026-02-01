package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	userID := "user-123"
	tenantID := "tenant-456"
	email := "test@example.com"
	role := "teacher"
	permissions := []string{"grade:read", "grade:create", "attendance:mark"}

	tokens, err := manager.GenerateTokenPair(userID, tenantID, email, role, permissions)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.False(t, tokens.ExpiresAt.IsZero())
	assert.True(t, tokens.ExpiresAt.After(time.Now()))
}

func TestJWTManager_ValidateAccessToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	userID := "user-123"
	tenantID := "tenant-456"
	email := "test@example.com"
	role := "teacher"
	permissions := []string{"grade:read", "grade:create"}

	// Generate token
	tokens, err := manager.GenerateTokenPair(userID, tenantID, email, role, permissions)
	require.NoError(t, err)

	// Validate token
	claims, err := manager.ValidateAccessToken(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, tenantID, claims.TenantID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, permissions, claims.Permissions)
	assert.Equal(t, "opensms-test", claims.Issuer)
}

func TestJWTManager_ValidateAccessToken_Invalid(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "not-a-valid-token"},
		{"malformed JWT", "header.payload.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.ValidateAccessToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestJWTManager_ValidateAccessToken_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager(
		"secret-key-1-min-32-characters-long",
		"refresh-secret-1-min-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	manager2 := NewJWTManager(
		"secret-key-2-min-32-characters-long",
		"refresh-secret-2-min-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	tokens, err := manager1.GenerateTokenPair("user-123", "tenant-456", "test@example.com", "teacher", []string{})
	require.NoError(t, err)

	// Try to validate with different secret
	_, err = manager2.ValidateAccessToken(tokens.AccessToken)
	assert.Error(t, err)
}

func TestJWTManager_ValidateAccessToken_Expired(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		-1*time.Second, // Already expired
		7*24*time.Hour,
		"opensms-test",
	)

	tokens, err := manager.GenerateTokenPair("user-123", "tenant-456", "test@example.com", "teacher", []string{})
	require.NoError(t, err)

	// Wait a bit to ensure expiration
	time.Sleep(100 * time.Millisecond)

	_, err = manager.ValidateAccessToken(tokens.AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTManager_ValidateRefreshToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	tokens, err := manager.GenerateTokenPair("user-123", "tenant-456", "test@example.com", "teacher", []string{})
	require.NoError(t, err)

	claims, err := manager.ValidateRefreshToken(tokens.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.Subject)
	assert.Equal(t, "opensms-test", claims.Issuer)
}

func TestJWTManager_RefreshTokens(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	userID := "user-123"
	tenantID := "tenant-456"
	email := "test@example.com"
	role := "teacher"
	permissions := []string{"grade:read"}

	// Generate initial tokens
	initialTokens, err := manager.GenerateTokenPair(userID, tenantID, email, role, permissions)
	require.NoError(t, err)

	// Refresh tokens
	newTokens, err := manager.RefreshTokens(
		initialTokens.RefreshToken,
		userID,
		tenantID,
		email,
		role,
		permissions,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	assert.NotEqual(t, initialTokens.AccessToken, newTokens.AccessToken)
}

func TestJWTManager_RefreshTokens_UserMismatch(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	tokens, err := manager.GenerateTokenPair("user-123", "tenant-456", "test@example.com", "teacher", []string{})
	require.NoError(t, err)

	// Try to refresh with different user ID
	_, err = manager.RefreshTokens(
		tokens.RefreshToken,
		"different-user",
		"tenant-456",
		"test@example.com",
		"teacher",
		[]string{},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatch")
}

func TestJWTManager_ValidateRefreshToken_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager(
		"access-secret-1-min-32-chars-long-key",
		"refresh-secret-1-min-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	manager2 := NewJWTManager(
		"access-secret-2-min-32-chars-long-key",
		"refresh-secret-2-min-32-chars-long",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	tokens, err := manager1.GenerateTokenPair("user-123", "tenant-456", "test@example.com", "teacher", []string{})
	require.NoError(t, err)

	_, err = manager2.ValidateRefreshToken(tokens.RefreshToken)
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkGenerateTokenPair(b *testing.B) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GenerateTokenPair("user-123", "tenant-456", "test@example.com", "teacher", []string{"grade:read"})
	}
}

func BenchmarkValidateAccessToken(b *testing.B) {
	manager := NewJWTManager(
		"test-access-secret-key-min-32-chars",
		"test-refresh-secret-key-min-32-chars",
		15*time.Minute,
		7*24*time.Hour,
		"opensms-test",
	)

	tokens, _ := manager.GenerateTokenPair("user-123", "tenant-456", "test@example.com", "teacher", []string{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.ValidateAccessToken(tokens.AccessToken)
	}
}
