package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/samuelayo/opensms/internal/infrastructure/cache"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
	"github.com/samuelayo/opensms/internal/infrastructure/database"
	"github.com/samuelayo/opensms/internal/infrastructure/eventbus"
	"github.com/samuelayo/opensms/internal/shared/security"
)

// HandlerDeps holds dependencies for the auth handler
type HandlerDeps struct {
	DB       *database.DB
	Cache    *cache.RedisClient
	EventBus *eventbus.NATSConnection
	JWT      *security.JWTManager
	Config   *config.Config
	Logger   *zap.Logger
}

// Handler handles authentication requests
type Handler struct {
	deps   HandlerDeps
	repo   Repository
	crypto *security.Crypto
}

// NewHandler creates a new auth handler
func NewHandler(deps HandlerDeps) *Handler {
	repo := NewPostgresRepository(deps.DB.Pool())
	crypto := security.NewCrypto(deps.Config.Security.EncryptionKey)

	return &Handler{
		deps:   deps,
		repo:   repo,
		crypto: crypto,
	}
}

// Register handles user registration
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate password strength
	if len(req.Password) < h.deps.Config.Security.PasswordMinLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "WEAK_PASSWORD",
			"message": fmt.Sprintf("Password must be at least %d characters", h.deps.Config.Security.PasswordMinLength),
		})
	}

	// Set tenant context
	ctx := c.Context()
	if err := h.deps.DB.SetTenant(ctx, req.TenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Check if email already exists
	existingUser, err := h.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && err != ErrUserNotFound {
		h.deps.Logger.Error("Failed to check email", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}
	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"code":    "EMAIL_EXISTS",
			"message": "Email already registered",
		})
	}

	// Hash password
	passwordHash, err := h.crypto.HashPassword(req.Password)
	if err != nil {
		h.deps.Logger.Error("Failed to hash password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process request",
		})
	}

	// Generate email verification token
	verificationToken := generateToken()

	// Create user
	user := &User{
		ID:                     uuid.New().String(),
		TenantID:               req.TenantID,
		Email:                  req.Email,
		PasswordHash:           passwordHash,
		Role:                   req.Role,
		Status:                 "active",
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		MiddleName:             stringPtr(req.MiddleName),
		Phone:                  stringPtr(req.Phone),
		EmailVerified:          false,
		EmailVerificationToken: &verificationToken,
		MustChangePassword:     false,
		TwoFactorEnabled:       false,
	}

	if err := h.repo.CreateUser(ctx, user); err != nil {
		if err == ErrEmailExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "EMAIL_EXISTS",
				"message": "Email already registered",
			})
		}
		h.deps.Logger.Error("Failed to create user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create user",
		})
	}

	// Publish user.registered event
	if h.deps.EventBus != nil {
		eventData := map[string]interface{}{
			"user_id":            user.ID,
			"email":              user.Email,
			"tenant_id":          user.TenantID,
			"role":               user.Role,
			"verification_token": verificationToken,
		}
		if err := h.deps.EventBus.Publish("user.registered", eventData); err != nil {
			h.deps.Logger.Warn("Failed to publish user.registered event", zap.Error(err))
		}
	}

	h.deps.Logger.Info("User registered", zap.String("user_id", user.ID), zap.String("email", user.Email))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Registration successful. Please check your email to verify your account.",
		"user":    ToUserDTO(user),
	})
}

