package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// UserRepository defines the interface for user database operations
type UserRepositoryInterface interface {
	AddUser(ctx context.Context, user *entity.User) error
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id string) error
	UpdateUserRole(ctx context.Context, userID string, role entity.UserRole) error
	GetUsersByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error)
	CountUsersByRole(ctx context.Context, role entity.UserRole) (int, error)
	GetAllUsersWithFilters(ctx context.Context, req *entity.GetUsersRequest) ([]*entity.User, error)
	CountUsersWithFilters(ctx context.Context, req *entity.GetUsersRequest) (int, error)
}

// UserRepositoryImpl implements the UserRepository interface using sqlx
type UserRepositoryImpl struct {
	db *sqlx.DB
}

// NewUserRepositoryImpl creates a new instance of UserRepository using sqlx
func NewUserRepositoryImpl(db *sqlx.DB) repository.UserRepository {
	return &UserRepositoryImpl{
		db: db,
	}
}

// GetUserByGoogleID retrieves a user by their Google ID (excluding soft-deleted users)
func (r *UserRepositoryImpl) GetUserByGoogleID(googleID string) (*entity.User, error) {
	var user entity.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE google_id = $1 AND deleted_at IS NULL`, googleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by Google ID: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email (excluding soft-deleted users)
func (r *UserRepositoryImpl) GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL`, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}
	return &user, nil
}

// GetUserByPhone retrieves a user by their phone number (excluding soft-deleted users)
func (r *UserRepositoryImpl) GetUserByPhone(phone string) (*entity.User, error) {
	var user entity.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE phone = $1 AND deleted_at IS NULL`, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by phone: %w", err)
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID (excluding soft-deleted users)
func (r *UserRepositoryImpl) GetUserByID(id string) (*entity.User, error) {
	var user entity.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}
	return &user, nil
}

// CreateUser creates a new user in the database
func (r *UserRepositoryImpl) CreateUser(user *entity.User) error {
	query := `
		INSERT INTO users (
			id, name, email, phone, google_id, picture, 
			password_hash, password_salt, user_type, user_tier, accept_terms, accept_promos,
			refresh_token, access_token, token_expiry, created_at, updated_at
		) VALUES (
			:id, :name, LOWER(:email), :phone, :google_id, :picture, 
			:password_hash, :password_salt, :user_type, :user_tier, :accept_terms, :accept_promos,
			:refresh_token, :access_token, :token_expiry, NOW(), NOW()
		)
	`

	_, err := r.db.NamedExec(query, user)

	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user in the database
func (r *UserRepositoryImpl) UpdateUser(user *entity.User) error {
	query := `
		UPDATE users SET
			name = :name,
			email = LOWER(:email),
			phone = :phone,
			picture = :picture,
			password_hash = :password_hash,
			password_salt = :password_salt,
			user_type = :user_type,
			user_tier = :user_tier,
			accept_terms = :accept_terms,
			accept_promos = :accept_promos,
			refresh_token = :refresh_token,
			access_token = :access_token,
			token_expiry = :token_expiry,
			updated_at = NOW()
		WHERE id = :id
	`

	_, err := r.db.NamedExec(query, user)

	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

// DeleteUser soft deletes a user from the database by ID
func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error soft deleting user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID %s", id)
	}

	return nil
}

// DeleteUserByEmail soft deletes a user from the database by email
func (r *UserRepositoryImpl) DeleteUserByEmail(ctx context.Context, email string) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE email = $1`
	result, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("error soft deleting user by email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with email %s", email)
	}

	return nil
}

// UndeleteUserByEmail restores a soft-deleted user by their email
func (r *UserRepositoryImpl) UndeleteUserByEmail(ctx context.Context, email string) error {
	query := `UPDATE users SET deleted_at = NULL WHERE email = $1 AND deleted_at IS NOT NULL`
	result, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("error undeleting user by email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no deleted user found with email %s", email)
	}

	return nil
}

// GetUserByRefreshToken retrieves a user by their refresh token (excluding soft-deleted users)
func (r *UserRepositoryImpl) GetUserByRefreshToken(refreshToken string) (*entity.User, error) {
	var user entity.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE refresh_token = $1 AND deleted_at IS NULL`, refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by refresh token: %w", err)
	}
	return &user, nil
}

// GetUsersByTier retrieves users by their tier with pagination
func (r *UserRepositoryImpl) GetUsersByTier(tier entity.UserTier, limit, offset int) ([]*entity.User, error) {
	users := []*entity.User{}

	query := `
		SELECT * FROM users
		WHERE user_tier = $1
		ORDER BY transaction_count DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.Select(&users, query, tier, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting users by tier: %w", err)
	}

	return users, nil
}

