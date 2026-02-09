package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims represents JWT claims
type TokenClaims struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTManager handles JWT token operations
type JWTManager struct {
	accessSecret       string
	refreshSecret      string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		accessSecret:       accessSecret,
		refreshSecret:      refreshSecret,
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
		issuer:             issuer,
	}
}

// GenerateTokenPair generates access and refresh tokens
func (j *JWTManager) GenerateTokenPair(userID, tenantID, email, role string, permissions []string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(j.accessTokenExpiry)
	refreshExpiry := now.Add(j.refreshTokenExpiry)

	// Generate access token
	accessClaims := TokenClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.accessSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (minimal claims)
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(refreshExpiry),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    j.issuer,
		Subject:   userID,
		ID:        uuid.New().String(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
	}, nil
}

// ValidateAccessToken validates and parses an access token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.accessSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Verify issuer
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Verify expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}

	// Verify not before
	if time.Now().Before(claims.NotBefore.Time) {
		return nil, fmt.Errorf("token not yet valid")
	}

	return claims, nil
}

// ValidateRefreshToken validates and parses a refresh token
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.refreshSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Verify issuer
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Verify expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("refresh token expired")
	}

	return claims, nil
}

// RefreshTokens generates a new token pair using a refresh token
func (j *JWTManager) RefreshTokens(refreshToken string, userID, tenantID, email, role string, permissions []string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verify user ID matches
	if claims.Subject != userID {
		return nil, fmt.Errorf("user ID mismatch")
	}

	// Generate new token pair
	return j.GenerateTokenPair(userID, tenantID, email, role, permissions)
}