// Login handles user login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Get user by email
	user, err := h.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == ErrUserNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "INVALID_CREDENTIALS",
				"message": "Invalid email or password",
			})
		}
		h.deps.Logger.Error("Failed to get user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process login",
		})
	}

	// Set tenant context
	if err := h.deps.DB.SetTenant(ctx, user.TenantID); err != nil {
		h.deps.Logger.Error("Failed to set tenant", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to process login",
		})
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"code":    "ACCOUNT_LOCKED",
			"message": fmt.Sprintf("Account locked until %s", user.LockedUntil.Format(time.RFC3339)),
		})
	}

	// Validate password
	if err := h.crypto.ValidatePassword(req.Password, user.PasswordHash); err != nil {
		// Increment failed login attempts
		newAttempts := user.FailedLoginAttempts + 1
		var lockedUntil *time.Time

		if newAttempts >= h.deps.Config.Security.MaxLoginAttempts {
			lockTime := time.Now().Add(h.deps.Config.Security.LoginAttemptWindow)
			lockedUntil = &lockTime
		}

		if err := h.repo.UpdateLoginAttempts(ctx, user.ID, newAttempts, lockedUntil); err != nil {
			h.deps.Logger.Error("Failed to update login attempts", zap.Error(err))
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid email or password",
		})
	}

	// Check if 2FA is enabled and code is required
	if user.TwoFactorEnabled {
		if req.TwoFactorCode == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "2FA_REQUIRED",
				"message": "Two-factor authentication code required",
			})
		}
		// TODO: Implement TOTP validation
		// For now, accept any 6-digit code
		if len(req.TwoFactorCode) != 6 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "INVALID_2FA_CODE",
				"message": "Invalid two-factor authentication code",
			})
		}
	}

	// Generate tokens
	permissions := []string{} // TODO: Load from RBAC based on role
	tokens, err := h.deps.JWT.GenerateTokenPair(user.ID, user.TenantID, user.Email, user.Role, permissions)
	if err != nil {
		h.deps.Logger.Error("Failed to generate tokens", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate authentication tokens",
		})
	}

	// Store refresh token in database
	refreshToken := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: hashToken(tokens.RefreshToken),
		ExpiresAt: time.Now().Add(h.deps.Config.JWT.RefreshExpiry),
	}
	if err := h.repo.CreateRefreshToken(ctx, refreshToken); err != nil {
		h.deps.Logger.Error("Failed to store refresh token", zap.Error(err))
		// Continue anyway, user can still use access token
	}

	// Update last login
	clientIP := c.IP()
	if err := h.repo.UpdateLastLogin(ctx, user.ID, clientIP); err != nil {
		h.deps.Logger.Error("Failed to update last login", zap.Error(err))
	}

	// Publish user.login event
	if h.deps.EventBus != nil {
		eventData := map[string]interface{}{
			"user_id":   user.ID,
			"email":     user.Email,
			"tenant_id": user.TenantID,
			"ip":        clientIP,
		}
		if err := h.deps.EventBus.Publish("user.login", eventData); err != nil {
			h.deps.Logger.Warn("Failed to publish user.login event", zap.Error(err))
		}
	}

	h.deps.Logger.Info("User logged in", zap.String("user_id", user.ID), zap.String("email", user.Email))

	return c.JSON(&LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    int(h.deps.Config.JWT.AccessExpiry.Seconds()),
		User:         ToUserDTO(user),
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	// Extract user context
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
	}

	// TODO: Blacklist current access token in Redis
	// TODO: Revoke refresh token

	h.deps.Logger.Info("User logged out", zap.String("user_id", userID))

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Validate refresh token
	claims, err := h.deps.JWT.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_TOKEN",
			"message": "Invalid or expired refresh token",
		})
	}

	// Check if token exists and is not revoked
	tokenHash := hashToken(req.RefreshToken)
	storedToken, err := h.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_TOKEN",
			"message": "Invalid or expired refresh token",
		})
	}

	if storedToken.RevokedAt != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "TOKEN_REVOKED",
			"message": "Refresh token has been revoked",
		})
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "TOKEN_EXPIRED",
			"message": "Refresh token has expired",
		})
	}

	// Get user
	user, err := h.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "USER_NOT_FOUND",
			"message": "User not found",
		})
	}

	// Generate new token pair
	permissions := []string{} // TODO: Load from RBAC
	newTokens, err := h.deps.JWT.GenerateTokenPair(user.ID, user.TenantID, user.Email, user.Role, permissions)
	if err != nil {
		h.deps.Logger.Error("Failed to generate tokens", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate tokens",
		})
	}

	// Revoke old refresh token
	newTokenHash := hashToken(newTokens.RefreshToken)
	if err := h.repo.RevokeRefreshToken(ctx, tokenHash, &newTokenHash); err != nil {
		h.deps.Logger.Error("Failed to revoke old token", zap.Error(err))
	}

	// Store new refresh token
	newRefreshToken := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(h.deps.Config.JWT.RefreshExpiry),
	}
	if err := h.repo.CreateRefreshToken(ctx, newRefreshToken); err != nil {
		h.deps.Logger.Error("Failed to store new refresh token", zap.Error(err))
	}

	return c.JSON(&RefreshTokenResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
		ExpiresIn:    int(h.deps.Config.JWT.AccessExpiry.Seconds()),
	})
}

