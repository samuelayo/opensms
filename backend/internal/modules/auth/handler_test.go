package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/shared/security"
)

// MockDB is a mock database for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Pool() interface{} {
	return nil
}

func (m *MockDB) SetTenant(ctx context.Context, tenantID string) {
	m.Called(ctx, tenantID)
}

func (m *MockDB) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockEventBus is a mock event bus for testing
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(subject string, data interface{}) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(subject string, handler func([]byte)) error {
	args := m.Called(subject, handler)
	return args.Error(0)
}

func (m *MockEventBus) Health() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEventBus) Close() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestHandler() (*Handler, *MockRepository, *security.JWTManager) {
	mockRepo := NewMockRepository()

	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:       "test-access-secret-key-32-bytes",
			RefreshSecret:      "test-refresh-secret-key-32bytes",
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

	jwtManager := security.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		cfg.JWT.Issuer,
	)

	mockDB := new(MockDB)
	mockEventBus := new(MockEventBus)

	handler := &Handler{
		deps: HandlerDeps{
			DB:       mockDB,
			Cache:    nil,
			EventBus: mockEventBus,
			JWT:      jwtManager,
			Config:   cfg,
			Logger:   zap.NewNop(),
		},
		repo:   mockRepo,
		crypto: &security.Crypto{},
	}

	return handler, mockRepo, jwtManager
}

