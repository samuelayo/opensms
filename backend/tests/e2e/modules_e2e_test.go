// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/infrastructure/storage"
	"github.com/samuelayo/opensms/internal/server"
)

// E2ETestSuite holds shared test infrastructure for e2e tests
type E2ETestSuite struct {
	app          *fiber.App
	db           *database.DB
	cache        *cache.RedisClient
	eventBus     *eventbus.NATSConnection
	storage      *storage.MinIOClient
	config       *config.Config
	logger       *zap.Logger
	accessToken  string
	refreshToken string
	tenantID     string
	userID       string
}

func setupE2ETest(t *testing.T) *E2ETestSuite {
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
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       2,
			PoolSize: 10,
		},
		NATS: config.NATSConfig{
			URL:              "nats://localhost:4222",
			MaxReconnect:     5,
			ReconnectWait:    2 * time.Second,
			PingInterval:     1 * time.Minute,
			MaxPingsOut:      2,
			ConnectTimeout:   5 * time.Second,
		},
		MinIO: config.MinIOConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			UseSSL:    false,
			Bucket:    "opensms-test",
		},
		JWT: config.JWTConfig{
			AccessSecret:       "test-access-secret-key-must-be-32-chars-min",
			RefreshSecret:      "test-refresh-secret-key-must-be-32-chars",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "opensms-test",
		},
		Security: config.SecurityConfig{
			EncryptionKey:       "12345678901234567890123456789012",
			MaxLoginAttempts:    5,
			LoginAttemptWindow:  15 * time.Minute,
			PasswordMinLength:   12,
			RequireStrongPasswd: true,
		},
	}

	// Setup infrastructure
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	db, err := database.NewPostgresDB(ctx, cfg.Database)
	require.NoError(t, err)

	redisClient := cache.NewRedisClient(cfg.Redis)
	err = redisClient.Health(ctx)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	natsConn, err := eventbus.NewNATSConnection(cfg.NATS)
	if err != nil {
		t.Skipf("NATS not available: %v", err)
	}

	minioClient, err := storage.NewMinIOClient(ctx, cfg.MinIO)
	if err != nil {
		t.Skipf("MinIO not available: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: server.ErrorHandler,
	})

	// Setup server dependencies
	serverDeps := server.Dependencies{
		Config:   cfg,
		DB:       db,
		Cache:    redisClient,
		EventBus: natsConn,
		Storage:  minioClient,
		Logger:   logger,
	}

	srv := server.NewServer(serverDeps)
	srv.RegisterRoutes(app)

	return &E2ETestSuite{
		app:      app,
		db:       db,
		cache:    redisClient,
		eventBus: natsConn,
		storage:  minioClient,
		config:   cfg,
		logger:   logger,
		tenantID: "550e8400-e29b-41d4-a716-446655440000",
	}
}

func (s *E2ETestSuite) teardown() {
	if s.db != nil {
		s.db.Close()
	}
	if s.cache != nil {
		s.cache.Close()
	}
	if s.eventBus != nil {
		s.eventBus.Close()
	}
}

// Auth Module E2E Tests

func TestE2E_Auth_RegisterAndLogin(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	// Step 1: Register a new user
	registerPayload := map[string]interface{}{
		"email":      fmt.Sprintf("test-%d@example.com", time.Now().Unix()),
		"password":   "SecurePass123!@#",
		"first_name": "Test",
		"last_name":  "User",
		"role":       "student",
		"tenant_id":  suite.tenantID,
	}

	registerBody, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Step 2: Login with registered credentials
	loginPayload := map[string]interface{}{
		"email":    registerPayload["email"],
		"password": registerPayload["password"],
	}

	loginBody, _ := json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err = suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	assert.NotEmpty(t, loginResp["access_token"])
	assert.NotEmpty(t, loginResp["refresh_token"])
	assert.NotEmpty(t, loginResp["user"])

	suite.accessToken = loginResp["access_token"].(string)
	suite.refreshToken = loginResp["refresh_token"].(string)
}

func TestE2E_Auth_RefreshToken(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	// First, register and login to get tokens
	registerPayload := map[string]interface{}{
		"email":      fmt.Sprintf("refresh-%d@example.com", time.Now().Unix()),
		"password":   "SecurePass123!@#",
		"first_name": "Refresh",
		"last_name":  "User",
		"role":       "teacher",
		"tenant_id":  suite.tenantID,
	}

	registerBody, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	suite.app.Test(req, -1)

	loginPayload := map[string]interface{}{
		"email":    registerPayload["email"],
		"password": registerPayload["password"],
	}

	loginBody, _ := json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := suite.app.Test(req, -1)
	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	refreshToken := loginResp["refresh_token"].(string)

	// Step 3: Use refresh token to get new access token
	refreshPayload := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	refreshBody, _ := json.Marshal(refreshPayload)
	req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(refreshBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var refreshResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&refreshResp)

	assert.NotEmpty(t, refreshResp["access_token"])
	assert.NotEmpty(t, refreshResp["refresh_token"])
}

