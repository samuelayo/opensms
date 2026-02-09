package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	// Set required fields
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	// Verify defaults
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
}

func TestLoad_CustomValues(t *testing.T) {
	os.Clearenv()

	// Set custom values
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "custom_user")
	os.Setenv("DB_PASSWORD", "custom_pass")
	os.Setenv("DB_NAME", "custom_db")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("JWT_ACCESS_SECRET", "custom-access-secret-key-32chars")
	os.Setenv("JWT_REFRESH_SECRET", "custom-refresh-secret-key-32char")
	os.Setenv("ENCRYPTION_KEY", "abcdefghijklmnopqrstuvwxyz123456")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 3000, cfg.Server.Port)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "custom_user", cfg.Database.User)
	assert.Equal(t, "custom_pass", cfg.Database.Password)
	assert.Equal(t, "custom_db", cfg.Database.Database)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, 50, cfg.Database.MaxOpenConns)
}

func TestLoad_RedisConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("REDIS_HOST", "redis.example.com")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_PASSWORD", "redis_pass")
	os.Setenv("REDIS_DB", "2")
	os.Setenv("REDIS_POOL_SIZE", "20")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, "redis_pass", cfg.Redis.Password)
	assert.Equal(t, 2, cfg.Redis.DB)
	assert.Equal(t, 20, cfg.Redis.PoolSize)
}

func TestLoad_JWTConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("JWT_ACCESS_SECRET", "my-super-secret-access-key-32char")
	os.Setenv("JWT_REFRESH_SECRET", "my-super-secret-refresh-key-32char")
	os.Setenv("JWT_ACCESS_EXPIRY", "30m")
	os.Setenv("JWT_REFRESH_EXPIRY", "14d")
	os.Setenv("JWT_ISSUER", "custom-issuer")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "my-super-secret-access-key-32char", cfg.JWT.AccessSecret)
	assert.Equal(t, "my-super-secret-refresh-key-32char", cfg.JWT.RefreshSecret)
	assert.Equal(t, 30*time.Minute, cfg.JWT.AccessTokenExpiry)
	assert.Equal(t, 14*24*time.Hour, cfg.JWT.RefreshTokenExpiry)
	assert.Equal(t, "custom-issuer", cfg.JWT.Issuer)
}

func TestLoad_SecurityConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
	os.Setenv("RATE_LIMIT_PER_MINUTE", "200")
	os.Setenv("MAX_LOGIN_ATTEMPTS", "3")
	os.Setenv("LOGIN_ATTEMPT_WINDOW", "10m")
	os.Setenv("PASSWORD_MIN_LENGTH", "8")
	os.Setenv("REQUIRE_STRONG_PASSWORDS", "false")
	os.Setenv("SESSION_TIMEOUT", "60m")
	os.Setenv("CSRF_ENABLED", "false")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "12345678901234567890123456789012", cfg.Security.EncryptionKey)
	assert.Equal(t, 200, cfg.Security.RateLimitPerMinute)
	assert.Equal(t, 3, cfg.Security.MaxLoginAttempts)
	assert.Equal(t, 10*time.Minute, cfg.Security.LoginAttemptWindow)
	assert.Equal(t, 8, cfg.Security.PasswordMinLength)
	assert.False(t, cfg.Security.RequireStrongPasswords)
	assert.Equal(t, 60*time.Minute, cfg.Security.SessionTimeout)
	assert.False(t, cfg.Security.CSRFEnabled)
}

func TestLoad_ObservabilityConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces")
	os.Setenv("PROMETHEUS_PORT", "9091")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENABLE_TRACING", "false")
	os.Setenv("ENABLE_METRICS", "false")
	os.Setenv("TRACING_SAMPLING_RATE", "0.5")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "http://jaeger:14268/api/traces", cfg.Observability.JaegerEndpoint)
	assert.Equal(t, 9091, cfg.Observability.PrometheusPort)
	assert.Equal(t, "debug", cfg.Observability.LogLevel)
	assert.False(t, cfg.Observability.EnableTracing)
	assert.False(t, cfg.Observability.EnableMetrics)
	assert.Equal(t, 0.5, cfg.Observability.SamplingRate)
}

