package database

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
)

// Integration tests require a running PostgreSQL instance
// Run with: go test -tags=integration

func setupTestDB(t *testing.T) *DB {
	cfg := config.DatabaseConfig{
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
	}

	db, err := NewPostgresDB(context.Background(), cfg)
	require.NoError(t, err)

	return db
}

func TestNewPostgresDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "opensms_test",
		Password:        "test_password",
		Database:        "opensms_test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db, err := NewPostgresDB(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	// Verify connection
	err = db.Health(context.Background())
	assert.NoError(t, err)
}

func TestNewPostgresDB_InvalidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name string
		cfg  config.DatabaseConfig
	}{
		{
			name: "invalid host",
			cfg: config.DatabaseConfig{
				Host:     "invalid-host-12345",
				Port:     5432,
				User:     "test",
				Password: "test",
				Database: "test",
				SSLMode:  "disable",
			},
		},
		{
			name: "invalid port",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     99999,
				User:     "test",
				Password: "test",
				Database: "test",
				SSLMode:  "disable",
			},
		},
		{
			name: "invalid credentials",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "invalid_user",
				Password: "wrong_password",
				Database: "opensms_test",
				SSLMode:  "disable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := NewPostgresDB(ctx, tt.cfg)
			assert.Error(t, err)
		})
	}
}

func TestDB_Health(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	err := db.Health(context.Background())
	assert.NoError(t, err)
}

func TestDB_HealthAfterClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	db.Close()

	err := db.Health(context.Background())
	assert.Error(t, err)
}

func TestDB_SetTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := "test-tenant-123"

	err := db.SetTenant(ctx, tenantID)
	assert.NoError(t, err)

	// Verify by querying current setting
	var currentTenant string
	err = db.QueryRow(ctx, "SELECT current_setting('app.current_tenant', true)").Scan(&currentTenant)
	assert.NoError(t, err)
	assert.Equal(t, tenantID, currentTenant)
}

func TestWithTenant(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-456"

	// Add tenant to context
	ctxWithTenant := WithTenant(ctx, tenantID)

	// Retrieve tenant from context
	retrievedID, ok := GetTenant(ctxWithTenant)
	assert.True(t, ok)
	assert.Equal(t, tenantID, retrievedID)
}

func TestGetTenant_NoTenant(t *testing.T) {
	ctx := context.Background()

	_, ok := GetTenant(ctx)
	assert.False(t, ok)
}

func TestDB_ConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Test concurrent queries
	done := make(chan bool)
	iterations := 50

	for i := 0; i < iterations; i++ {
		go func() {
			var result int
			err := db.QueryRow(context.Background(), "SELECT 1").Scan(&result)
			assert.NoError(t, err)
			assert.Equal(t, 1, result)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < iterations; i++ {
		<-done
	}
}

func TestDB_Transaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Begin transaction
	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	// Execute query in transaction
	var result int
	err = tx.QueryRow(ctx, "SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	// Commit transaction
	err = tx.Commit(ctx)
	assert.NoError(t, err)
}

func TestDB_TransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Begin transaction
	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	// Execute query
	var result int
	err = tx.QueryRow(ctx, "SELECT 1").Scan(&result)
	assert.NoError(t, err)

	// Rollback transaction
	err = tx.Rollback(ctx)
	assert.NoError(t, err)
}

func TestDB_QueryTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This query should timeout
	var result int
	err := db.QueryRow(ctx, "SELECT pg_sleep(1)").Scan(&result)
	assert.Error(t, err)
}

func TestDB_PreparedStatement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Test parameterized query (prepared statement)
	var result int
	err := db.QueryRow(ctx, "SELECT $1::int + $2::int", 5, 3).Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 8, result)
}

func TestDB_BulkInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create temporary table
	_, err := db.Exec(ctx, `
		CREATE TEMP TABLE test_bulk (
			id SERIAL PRIMARY KEY,
			value TEXT
		)
	`)
	require.NoError(t, err)

	// Bulk insert using batch
	batch := &pgxpool.Batch{}
	for i := 0; i < 100; i++ {
		batch.Queue("INSERT INTO test_bulk (value) VALUES ($1)", "value")
	}

	results := db.SendBatch(ctx, batch)
	defer results.Close()

	// Check all inserts succeeded
	for i := 0; i < 100; i++ {
		_, err := results.Exec()
		assert.NoError(t, err)
	}
}

func TestDB_ConnectionPoolStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	stats := db.Stat()
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.MaxConns(), int32(1))
}

// Benchmark tests
func BenchmarkDB_SimpleQuery(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	db := setupTestDB(&testing.T{})
	defer db.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result int
		_ = db.QueryRow(ctx, "SELECT 1").Scan(&result)
	}
}

func BenchmarkDB_PreparedQuery(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	db := setupTestDB(&testing.T{})
	defer db.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result int
		_ = db.QueryRow(ctx, "SELECT $1::int", i).Scan(&result)
	}
}
