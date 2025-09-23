package usecase

import (
	"context"
	"time"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// UserUseCase defines the complete interface for user-related use cases
type UserUseCase interface {
	// CreateOrUpdateUser creates a new user or updates an existing one
	CreateOrUpdateUser(user *entity.User) error

	// GetUserByGoogleID retrieves a user by their Google ID
	GetUserByGoogleID(googleID string) (*entity.User, error)

	// ValidateSession checks if a user's session is valid
	ValidateSession(googleID string) (bool, error)

	// RefreshSession refreshes a user's session using their refresh token
	RefreshSession(refreshToken string) (string, string, time.Time, error)

	// InvalidateSession invalidates a user's session
	InvalidateSession(googleID string) error

	// Register registers a new user with email/phone and password
	Register(name, email, phone, password string, userType entity.UserType, acceptTerms, acceptPromos bool) (*entity.User, string, string, time.Time, error)

	// LoginWithCredentials logs in a user with email/phone and password
	LoginWithCredentials(emailOrPhone, password string) (*entity.User, string, string, time.Time, error)

	// GetUserByEmailOrPhone retrieves a user by email or phone number
	GetUserByEmailOrPhone(emailOrPhone string) (*entity.User, error)

	// DeleteUser soft deletes a user by their ID
	DeleteUser(ctx context.Context, id string) error

	// DeleteUserByEmail soft deletes a user by their email
	DeleteUserByEmail(ctx context.Context, email string) error

	// UndeleteUserByEmail restores a soft-deleted user by their email
	UndeleteUserByEmail(ctx context.Context, email string) error

	// GetUserByID retrieves a user by their ID
	GetUserByID(id string) (*entity.User, error)

	// GetUserProfile retrieves user profile with wallet information
	GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileResponse, error)

	// UpdateUserRole updates a user's role
	UpdateUserRole(ctx context.Context, userID string, role entity.UserRole) error

	// GetUsersByRole retrieves all users with a specific role
	GetUsersByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error)

	// CountUsersByRole counts the number of users with a specific role
	CountUsersByRole(ctx context.Context, role entity.UserRole) (int, error)

	// ForgotPassword initiates password reset process by sending reset email
	ForgotPassword(ctx context.Context, emailOrPhone string) error

	// ResetPassword resets password using reset token
	ResetPassword(ctx context.Context, token, newPassword string) error
}
