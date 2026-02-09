package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// User represents a user in the system
type User struct {
	ID                     string
	TenantID               string
	Email                  string
	PasswordHash           string
	Role                   string
	Status                 string
	FirstName              string
	LastName               string
	MiddleName             *string
	Phone                  *string
	AvatarURL              *string
	LastLoginAt            *time.Time
	LastLoginIP            *string
	FailedLoginAttempts    int
	LockedUntil            *time.Time
	EmailVerified          bool
	EmailVerificationToken *string
	PasswordResetToken     *string
	PasswordResetExpires   *time.Time
	MustChangePassword     bool
	TwoFactorEnabled       bool
	TwoFactorSecret        *string
	Preferences            map[string]interface{}
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID              string
	UserID          string
	TokenHash       string
	ExpiresAt       time.Time
	CreatedAt       time.Time
	RevokedAt       *time.Time
	ReplacedByToken *string
}

// Repository defines the interface for auth data access
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	UpdateLoginAttempts(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error
	UpdateLastLogin(ctx context.Context, userID string, loginIP string) error
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
	VerifyEmail(ctx context.Context, token string) error
	SetPasswordResetToken(ctx context.Context, email string, token string, expiresAt time.Time) error
	ResetPassword(ctx context.Context, token string, passwordHash string) error
	Enable2FA(ctx context.Context, userID string, secret string) error
	Disable2FA(ctx context.Context, userID string) error

	// Refresh token operations
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string, replacedBy *string) error
	DeleteExpiredTokens(ctx context.Context) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (
			id, tenant_id, email, password_hash, role, status,
			first_name, last_name, middle_name, phone, email_verified,
			must_change_password, two_factor_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Role, user.Status,
		user.FirstName, user.LastName, user.MiddleName, user.Phone, user.EmailVerified,
		user.MustChangePassword, user.TwoFactorEnabled, time.Now(), time.Now(),
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_tenant_id_email_key\" (SQLSTATE 23505)" {
			return ErrEmailExists
		}
		return err
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT
			id, tenant_id, email, password_hash, role, status,
			first_name, last_name, middle_name, phone, avatar_url,
			last_login_at, last_login_ip, failed_login_attempts, locked_until,
			email_verified, email_verification_token, password_reset_token,
			password_reset_expires, must_change_password, two_factor_enabled,
			two_factor_secret, created_at, updated_at
		FROM users
		WHERE email = $1 AND status != 'deleted'
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.FirstName, &user.LastName, &user.MiddleName, &user.Phone, &user.AvatarURL,
		&user.LastLoginAt, &user.LastLoginIP, &user.FailedLoginAttempts, &user.LockedUntil,
		&user.EmailVerified, &user.EmailVerificationToken, &user.PasswordResetToken,
		&user.PasswordResetExpires, &user.MustChangePassword, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT
			id, tenant_id, email, password_hash, role, status,
			first_name, last_name, middle_name, phone, avatar_url,
			last_login_at, last_login_ip, failed_login_attempts, locked_until,
			email_verified, email_verification_token, password_reset_token,
			password_reset_expires, must_change_password, two_factor_enabled,
			two_factor_secret, created_at, updated_at
		FROM users
		WHERE id = $1 AND status != 'deleted'
	`

	user := &User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.FirstName, &user.LastName, &user.MiddleName, &user.Phone, &user.AvatarURL,
		&user.LastLoginAt, &user.LastLoginIP, &user.FailedLoginAttempts, &user.LockedUntil,
		&user.EmailVerified, &user.EmailVerificationToken, &user.PasswordResetToken,
		&user.PasswordResetExpires, &user.MustChangePassword, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users SET
			first_name = $1, last_name = $2, middle_name = $3,
			phone = $4, avatar_url = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, query,
		user.FirstName, user.LastName, user.MiddleName,
		user.Phone, user.AvatarURL, time.Now(), user.ID,
	)

	return err
}

// UpdateLoginAttempts updates failed login attempts and lock status
func (r *PostgresRepository) UpdateLoginAttempts(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	query := `
		UPDATE users SET
			failed_login_attempts = $1,
			locked_until = $2,
			updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, attempts, lockedUntil, time.Now(), userID)
	return err
}

// UpdateLastLogin updates the last login timestamp and IP
func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, userID string, loginIP string) error {
	query := `
		UPDATE users SET
			last_login_at = $1,
			last_login_ip = $2,
			failed_login_attempts = 0,
			updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, time.Now(), loginIP, time.Now(), userID)
	return err
}

// UpdatePassword updates a user's password
func (r *PostgresRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	query := `
		UPDATE users SET
			password_hash = $1,
			must_change_password = false,
			updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, passwordHash, time.Now(), userID)
	return err
}

// VerifyEmail marks an email as verified
func (r *PostgresRepository) VerifyEmail(ctx context.Context, token string) error {
	query := `
		UPDATE users SET
			email_verified = true,
			email_verification_token = NULL,
			updated_at = $1
		WHERE email_verification_token = $2
	`

	result, err := r.db.Exec(ctx, query, time.Now(), token)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("invalid or expired verification token")
	}

	return nil
}

// SetPasswordResetToken sets a password reset token for a user
func (r *PostgresRepository) SetPasswordResetToken(ctx context.Context, email string, token string, expiresAt time.Time) error {
	query := `
		UPDATE users SET
			password_reset_token = $1,
			password_reset_expires = $2,
			updated_at = $3
		WHERE email = $4
	`

	result, err := r.db.Exec(ctx, query, token, expiresAt, time.Now(), email)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ResetPassword resets a user's password using a reset token
func (r *PostgresRepository) ResetPassword(ctx context.Context, token string, passwordHash string) error {
	query := `
		UPDATE users SET
			password_hash = $1,
			password_reset_token = NULL,
			password_reset_expires = NULL,
			updated_at = $2
		WHERE password_reset_token = $3
			AND password_reset_expires > $4
	`

	result, err := r.db.Exec(ctx, query, passwordHash, time.Now(), token, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("invalid or expired reset token")
	}

	return nil
}

// Enable2FA enables two-factor authentication for a user
func (r *PostgresRepository) Enable2FA(ctx context.Context, userID string, secret string) error {
	query := `
		UPDATE users SET
			two_factor_enabled = true,
			two_factor_secret = $1,
			updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, secret, time.Now(), userID)
	return err
}

// Disable2FA disables two-factor authentication for a user
func (r *PostgresRepository) Disable2FA(ctx context.Context, userID string) error {
	query := `
		UPDATE users SET
			two_factor_enabled = false,
			two_factor_secret = NULL,
			updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	return err
}

// CreateRefreshToken inserts a new refresh token
func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, time.Now(),
	)

	return err
}

// GetRefreshToken retrieves a refresh token by its hash
func (r *PostgresRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at, replaced_by_token
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	token := &RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt,
		&token.CreatedAt, &token.RevokedAt, &token.ReplacedByToken,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("token not found")
		}
		return nil, err
	}

	return token, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string, replacedBy *string) error {
	query := `
		UPDATE refresh_tokens SET
			revoked_at = $1,
			replaced_by_token = $2
		WHERE token_hash = $3
	`

	_, err := r.db.Exec(ctx, query, time.Now(), replacedBy, tokenHash)
	return err
}

// DeleteExpiredTokens removes expired refresh tokens
func (r *PostgresRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := r.db.Exec(ctx, query, time.Now())
	return err
}