// ForgotPassword handles password reset request
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Generate reset token
	resetToken := generateToken()
	expiresAt := time.Now().Add(1 * time.Hour)

	// Set password reset token (this will fail silently if email doesn't exist for security)
	if err := h.repo.SetPasswordResetToken(ctx, req.Email, resetToken, expiresAt); err != nil {
		if err != ErrUserNotFound {
			h.deps.Logger.Error("Failed to set reset token", zap.Error(err))
		}
	}

	// Publish password_reset_requested event
	if h.deps.EventBus != nil {
		eventData := map[string]interface{}{
			"email":       req.Email,
			"reset_token": resetToken,
			"expires_at":  expiresAt,
		}
		if err := h.deps.EventBus.Publish("password.reset.requested", eventData); err != nil {
			h.deps.Logger.Warn("Failed to publish password reset event", zap.Error(err))
		}
	}

	// Always return success to prevent email enumeration
	return c.JSON(fiber.Map{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Validate new password
	if len(req.NewPassword) < h.deps.Config.Security.PasswordMinLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "WEAK_PASSWORD",
			"message": fmt.Sprintf("Password must be at least %d characters", h.deps.Config.Security.PasswordMinLength),
		})
	}

	// Hash new password
	passwordHash, err := h.crypto.HashPassword(req.NewPassword)
	if err != nil {
		h.deps.Logger.Error("Failed to hash password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to reset password",
		})
	}

	ctx := c.Context()

	// Reset password
	if err := h.repo.ResetPassword(ctx, req.Token, passwordHash); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_TOKEN",
			"message": "Invalid or expired reset token",
		})
	}

	h.deps.Logger.Info("Password reset successful", zap.String("token", req.Token))

	return c.JSON(fiber.Map{
		"message": "Password reset successful",
	})
}

// VerifyEmail handles email verification
func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Verification token required",
		})
	}

	ctx := c.Context()

	if err := h.repo.VerifyEmail(ctx, token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_TOKEN",
			"message": "Invalid or expired verification token",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Email verified successfully",
	})
}

// GetCurrentUser returns the current authenticated user
func (h *Handler) GetCurrentUser(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
	}

	ctx := c.Context()

	user, err := h.repo.GetUserByID(ctx, userID)
	if err != nil {
		if err == ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			})
		}
		h.deps.Logger.Error("Failed to get user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to get user",
		})
	}

	return c.JSON(ToUserDTO(user))
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
	}

	ctx := c.Context()

	// Get user
	user, err := h.repo.GetUserByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "USER_NOT_FOUND",
			"message": "User not found",
		})
	}

	// Validate current password
	if err := h.crypto.ValidatePassword(req.CurrentPassword, user.PasswordHash); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_PASSWORD",
			"message": "Current password is incorrect",
		})
	}

	// Validate new password
	if len(req.NewPassword) < h.deps.Config.Security.PasswordMinLength {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "WEAK_PASSWORD",
			"message": fmt.Sprintf("Password must be at least %d characters", h.deps.Config.Security.PasswordMinLength),
		})
	}

	// Hash new password
	passwordHash, err := h.crypto.HashPassword(req.NewPassword)
	if err != nil {
		h.deps.Logger.Error("Failed to hash password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to change password",
		})
	}

	// Update password
	if err := h.repo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		h.deps.Logger.Error("Failed to update password", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to change password",
		})
	}

	h.deps.Logger.Info("Password changed", zap.String("user_id", userID))

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// Enable2FA enables two-factor authentication
func (h *Handler) Enable2FA(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
	}

	// TODO: Generate TOTP secret
	// For now, return placeholder
	secret := generateToken()[:32] // 32 character secret

	return c.JSON(&Enable2FAResponse{
		Secret:    secret,
		QRCodeURL: fmt.Sprintf("otpauth://totp/OpenSMS:%s?secret=%s&issuer=OpenSMS", userID, secret),
	})
}

// Verify2FA verifies two-factor authentication
func (h *Handler) Verify2FA(c *fiber.Ctx) error {
	var req Verify2FARequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
	}

	// TODO: Validate TOTP code
	// For now, accept any 6-digit code
	if len(req.Code) != 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_CODE",
			"message": "Code must be 6 digits",
		})
	}

	ctx := c.Context()

	// Enable 2FA
	secret := generateToken()[:32]
	if err := h.repo.Enable2FA(ctx, userID, secret); err != nil {
		h.deps.Logger.Error("Failed to enable 2FA", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to enable two-factor authentication",
		})
	}

	h.deps.Logger.Info("2FA enabled", zap.String("user_id", userID))

	return c.JSON(fiber.Map{
		"message": "Two-factor authentication enabled successfully",
	})
}

// Helper functions

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	// Simple hash for now - in production use crypto/sha256
	return fmt.Sprintf("%x", token)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
