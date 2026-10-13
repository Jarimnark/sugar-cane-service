package userdto

import "time"

type RegisterRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
	Email       string `json:"email"        validate:"required,email"`
	Password    string `json:"password"     validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
	Bio         string `json:"bio"          validate:"max=500"`
}

type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	Bio           string    `json:"bio,omitempty"`
	PhotoURL      string    `json:"photo_url,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

type SessionResponse struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginOutput is the internal result of a successful login or OAuth flow.
// The handler uses SessionID to set the httpOnly cookie; User is written to the response body.
type LoginOutput struct {
	SessionID string
	ExpiresAt time.Time
	User      *UserResponse
}
