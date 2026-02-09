package auth

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	mock.Mock
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) CreateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) UpdateLoginAttempts(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	args := m.Called(ctx, userID, attempts, lockedUntil)
	return args.Error(0)
}

func (m *MockRepository) UpdateLastLogin(ctx context.Context, userID string, loginIP string) error {
	args := m.Called(ctx, userID, loginIP)
	return args.Error(0)
}

func (m *MockRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *MockRepository) VerifyEmail(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRepository) SetPasswordResetToken(ctx context.Context, email string, token string, expiresAt time.Time) error {
	args := m.Called(ctx, email, token, expiresAt)
	return args.Error(0)
}

func (m *MockRepository) ResetPassword(ctx context.Context, token string, passwordHash string) error {
	args := m.Called(ctx, token, passwordHash)
	return args.Error(0)
}

func (m *MockRepository) Enable2FA(ctx context.Context, userID string, secret string) error {
	args := m.Called(ctx, userID, secret)
	return args.Error(0)
}

func (m *MockRepository) Disable2FA(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RefreshToken), args.Error(1)
}

func (m *MockRepository) RevokeRefreshToken(ctx context.Context, tokenHash string, replacedBy *string) error {
	args := m.Called(ctx, tokenHash, replacedBy)
	return args.Error(0)
}

func (m *MockRepository) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
