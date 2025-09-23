package repository

import (
	"context"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// GetUserByGoogleID retrieves a user by their Google ID
	GetUserByGoogleID(googleID string) (*entity.User, error)

	// GetUserByRefreshToken retrieves a user by refresh token
	GetUserByRefreshToken(refreshToken string) (*entity.User, error)

	// GetUserByEmail retrieves a user by their email
	GetUserByEmail(email string) (*entity.User, error)

	// GetUserByPhone retrieves a user by their phone number
	GetUserByPhone(phone string) (*entity.User, error)

	// GetUserByID retrieves a user by their ID
	GetUserByID(id string) (*entity.User, error)

	// CreateUser creates a new user in the database
	CreateUser(user *entity.User) error

	// UpdateUser updates an existing user in the database
	UpdateUser(user *entity.User) error

	// DeleteUser soft deletes a user from the database by ID
	DeleteUser(ctx context.Context, id string) error

	// DeleteUserByEmail soft deletes a user from the database by email
	DeleteUserByEmail(ctx context.Context, email string) error

	// UndeleteUserByEmail restores a soft-deleted user by email
	UndeleteUserByEmail(ctx context.Context, email string) error

	// IncrementTransactionCount increments a user's transaction count and updates their tier if necessary
	IncrementTransactionCount(userID string) error

	// UpdateUserTier updates a user's tier based on their transaction count
	UpdateUserTier(userID string) error

	// GetUsersByTier retrieves users by their tier
	GetUsersByTier(tier entity.UserTier, limit, offset int) ([]*entity.User, error)

	// CountUsersByTier counts the number of users in a specific tier
	CountUsersByTier(tier entity.UserTier) (int, error)

	// UpdateUserRole updates a user's role
	UpdateUserRole(ctx context.Context, userID string, role entity.UserRole) error

	// GetUsersByRole retrieves users with a specific role
	GetUsersByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error)

	// CountUsersByRole counts the number of users with a specific role
	CountUsersByRole(ctx context.Context, role entity.UserRole) (int, error)

	// GetAllUsersWithFilters retrieves users with pagination, search, and filters
	GetAllUsersWithFilters(ctx context.Context, req *entity.GetUsersRequest) ([]*entity.User, error)

	// CountUsersWithFilters counts users matching the filters
	CountUsersWithFilters(ctx context.Context, req *entity.GetUsersRequest) (int, error)

	// SetPasswordResetToken sets a password reset token for a user
	SetPasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error

	// GetUserByPasswordResetToken retrieves a user by password reset token
	GetUserByPasswordResetToken(ctx context.Context, token string) (*entity.User, error)

	// ClearPasswordResetToken clears the password reset token for a user
	ClearPasswordResetToken(ctx context.Context, userID string) error

	// UpdatePassword updates user's password hash and salt
	UpdatePassword(ctx context.Context, userID, passwordHash, passwordSalt string) error

	// CleanExpiredPasswordResetTokens removes expired password reset tokens
	CleanExpiredPasswordResetTokens(ctx context.Context) error
}
