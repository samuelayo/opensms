package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
)

// DB wraps pgxpool.Pool with additional functionality
type DB struct {
	*pgxpool.Pool
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
		cfg.MaxOpenConns,
		cfg.MaxIdleConns,
		cfg.ConnMaxLifetime,
		cfg.ConnMaxIdleTime,
	)

	// Parse config
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Custom pool configuration for extreme scale
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Connection timeout
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// WithTenant returns a context with tenant ID for row-level security
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenantID)
}

// GetTenant retrieves tenant ID from context
func GetTenant(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantContextKey{}).(string)
	return tenantID, ok
}

type tenantContextKey struct{}

// SetTenant sets the tenant for the current session (row-level security)
func (db *DB) SetTenant(ctx context.Context, tenantID string) error {
	_, err := db.Exec(ctx, "SET app.current_tenant = $1", tenantID)
	return err
}

// Health checks database health
func (db *DB) Health(ctx context.Context) error {
	return db.Ping(ctx)
}