// CountUsersByTier counts the number of users in a specific tier
func (r *UserRepositoryImpl) CountUsersByTier(tier entity.UserTier) (int, error) {
	var count int
	var err error

	if tier == "" {
		// Count users with no tier (NULL or empty string)
		query := `
			SELECT COUNT(*) FROM users 
			WHERE user_tier IS NULL OR user_tier = '' OR user_tier = 'null'
		`
		err = r.db.Get(&count, query)
	} else {
		// Count users with the specified tier
		query := `
			SELECT COUNT(*) FROM users 
			WHERE user_tier = $1
		`
		err = r.db.Get(&count, query, tier)
	}

	if err != nil {
		return 0, fmt.Errorf("error counting users by tier: %w", err)
	}

	return count, nil
}

// IncrementTransactionCount increments a user's transaction count
func (r *UserRepositoryImpl) IncrementTransactionCount(userID string) error {
	ctx := context.Background()

	// Get current transaction count
	var currentCount int
	query := `SELECT transaction_count FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&currentCount)
	if err != nil {
		logger.Error("get_user_data_failed", "Failed to get user data", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to get user data: %w", err)
	}

	// Increment transaction count
	currentCount++
	updateQuery := `UPDATE users SET transaction_count = $1 WHERE id = $2`

	_, err = r.db.ExecContext(ctx, updateQuery, currentCount, userID)
	if err != nil {
		logger.Error("update_transaction_count_failed", "Failed to update transaction count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to update transaction count: %w", err)
	}

	return nil
}

// determineUserTier determines the appropriate tier based on transaction count
func determineUserTier(transactionCount int) entity.UserTier {
	if transactionCount >= 100 {
		return entity.UserTierEnterprise
	} else if transactionCount >= 50 {
		return entity.UserTierPro
	} else if transactionCount >= 20 {
		return entity.UserTierPremium
	} else {
		return entity.UserTierBasic
	}
}

// UpdateUserTier updates a user's tier based on their transaction count
func (r *UserRepositoryImpl) UpdateUserTier(userID string) error {
	ctx := context.Background()

	// First get the user's transaction count
	var transactionCount int
	query := `SELECT transaction_count FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&transactionCount)
	if err != nil {
		logger.Error("get_user_transaction_count_failed", "Failed to get user transaction count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to get user transaction count: %w", err)
	}

	// Determine the appropriate tier based on transaction count
	newTier := determineUserTier(transactionCount)

	// Update the user's tier
	updateQuery := `UPDATE users SET user_tier = $1 WHERE id = $2`
	_, err = r.db.ExecContext(ctx, updateQuery, string(newTier), userID)
	if err != nil {
		logger.Error("update_user_tier_failed", "Failed to update user tier", map[string]interface{}{
			"user_id": userID,
			"tier":    newTier,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to update user tier: %w", err)
	}

	return nil
}

// UpdateUserRole updates a user's role in the database
func (r *UserRepositoryImpl) UpdateUserRole(ctx context.Context, userID string, role entity.UserRole) error {
	query := `UPDATE users SET user_role = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, role, userID)
	if err != nil {
		return fmt.Errorf("error updating user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID %s", userID)
	}

	return nil
}

// GetUsersByRole retrieves users by their role with pagination
func (r *UserRepositoryImpl) GetUsersByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error) {
	users := []*entity.User{}

	query := `
		SELECT * FROM users
		WHERE user_role = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &users, query, role, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting users by role: %w", err)
	}

	return users, nil
}

// CountUsersByRole counts the number of users with a specific role
func (r *UserRepositoryImpl) CountUsersByRole(ctx context.Context, role entity.UserRole) (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM users WHERE user_role = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, role)
	if err != nil {
		return 0, fmt.Errorf("error counting users by role: %w", err)
	}

	return count, nil
}

