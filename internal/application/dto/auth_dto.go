package dto

import (
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// LoginRequest represents the request for login callback
type LoginRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// RegistrationRequest represents the request for user registration
type RegistrationRequest struct {
	Name         string          `json:"name" binding:"required,min=3,max=100"`
	Email        string          `json:"email" binding:"required,email"`
	Phone        string          `json:"phone" binding:"required,min=10,max=20"`
	Password     string          `json:"password" binding:"required,min=8,max=100"`
	UserType     entity.UserType `json:"user_type" binding:"required,oneof=individual business enterprise"`
	AcceptTerms  bool            `json:"accept_terms" binding:"required,eq=true"`
	AcceptPromos bool            `json:"accept_promos"`
}

// LoginCredentialsRequest represents the request for email/phone + password login
type LoginCredentialsRequest struct {
	EmailOrPhone     string `json:"email_or_phone" binding:"required"`
	Password         string `json:"password" binding:"required"`
	UseSecureTokens  bool   `json:"use_secure_tokens,omitempty"` // Optional: if true, tokens will be stored in httpOnly cookies
}

// ForgotPasswordRequest represents the request for password reset
type ForgotPasswordRequest struct {
	EmailOrPhone string `json:"email_or_phone" binding:"required"`
}

// ResetPasswordRequest represents the request for password reset with token
type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// ForgotPasswordResponse represents the response for forgot password request
type ForgotPasswordResponse struct {
	Message   string    `json:"message"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ResetPasswordResponse represents the response for reset password request
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// SetSecureTokensRequest represents the request for setting secure tokens
type SetSecureTokensRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserDTO represents user data
type UserDTO struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Email        string          `json:"email"`
	Phone        string          `json:"phone"`
	UserType     entity.UserType `json:"user_type"`
	UserRole     entity.UserRole `json:"user_role"`
	Picture      string          `json:"picture"`
	AcceptTerms  bool            `json:"accept_terms"`
	AcceptPromos bool            `json:"accept_promos"`
	IsAdmin      bool            `json:"is_admin"` // For backward compatibility
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiry  time.Time `json:"token_expiry"`
	User         UserDTO   `json:"user"`
}
