package auth

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=12"`
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	MiddleName string `json:"middle_name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Role       string `json:"role" validate:"required,oneof=student teacher parent"`
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email         string `json:"email" validate:"required,email"`
	Password      string `json:"password" validate:"required"`
	TwoFactorCode string `json:"two_factor_code,omitempty"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"` // seconds
	User         *UserDTO `json:"user"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=12"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=12"`
}

// Enable2FAResponse represents the 2FA enablement response
type Enable2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// Verify2FARequest represents 2FA verification request
type Verify2FARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// UserDTO represents a user in API responses
type UserDTO struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	Email            string                 `json:"email"`
	Role             string                 `json:"role"`
	Status           string                 `json:"status"`
	FirstName        string                 `json:"first_name"`
	LastName         string                 `json:"last_name"`
	MiddleName       *string                `json:"middle_name,omitempty"`
	Phone            *string                `json:"phone,omitempty"`
	AvatarURL        *string                `json:"avatar_url,omitempty"`
	EmailVerified    bool                   `json:"email_verified"`
	TwoFactorEnabled bool                   `json:"two_factor_enabled"`
	Preferences      map[string]interface{} `json:"preferences,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

// ToUserDTO converts a User to UserDTO
func ToUserDTO(user *User) *UserDTO {
	return &UserDTO{
		ID:               user.ID,
		TenantID:         user.TenantID,
		Email:            user.Email,
		Role:             user.Role,
		Status:           user.Status,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		MiddleName:       user.MiddleName,
		Phone:            user.Phone,
		AvatarURL:        user.AvatarURL,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		Preferences:      user.Preferences,
		CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