// GetAllUsersWithFilters retrieves users with pagination, search, and filters
func (r *UserRepositoryImpl) GetAllUsersWithFilters(ctx context.Context, req *entity.GetUsersRequest) ([]*entity.User, error) {
	var users []*entity.User

	// Build the base query
	query := `SELECT id, google_id, name, email, phone, picture, user_type, user_tier, 
			  transaction_count, is_admin, user_role, accept_terms, accept_promos, 
			  created_at, updated_at 
			  FROM users WHERE deleted_at IS NULL`

	var args []interface{}
	argIndex := 1

	// Add search condition
	if req.Search != "" {
		query += fmt.Sprintf(" AND (email ILIKE $%d OR phone ILIKE $%d)", argIndex, argIndex+1)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// Add user type filter
	if req.UserType != nil {
		query += fmt.Sprintf(" AND user_type = $%d", argIndex)
		args = append(args, *req.UserType)
		argIndex++
	}

	// Add user tier filter
	if req.UserTier != nil {
		query += fmt.Sprintf(" AND user_tier = $%d", argIndex)
		args = append(args, *req.UserTier)
		argIndex++
	}

	// Add user role filter
	if req.UserRole != nil {
		query += fmt.Sprintf(" AND user_role = $%d", argIndex)
		args = append(args, *req.UserRole)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	// Calculate offset
	offset := (req.Page - 1) * req.Limit
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, req.Limit, offset)

	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting users with filters: %w", err)
	}

	return users, nil
}

// CountUsersWithFilters counts users matching the filters
func (r *UserRepositoryImpl) CountUsersWithFilters(ctx context.Context, req *entity.GetUsersRequest) (int, error) {
	var count int

	// Build the count query
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	var args []interface{}
	argIndex := 1

	// Add search condition
	if req.Search != "" {
		query += fmt.Sprintf(" AND (email ILIKE $%d OR phone ILIKE $%d)", argIndex, argIndex+1)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// Add user type filter
	if req.UserType != nil {
		query += fmt.Sprintf(" AND user_type = $%d", argIndex)
		args = append(args, *req.UserType)
		argIndex++
	}

	// Add user tier filter
	if req.UserTier != nil {
		query += fmt.Sprintf(" AND user_tier = $%d", argIndex)
		args = append(args, *req.UserTier)
		argIndex++
	}

	// Add user role filter
	if req.UserRole != nil {
		query += fmt.Sprintf(" AND user_role = $%d", argIndex)
		args = append(args, *req.UserRole)
		argIndex++
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error counting users with filters: %w", err)
	}

	return count, nil
}

// SetPasswordResetToken sets a password reset token for a user
func (r *UserRepositoryImpl) SetPasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `UPDATE users SET password_reset_token = $1, password_reset_expires = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, token, expiresAt, userID)
	if err != nil {
		return fmt.Errorf("error setting password reset token: %w", err)
	}
	return nil
}

// GetUserByPasswordResetToken retrieves a user by password reset token (excluding soft-deleted users)
func (r *UserRepositoryImpl) GetUserByPasswordResetToken(ctx context.Context, token string) (*entity.User, error) {
	var user entity.User
	query := `SELECT * FROM users WHERE password_reset_token = $1 AND password_reset_expires > CURRENT_TIMESTAMP AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Token not found or expired
		}
		return nil, fmt.Errorf("error getting user by password reset token: %w", err)
	}
	return &user, nil
}

// ClearPasswordResetToken clears the password reset token for a user
func (r *UserRepositoryImpl) ClearPasswordResetToken(ctx context.Context, userID string) error {
	query := `UPDATE users SET password_reset_token = NULL, password_reset_expires = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error clearing password reset token: %w", err)
	}
	return nil
}

// UpdatePassword updates user's password hash and salt
func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID, passwordHash, passwordSalt string) error {
	query := `UPDATE users SET password_hash = $1, password_salt = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, passwordHash, passwordSalt, userID)
	if err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}
	return nil
}

// CleanExpiredPasswordResetTokens removes expired password reset tokens
func (r *UserRepositoryImpl) CleanExpiredPasswordResetTokens(ctx context.Context) error {
	query := `UPDATE users SET password_reset_token = NULL, password_reset_expires = NULL, updated_at = CURRENT_TIMESTAMP WHERE password_reset_expires <= CURRENT_TIMESTAMP`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error cleaning expired password reset tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Warn("cleanup_password_reset_tokens", "Could not get rows affected count", err)
	} else if rowsAffected > 0 {
		logger.Info("cleanup_password_reset_tokens", map[string]interface{}{
			"expired_tokens_cleaned": rowsAffected,
		})
	}

	return nil
}