func TestHandler_Register_Success(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	reqBody := RegisterRequest{
		Email:      "test@example.com",
		Password:   "SecurePass123!@#",
		FirstName:  "John",
		LastName:   "Doe",
		Role:       "student",
		TenantID:   "550e8400-e29b-41d4-a716-446655440000",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, reqBody.Email).Return(nil, ErrUserNotFound)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, reqBody.TenantID).Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "user.registered", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// We expect the handler to be registered in a real router, so we'll call it directly
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err = handler.Register(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestHandler_Register_DuplicateEmail(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	existingUser := &User{
		ID:    "existing-user-id",
		Email: "test@example.com",
	}

	reqBody := RegisterRequest{
		Email:     "test@example.com",
		Password:  "SecurePass123!@#",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "student",
		TenantID:  "550e8400-e29b-41d4-a716-446655440000",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, reqBody.Email).Return(existingUser, nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, reqBody.TenantID).Return()

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err := handler.Register(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_Register_WeakPassword(t *testing.T) {
	handler, _, _ := setupTestHandler()
	app := fiber.New()

	reqBody := RegisterRequest{
		Email:     "test@example.com",
		Password:  "weak",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "student",
		TenantID:  "550e8400-e29b-41d4-a716-446655440000",
	}

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err := handler.Register(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
}

func TestHandler_Login_Success(t *testing.T) {
	handler, mockRepo, jwtManager := setupTestHandler()
	app := fiber.New()

	passwordHash, _ := security.HashPassword("SecurePass123!@#")

	user := &User{
		ID:                  "user-123",
		TenantID:            "tenant-123",
		Email:               "test@example.com",
		PasswordHash:        passwordHash,
		Role:                "student",
		Status:              "active",
		FirstName:           "John",
		LastName:            "Doe",
		EmailVerified:       true,
		FailedLoginAttempts: 0,
		LockedUntil:         nil,
		TwoFactorEnabled:    false,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "SecurePass123!@#",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, reqBody.Email).Return(user, nil)
	mockRepo.On("UpdateLastLogin", mock.Anything, user.ID, mock.Anything).Return(nil)
	mockRepo.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, user.TenantID).Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "user.login", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")
	ctx.Request().Header.Set("X-Forwarded-For", "127.0.0.1")

	err := handler.Login(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	passwordHash, _ := security.HashPassword("CorrectPassword123!@#")

	user := &User{
		ID:                  "user-123",
		TenantID:            "tenant-123",
		Email:               "test@example.com",
		PasswordHash:        passwordHash,
		FailedLoginAttempts: 2,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123!@#",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, reqBody.Email).Return(user, nil)
	mockRepo.On("UpdateLoginAttempts", mock.Anything, user.ID, 3, mock.Anything).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, user.TenantID).Return()

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err := handler.Login(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_Login_AccountLocked(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	lockedUntil := time.Now().Add(15 * time.Minute)

	user := &User{
		ID:                  "user-123",
		TenantID:            "tenant-123",
		Email:               "test@example.com",
		FailedLoginAttempts: 5,
		LockedUntil:         &lockedUntil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "SecurePass123!@#",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, reqBody.Email).Return(user, nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, user.TenantID).Return()

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err := handler.Login(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_Login_UserNotFound(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	reqBody := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "SecurePass123!@#",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, reqBody.Email).Return(nil, ErrUserNotFound)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, mock.Anything).Maybe().Return()

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err := handler.Login(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_RefreshToken_Success(t *testing.T) {
	handler, mockRepo, jwtManager := setupTestHandler()
	app := fiber.New()

	// Create a valid refresh token
	userID := "user-123"
	tokens, err := jwtManager.GenerateTokenPair(userID, "tenant-123", "student", map[string]interface{}{})
	require.NoError(t, err)

	refreshTokenHash := security.HashSHA256(tokens.RefreshToken)

	storedToken := &RefreshToken{
		ID:        "token-123",
		UserID:    userID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	user := &User{
		ID:        userID,
		TenantID:  "tenant-123",
		Email:     "test@example.com",
		Role:      "student",
		Status:    "active",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	reqBody := RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	}

	mockRepo.On("GetRefreshToken", mock.Anything, refreshTokenHash).Return(storedToken, nil)
	mockRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	mockRepo.On("RevokeRefreshToken", mock.Anything, refreshTokenHash, mock.Anything).Return(nil)
	mockRepo.On("CreateRefreshToken", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, user.TenantID).Return()

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err = handler.RefreshToken(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_RefreshToken_ExpiredToken(t *testing.T) {
	handler, mockRepo, jwtManager := setupTestHandler()
	app := fiber.New()

	// Create an expired refresh token
	userID := "user-123"
	tokens, err := jwtManager.GenerateTokenPair(userID, "tenant-123", "student", map[string]interface{}{})
	require.NoError(t, err)

	refreshTokenHash := security.HashSHA256(tokens.RefreshToken)

	storedToken := &RefreshToken{
		ID:        "token-123",
		UserID:    userID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
	}

	reqBody := RefreshTokenRequest{
		RefreshToken: tokens.RefreshToken,
	}

	mockRepo.On("GetRefreshToken", mock.Anything, refreshTokenHash).Return(storedToken, nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, mock.Anything).Maybe().Return()

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err = handler.RefreshToken(ctx)
	app.ReleaseCtx(ctx)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_ForgotPassword_Success(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	user := &User{
		ID:        "user-123",
		TenantID:  "tenant-123",
		Email:     "test@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	reqBody := ForgotPasswordRequest{
		Email: "test@example.com",
	}

	mockRepo.On("GetUserByEmail", mock.Anything, reqBody.Email).Return(user, nil)
	mockRepo.On("SetPasswordResetToken", mock.Anything, reqBody.Email, mock.Anything, mock.Anything).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, user.TenantID).Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "password.reset_requested", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err := handler.ForgotPassword(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestHandler_ResetPassword_Success(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	reqBody := ResetPasswordRequest{
		Token:       "valid-reset-token",
		NewPassword: "NewSecurePass123!@#",
	}

	mockRepo.On("ResetPassword", mock.Anything, reqBody.Token, mock.Anything).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, mock.Anything).Maybe().Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "password.reset", mock.Anything).Return(nil)

	body, _ := json.Marshal(reqBody)
	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().SetBody(body)
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().Header.SetContentType("application/json")

	err := handler.ResetPassword(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandler_VerifyEmail_Success(t *testing.T) {
	handler, mockRepo, _ := setupTestHandler()
	app := fiber.New()

	token := "valid-verification-token"

	mockRepo.On("VerifyEmail", mock.Anything, token).Return(nil)

	mockDB := handler.deps.DB.(*MockDB)
	mockDB.On("SetTenant", mock.Anything, mock.Anything).Maybe().Return()

	mockEventBus := handler.deps.EventBus.(*MockEventBus)
	mockEventBus.On("Publish", "email.verified", mock.Anything).Return(nil)

	ctx := app.AcquireCtx(&fiber.Ctx{})
	ctx.Request().Header.SetMethod("GET")
	ctx.Params("token", token)

	err := handler.VerifyEmail(ctx)
	app.ReleaseCtx(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestToUserDTO(t *testing.T) {
	now := time.Now()
	middleName := "Middle"
	phone := "+1234567890"
	avatarURL := "https://example.com/avatar.jpg"

	user := &User{
		ID:               "user-123",
		TenantID:         "tenant-123",
		Email:            "test@example.com",
		Role:             "student",
		Status:           "active",
		FirstName:        "John",
		LastName:         "Doe",
		MiddleName:       &middleName,
		Phone:            &phone,
		AvatarURL:        &avatarURL,
		EmailVerified:    true,
		TwoFactorEnabled: false,
		Preferences:      map[string]interface{}{"theme": "dark"},
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	dto := ToUserDTO(user)

	assert.Equal(t, user.ID, dto.ID)
	assert.Equal(t, user.TenantID, dto.TenantID)
	assert.Equal(t, user.Email, dto.Email)
	assert.Equal(t, user.Role, dto.Role)
	assert.Equal(t, user.Status, dto.Status)
	assert.Equal(t, user.FirstName, dto.FirstName)
	assert.Equal(t, user.LastName, dto.LastName)
	assert.Equal(t, middleName, *dto.MiddleName)
	assert.Equal(t, phone, *dto.Phone)
	assert.Equal(t, avatarURL, *dto.AvatarURL)
	assert.Equal(t, user.EmailVerified, dto.EmailVerified)
	assert.Equal(t, user.TwoFactorEnabled, dto.TwoFactorEnabled)
	assert.NotEmpty(t, dto.CreatedAt)
	assert.NotEmpty(t, dto.UpdatedAt)
}