func TestValidate_MissingRequired(t *testing.T) {
	tests := []struct {
		name   string
		setup  func()
		errMsg string
	}{
		{
			name: "missing DB_HOST",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_ACCESS_SECRET", "test-secret-key")
				os.Setenv("JWT_REFRESH_SECRET", "test-refresh-key")
				os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
			},
			errMsg: "DB_HOST is required",
		},
		{
			name: "missing DB_USER",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_ACCESS_SECRET", "test-secret-key")
				os.Setenv("JWT_REFRESH_SECRET", "test-refresh-key")
				os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
			},
			errMsg: "DB_USER is required",
		},
		{
			name: "missing DB_NAME",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("JWT_ACCESS_SECRET", "test-secret-key")
				os.Setenv("JWT_REFRESH_SECRET", "test-refresh-key")
				os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
			},
			errMsg: "DB_NAME is required",
		},
		{
			name: "missing JWT_ACCESS_SECRET",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_REFRESH_SECRET", "test-refresh-key")
				os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
			},
			errMsg: "JWT_ACCESS_SECRET is required",
		},
		{
			name: "missing JWT_REFRESH_SECRET",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_ACCESS_SECRET", "test-secret-key")
				os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
			},
			errMsg: "JWT_REFRESH_SECRET is required",
		},
		{
			name: "missing ENCRYPTION_KEY",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_ACCESS_SECRET", "test-secret-key")
				os.Setenv("JWT_REFRESH_SECRET", "test-refresh-key")
			},
			errMsg: "ENCRYPTION_KEY is required",
		},
		{
			name: "invalid ENCRYPTION_KEY length",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_USER", "test")
				os.Setenv("DB_NAME", "test")
				os.Setenv("JWT_ACCESS_SECRET", "test-secret-key")
				os.Setenv("JWT_REFRESH_SECRET", "test-refresh-key")
				os.Setenv("ENCRYPTION_KEY", "short")
			},
			errMsg: "must be exactly 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			_, err := Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestLoad_MinIOConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("MINIO_ENDPOINT", "minio.example.com:9000")
	os.Setenv("MINIO_ACCESS_KEY", "minioaccess")
	os.Setenv("MINIO_SECRET_KEY", "miniosecret")
	os.Setenv("MINIO_USE_SSL", "true")
	os.Setenv("MINIO_BUCKET", "custom-bucket")
	os.Setenv("MINIO_REGION", "eu-west-1")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "minio.example.com:9000", cfg.MinIO.Endpoint)
	assert.Equal(t, "minioaccess", cfg.MinIO.AccessKey)
	assert.Equal(t, "miniosecret", cfg.MinIO.SecretKey)
	assert.True(t, cfg.MinIO.UseSSL)
	assert.Equal(t, "custom-bucket", cfg.MinIO.Bucket)
	assert.Equal(t, "eu-west-1", cfg.MinIO.Region)
}

func TestLoad_NATSConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("NATS_URL", "nats://nats.example.com:4222")
	os.Setenv("NATS_MAX_RECONNECTS", "20")
	os.Setenv("NATS_RECONNECT_WAIT", "5s")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "nats://nats.example.com:4222", cfg.NATS.URL)
	assert.Equal(t, 20, cfg.NATS.MaxReconnects)
	assert.Equal(t, 5*time.Second, cfg.NATS.ReconnectWait)
}

func TestLoad_ServerConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("SERVER_HOST", "api.example.com")
	os.Setenv("SERVER_PORT", "8443")
	os.Setenv("ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")
	os.Setenv("SERVER_READ_TIMEOUT", "60s")
	os.Setenv("SERVER_WRITE_TIMEOUT", "60s")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "api.example.com", cfg.Server.Host)
	assert.Equal(t, 8443, cfg.Server.Port)
	assert.Equal(t, "https://app.example.com,https://admin.example.com", cfg.Server.AllowedOrigins)
	assert.Equal(t, 60*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.WriteTimeout)
}

func TestLoad_DatabasePoolConfig(t *testing.T) {
	os.Clearenv()

	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("DB_MAX_OPEN_CONNS", "100")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	os.Setenv("DB_CONN_MAX_LIFETIME", "10m")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "5m")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-min-32-chars")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-min-32-chr")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 100, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, 10*time.Minute, cfg.Database.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxIdleTime)
}
