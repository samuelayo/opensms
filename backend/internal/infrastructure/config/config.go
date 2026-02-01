package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	NATS     NATSConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
	Security SecurityConfig
	Observability ObservabilityConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host           string
	Port           int
	AllowedOrigins string
	TrustedProxies []string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL           string
	MaxReconnects int
	ReconnectWait time.Duration
}

// MinIOConfig holds MinIO/S3 configuration
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
	Region    string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessSecret       string
	RefreshSecret      string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EncryptionKey          string // AES-256 key (32 bytes)
	RateLimitPerMinute     int
	MaxLoginAttempts       int
	LoginAttemptWindow     time.Duration
	PasswordMinLength      int
	RequireStrongPasswords bool
	SessionTimeout         time.Duration
	CSRFEnabled            bool
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	JaegerEndpoint     string
	PrometheusPort     int
	LogLevel           string
	EnableTracing      bool
	EnableMetrics      bool
	SamplingRate       float64
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (optional, env vars take precedence)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config

	// Server config
	cfg.Server = ServerConfig{
		Host:           viper.GetString("SERVER_HOST"),
		Port:           viper.GetInt("SERVER_PORT"),
		AllowedOrigins: viper.GetString("ALLOWED_ORIGINS"),
		TrustedProxies: viper.GetStringSlice("TRUSTED_PROXIES"),
		ReadTimeout:    viper.GetDuration("SERVER_READ_TIMEOUT"),
		WriteTimeout:   viper.GetDuration("SERVER_WRITE_TIMEOUT"),
	}

	// Database config
	cfg.Database = DatabaseConfig{
		Host:            viper.GetString("DB_HOST"),
		Port:            viper.GetInt("DB_PORT"),
		User:            viper.GetString("DB_USER"),
		Password:        viper.GetString("DB_PASSWORD"),
		Database:        viper.GetString("DB_NAME"),
		SSLMode:         viper.GetString("DB_SSLMODE"),
		MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
		ConnMaxIdleTime: viper.GetDuration("DB_CONN_MAX_IDLE_TIME"),
	}

	// Redis config
	cfg.Redis = RedisConfig{
		Host:     viper.GetString("REDIS_HOST"),
		Port:     viper.GetInt("REDIS_PORT"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
		PoolSize: viper.GetInt("REDIS_POOL_SIZE"),
	}

	// NATS config
	cfg.NATS = NATSConfig{
		URL:           viper.GetString("NATS_URL"),
		MaxReconnects: viper.GetInt("NATS_MAX_RECONNECTS"),
		ReconnectWait: viper.GetDuration("NATS_RECONNECT_WAIT"),
	}

	// MinIO config
	cfg.MinIO = MinIOConfig{
		Endpoint:  viper.GetString("MINIO_ENDPOINT"),
		AccessKey: viper.GetString("MINIO_ACCESS_KEY"),
		SecretKey: viper.GetString("MINIO_SECRET_KEY"),
		UseSSL:    viper.GetBool("MINIO_USE_SSL"),
		Bucket:    viper.GetString("MINIO_BUCKET"),
		Region:    viper.GetString("MINIO_REGION"),
	}

	// JWT config
	cfg.JWT = JWTConfig{
		AccessSecret:       viper.GetString("JWT_ACCESS_SECRET"),
		RefreshSecret:      viper.GetString("JWT_REFRESH_SECRET"),
		AccessTokenExpiry:  viper.GetDuration("JWT_ACCESS_EXPIRY"),
		RefreshTokenExpiry: viper.GetDuration("JWT_REFRESH_EXPIRY"),
		Issuer:             viper.GetString("JWT_ISSUER"),
	}

	// Security config
	cfg.Security = SecurityConfig{
		EncryptionKey:          viper.GetString("ENCRYPTION_KEY"),
		RateLimitPerMinute:     viper.GetInt("RATE_LIMIT_PER_MINUTE"),
		MaxLoginAttempts:       viper.GetInt("MAX_LOGIN_ATTEMPTS"),
		LoginAttemptWindow:     viper.GetDuration("LOGIN_ATTEMPT_WINDOW"),
		PasswordMinLength:      viper.GetInt("PASSWORD_MIN_LENGTH"),
		RequireStrongPasswords: viper.GetBool("REQUIRE_STRONG_PASSWORDS"),
		SessionTimeout:         viper.GetDuration("SESSION_TIMEOUT"),
		CSRFEnabled:            viper.GetBool("CSRF_ENABLED"),
	}

	// Observability config
	cfg.Observability = ObservabilityConfig{
		JaegerEndpoint: viper.GetString("JAEGER_ENDPOINT"),
		PrometheusPort: viper.GetInt("PROMETHEUS_PORT"),
		LogLevel:       viper.GetString("LOG_LEVEL"),
		EnableTracing:  viper.GetBool("ENABLE_TRACING"),
		EnableMetrics:  viper.GetBool("ENABLE_METRICS"),
		SamplingRate:   viper.GetFloat64("TRACING_SAMPLING_RATE"),
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("ALLOWED_ORIGINS", "*")
	viper.SetDefault("SERVER_READ_TIMEOUT", 30*time.Second)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)

	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	viper.SetDefault("DB_CONN_MAX_IDLE_TIME", 10*time.Minute)

	// Redis defaults
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_POOL_SIZE", 10)

	// NATS defaults
	viper.SetDefault("NATS_URL", "nats://localhost:4222")
	viper.SetDefault("NATS_MAX_RECONNECTS", 10)
	viper.SetDefault("NATS_RECONNECT_WAIT", 2*time.Second)

	// MinIO defaults
	viper.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	viper.SetDefault("MINIO_USE_SSL", false)
	viper.SetDefault("MINIO_BUCKET", "opensms")
	viper.SetDefault("MINIO_REGION", "us-east-1")

	// JWT defaults
	viper.SetDefault("JWT_ACCESS_EXPIRY", 15*time.Minute)
	viper.SetDefault("JWT_REFRESH_EXPIRY", 7*24*time.Hour)
	viper.SetDefault("JWT_ISSUER", "opensms")

	// Security defaults
	viper.SetDefault("RATE_LIMIT_PER_MINUTE", 100)
	viper.SetDefault("MAX_LOGIN_ATTEMPTS", 5)
	viper.SetDefault("LOGIN_ATTEMPT_WINDOW", 15*time.Minute)
	viper.SetDefault("PASSWORD_MIN_LENGTH", 12)
	viper.SetDefault("REQUIRE_STRONG_PASSWORDS", true)
	viper.SetDefault("SESSION_TIMEOUT", 30*time.Minute)
	viper.SetDefault("CSRF_ENABLED", true)

	// Observability defaults
	viper.SetDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	viper.SetDefault("PROMETHEUS_PORT", 9090)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("ENABLE_TRACING", true)
	viper.SetDefault("ENABLE_METRICS", true)
	viper.SetDefault("TRACING_SAMPLING_RATE", 0.1)
}

func validate(cfg *Config) error {
	// Validate required fields
	if cfg.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if cfg.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if cfg.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	if cfg.Security.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required (must be 32 bytes for AES-256)")
	}
	if len(cfg.Security.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}

	return nil
}
