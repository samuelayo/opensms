package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_Struct(t *testing.T) {
	now := time.Now()
	middleName := "Middle"
	phone := "+1234567890"

	user := &User{
		ID:                  "user-123",
		TenantID:            "tenant-123",
		Email:               "test@example.com",
		PasswordHash:        "hashed-password",
		Role:                "student",
		Status:              "active",
		FirstName:           "John",
		LastName:            "Doe",
		MiddleName:          &middleName,
		Phone:               &phone,
		FailedLoginAttempts: 0,
		EmailVerified:       false,
		TwoFactorEnabled:    false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "tenant-123", user.TenantID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "student", user.Role)
	assert.Equal(t, "active", user.Status)
	assert.Equal(t, middleName, *user.MiddleName)
	assert.Equal(t, phone, *user.Phone)
	assert.False(t, user.EmailVerified)
	assert.False(t, user.TwoFactorEnabled)
}

func TestRefreshToken_Struct(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	token := &RefreshToken{
		ID:        "token-123",
		UserID:    "user-123",
		TokenHash: "hashed-token",
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	assert.Equal(t, "token-123", token.ID)
	assert.Equal(t, "user-123", token.UserID)
	assert.Equal(t, "hashed-token", token.TokenHash)
	assert.Equal(t, expiresAt, token.ExpiresAt)
	assert.Nil(t, token.RevokedAt)
	assert.Nil(t, token.ReplacedByToken)
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "user not found",
			err:  ErrUserNotFound,
			want: "user not found",
		},
		{
			name: "email exists",
			err:  ErrEmailExists,
			want: "email already exists",
		},
		{
			name: "invalid credentials",
			err:  ErrInvalidCredentials,
			want: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}