func TestE2E_Auth_ForgotPassword(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	// Register a user first
	email := fmt.Sprintf("forgot-%d@example.com", time.Now().Unix())
	registerPayload := map[string]interface{}{
		"email":      email,
		"password":   "SecurePass123!@#",
		"first_name": "Forgot",
		"last_name":  "Password",
		"role":       "parent",
		"tenant_id":  suite.tenantID,
	}

	registerBody, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	suite.app.Test(req, -1)

	// Request password reset
	forgotPayload := map[string]interface{}{
		"email": email,
	}

	forgotBody, _ := json.Marshal(forgotPayload)
	req = httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewReader(forgotBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Users Module E2E Tests

func TestE2E_Users_CRUD(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	// Login as admin first
	suite.loginAsAdmin(t)

	// Create a new user
	createPayload := map[string]interface{}{
		"email":      fmt.Sprintf("crud-%d@example.com", time.Now().Unix()),
		"password":   "SecurePass123!@#",
		"first_name": "CRUD",
		"last_name":  "Test",
		"role":       "student",
		"tenant_id":  suite.tenantID,
	}

	createBody, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResp)
	userID := createResp["id"].(string)

	// Get the user
	req = httptest.NewRequest("GET", "/api/v1/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err = suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Update the user
	updatePayload := map[string]interface{}{
		"first_name": "Updated",
		"last_name":  "Name",
	}

	updateBody, _ := json.Marshal(updatePayload)
	req = httptest.NewRequest("PUT", "/api/v1/users/"+userID, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err = suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Delete the user
	req = httptest.NewRequest("DELETE", "/api/v1/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err = suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestE2E_Students_CreateAndList(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	suite.loginAsAdmin(t)

	// Create a student
	createPayload := map[string]interface{}{
		"email":            fmt.Sprintf("student-%d@example.com", time.Now().Unix()),
		"password":         "SecurePass123!@#",
		"first_name":       "Student",
		"last_name":        "Test",
		"tenant_id":        suite.tenantID,
		"school_id":        "550e8400-e29b-41d4-a716-446655440001",
		"admission_number": fmt.Sprintf("ADM%d", time.Now().Unix()),
		"admission_date":   "2024-01-01",
		"date_of_birth":    "2005-05-15",
		"gender":           "male",
	}

	createBody, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/v1/students", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// List students
	req = httptest.NewRequest("GET", "/api/v1/students", nil)
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err = suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResp)

	assert.NotNil(t, listResp["data"])
}

// Academic Module E2E Tests

func TestE2E_Academic_GradesCRUD(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	suite.loginAsAdmin(t)

	// Create a grade
	createPayload := map[string]interface{}{
		"tenant_id":       suite.tenantID,
		"student_id":      "550e8400-e29b-41d4-a716-446655440002",
		"class_id":        "550e8400-e29b-41d4-a716-446655440003",
		"subject_id":      "550e8400-e29b-41d4-a716-446655440004",
		"assessment_type": "exam",
		"score":           85.5,
		"max_score":       100.0,
		"grade":           "A",
		"grade_point":     4.0,
	}

	createBody, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/v1/grades", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResp)
	gradeID := createResp["id"].(string)

	// Update the grade
	updatePayload := map[string]interface{}{
		"score": 90.0,
		"grade": "A+",
	}

	updateBody, _ := json.Marshal(updatePayload)
	req = httptest.NewRequest("PUT", "/api/v1/grades/"+gradeID, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err = suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestE2E_Academic_AttendanceMarking(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	suite.loginAsAdmin(t)

	// Mark attendance for a single student
	markPayload := map[string]interface{}{
		"tenant_id":  suite.tenantID,
		"student_id": "550e8400-e29b-41d4-a716-446655440002",
		"class_id":   "550e8400-e29b-41d4-a716-446655440003",
		"date":       time.Now().Format("2006-01-02"),
		"status":     "present",
	}

	markBody, _ := json.Marshal(markPayload)
	req := httptest.NewRequest("POST", "/api/v1/attendance", bytes.NewReader(markBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestE2E_Academic_BulkAttendance(t *testing.T) {
	suite := setupE2ETest(t)
	defer suite.teardown()

	suite.loginAsAdmin(t)

	// Bulk mark attendance
	bulkPayload := map[string]interface{}{
		"tenant_id": suite.tenantID,
		"class_id":  "550e8400-e29b-41d4-a716-446655440003",
		"date":      time.Now().Format("2006-01-02"),
		"records": []map[string]interface{}{
			{"student_id": "550e8400-e29b-41d4-a716-446655440005", "status": "present"},
			{"student_id": "550e8400-e29b-41d4-a716-446655440006", "status": "absent"},
			{"student_id": "550e8400-e29b-41d4-a716-446655440007", "status": "late"},
		},
	}

	bulkBody, _ := json.Marshal(bulkPayload)
	req := httptest.NewRequest("POST", "/api/v1/attendance/bulk", bytes.NewReader(bulkBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.accessToken)

	resp, err := suite.app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var bulkResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&bulkResp)

	assert.Equal(t, float64(3), bulkResp["count"])
}

// Helper methods

func (s *E2ETestSuite) loginAsAdmin(t *testing.T) {
	// Register admin user
	registerPayload := map[string]interface{}{
		"email":      fmt.Sprintf("admin-%d@example.com", time.Now().Unix()),
		"password":   "AdminPass123!@#",
		"first_name": "Admin",
		"last_name":  "User",
		"role":       "student", // In production, this would be 'admin'
		"tenant_id":  s.tenantID,
	}

	registerBody, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	s.app.Test(req, -1)

	// Login
	loginPayload := map[string]interface{}{
		"email":    registerPayload["email"],
		"password": registerPayload["password"],
	}

	loginBody, _ := json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.app.Test(req, -1)
	require.NoError(t, err)

	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	s.accessToken = loginResp["access_token"].(string)
	s.refreshToken = loginResp["refresh_token"].(string)
	if user, ok := loginResp["user"].(map[string]interface{}); ok {
		s.userID = user["id"].(string)
	}
}
