package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/email"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// userUseCase implements the UserUseCase interface
type userUseCase struct {
	userRepo     repository.UserRepository
	emailService email.EmailSender // Changed type to interface
}

// NewUserUseCase creates a new instance of UserUseCase
func NewUserUseCase(userRepo repository.UserRepository, emailService email.EmailSender) UserUseCase {
	return &userUseCase{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

// CreateOrUpdateUser creates a new user or updates an existing one
func (uc *userUseCase) CreateOrUpdateUser(user *entity.User) error {
	existingUser, err := uc.userRepo.GetUserByGoogleID(user.GoogleID)
	if err == nil && existingUser != nil {
		// Update existing user
		existingUser.Name = user.Name
		existingUser.Email = user.Email
		existingUser.Picture = user.Picture
		existingUser.AccessToken = user.AccessToken
		existingUser.TokenExpiry = user.TokenExpiry
		if user.RefreshToken != "" {
			existingUser.RefreshToken = user.RefreshToken
		}
		return uc.userRepo.UpdateUser(existingUser)
	}

	// Create new user
	// Check if user is registering during the promotion period (before November 30, 2025)
	// Campaign: Users registering before November 30, 2025 get "Tuan Besar" tier immediately
	campaignEndDate := time.Date(2025, time.November, 30, 23, 59, 59, 0, time.Local)
	currentTime := time.Now()

	if currentTime.Before(campaignEndDate) {
		// User is registering during the campaign period, set to premium tier
		user.UserTier = entity.UserTierPremium
	} else {
		// Default tier for users outside campaign period
		user.UserTier = entity.UserTierBasic
	}

	err = uc.userRepo.CreateUser(user)
	if err != nil {
		return err
	}

	// For new users, send welcome email directly - no queue needed
	isNewUser := err != nil || existingUser == nil
	if isNewUser && user.Email != "" {
		go func() {
			// Send email in a goroutine to avoid blocking
			err := uc.emailService.SendWelcomeEmail(user.Name, user.Email)
			if err != nil {
				// Just log error, don't fail registration
				// In a production app, you might want to log this to a monitoring system
				fmt.Printf("Error sending welcome email: %v\n", err)
			}
		}()
	}

	return nil
}

// GetUserByGoogleID retrieves a user by their Google ID
func (uc *userUseCase) GetUserByGoogleID(googleID string) (*entity.User, error) {
	return uc.userRepo.GetUserByGoogleID(googleID)
}

// ValidateSession checks if a user's session is valid
func (uc *userUseCase) ValidateSession(googleID string) (bool, error) {
	existingUser, err := uc.userRepo.GetUserByGoogleID(googleID)
	if err != nil {
		return false, err
	}
	if existingUser == nil {
		return false, errors.New("user not found")
	}

	// Check if the session is valid
	if existingUser.TokenExpiry.Valid && existingUser.TokenExpiry.Time.After(time.Now()) {
		return true, nil
	}

	return false, nil
}

// RefreshSession refreshes a user's session using their refresh token
func (uc *userUseCase) RefreshSession(refreshToken string) (string, string, time.Time, error) {
	// Step 1: Find the user with this refresh token
	user, err := uc.userRepo.GetUserByRefreshToken(refreshToken)
	if err != nil {
		return "", "", time.Time{}, err
	}

	if user == nil {
		return "", "", time.Time{}, errors.New("invalid refresh token")
	}

	// Step 2: Generate new tokens
	// JWT expiry (24 hours for access token)
	expiryTime := time.Now().Add(24 * time.Hour)

	// Create claims for the JWT
	claims := map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     expiryTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	// Get the secret key from environment
	secretKey := os.Getenv("SESSION_KEY")
	if secretKey == "" {
		secretKey = "your-default-secret-key-for-development-only" // Default for development
	}

	// Create a new signed token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", time.Time{}, err
	}

	// For security, generate a new refresh token (optional but recommended)
	newRefreshToken := uuid.New().String()

	// Step 3: Update the user record with new tokens
	user.AccessToken = accessToken
	user.RefreshToken = newRefreshToken
	user.TokenExpiry = sql.NullTime{
		Time:  expiryTime,
		Valid: true,
	}

	if err := uc.userRepo.UpdateUser(user); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, newRefreshToken, expiryTime, nil
}

// InvalidateSession invalidates a user's session
func (uc *userUseCase) InvalidateSession(googleID string) error {
	existingUser, err := uc.userRepo.GetUserByGoogleID(googleID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// Invalidate tokens
	existingUser.AccessToken = ""
	existingUser.RefreshToken = ""
	existingUser.TokenExpiry = sql.NullTime{Valid: false}

	return uc.userRepo.UpdateUser(existingUser)
}

// Register registers a new user with email/phone and password
func (uc *userUseCase) Register(name, email, phone, password string, userType entity.UserType, acceptTerms, acceptPromos bool) (*entity.User, string, string, time.Time, error) {
	// Validate inputs
	if name == "" {
		return nil, "", "", time.Time{}, errors.New("name is required")
	}

	if email == "" {
		return nil, "", "", time.Time{}, errors.New("email is required")
	}

	if phone == "" {
		return nil, "", "", time.Time{}, errors.New("phone number is required")
	}

	if password == "" {
		return nil, "", "", time.Time{}, errors.New("password is required")
	}

	if !acceptTerms {
		return nil, "", "", time.Time{}, errors.New("terms must be accepted")
	}

	// Check if user with email already exists
	existingUser, err := uc.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error checking existing email: %w", err)
	}
	if existingUser != nil {
		return nil, "", "", time.Time{}, errors.New("email already registered")
	}

	// Check if user with phone already exists
	existingUser, err = uc.userRepo.GetUserByPhone(phone)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error checking existing phone: %w", err)
	}
	if existingUser != nil {
		return nil, "", "", time.Time{}, errors.New("phone already registered")
	}

	// Generate a new salt
	saltBytes, err := utils.GenerateSalt(utils.DefaultPasswordConfig().SaltLength)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error generating salt: %w", err)
	}

	// Hash the password
	passwordHash, err := utils.HashPassword(password, saltBytes, nil)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error hashing password: %w", err)
	}

	// Generate JWT token
	expiryTime := time.Now().Add(24 * time.Hour)
	userID := uuid.New().String()

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"phone":   phone,
		"name":    name,
		"exp":     expiryTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	// Get the secret key from environment
	secretKey := os.Getenv("SESSION_KEY")
	if secretKey == "" {
		secretKey = "your-default-secret-key-for-development-only" // Default for development
	}

	// Create a new signed token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error signing token: %w", err)
	}

	// Generate refresh token
	refreshToken := uuid.New().String()

	// Check if user is registering during the promotion period (before November 30, 2025)
	// Campaign: Users registering before November 30, 2025 get "Premium" tier immediately
	var userTier entity.UserTier = entity.UserTierBasic // Default tier
	campaignEndDate := time.Date(2025, time.November, 30, 23, 59, 59, 0, time.Local)
	currentTime := time.Now()

	if currentTime.Before(campaignEndDate) {
		// User is registering during the campaign period, set to Premium tier
		userTier = entity.UserTierPremium
	}

	// Create new user
	user := &entity.User{
		ID:           userID,
		Name:         name,
		Email:        email,
		Phone:        phone,
		PasswordHash: passwordHash,
		PasswordSalt: utils.EncodeSalt(saltBytes),
		UserType:     userType,
		UserTier:     userTier, // Use the determined tier based on campaign
		AcceptTerms:  acceptTerms,
		AcceptPromos: acceptPromos,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenExpiry:  sql.NullTime{Time: expiryTime, Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user to database
	err = uc.userRepo.CreateUser(user)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error creating user: %w", err)
	}

	// Send welcome email if email is provided - directly without queue
	if email != "" {
		go func() {
			// Send email in a goroutine to avoid blocking
			err := uc.emailService.SendWelcomeEmail(name, email)
			if err != nil {
				// Just log error, don't fail registration
				fmt.Printf("Error sending welcome email: %v\n", err)
			}
		}()
	}

	// Auto-create a wallet for the user - this is critical for cashback processing
	// Using the wallet service via dependency injection
	// Note: Wallet creation can be added here when wallet service is implemented
	// go func(userID string) {
	//     // Create wallet asynchronously
	//     ctx := context.Background()
	//     // wallet creation logic here
	// }(userID)

	return user, accessToken, refreshToken, expiryTime, nil
}

// LoginWithCredentials logs in a user with email/phone and password
func (uc *userUseCase) LoginWithCredentials(emailOrPhone, password string) (*entity.User, string, string, time.Time, error) {
	// Get user by email or phone
	user, err := uc.GetUserByEmailOrPhone(emailOrPhone)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error retrieving user: %w", err)
	}

	if user == nil {
		return nil, "", "", time.Time{}, errors.New("invalid credentials")
	}

	// Decode the salt
	salt, err := utils.DecodeSalt(user.PasswordSalt)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error decoding salt: %w", err)
	}

	// Verify the password
	isValid, err := utils.VerifyPassword(password, user.PasswordHash, salt, nil)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error verifying password: %w", err)
	}

	if !isValid {
		return nil, "", "", time.Time{}, errors.New("invalid credentials")
	}

	// Generate new tokens, similar to RefreshSession
	expiryTime := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"phone":   user.Phone,
		"name":    user.Name,
		"exp":     expiryTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	// Get the secret key from environment
	secretKey := os.Getenv("SESSION_KEY")
	if secretKey == "" {
		secretKey = "your-default-secret-key-for-development-only" // Default for development
	}

	// Create a new signed token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error signing token: %w", err)
	}

	// Generate a new refresh token
	refreshToken := uuid.New().String()

	// Update the user record
	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	user.TokenExpiry = sql.NullTime{
		Time:  expiryTime,
		Valid: true,
	}

	// Save user to database
	err = uc.userRepo.UpdateUser(user)
	if err != nil {
		return nil, "", "", time.Time{}, fmt.Errorf("error updating user: %w", err)
	}

	return user, accessToken, refreshToken, expiryTime, nil
}

// GetUserByEmailOrPhone retrieves a user by email or phone number
func (uc *userUseCase) GetUserByEmailOrPhone(emailOrPhone string) (*entity.User, error) {
	// Try to get user by email first
	user, err := uc.userRepo.GetUserByEmail(emailOrPhone)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	// If not found by email, try phone
	user, err = uc.userRepo.GetUserByPhone(emailOrPhone)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user by their ID
func (uc *userUseCase) DeleteUser(ctx context.Context, id string) error {
	return uc.userRepo.DeleteUser(ctx, id)
}

// DeleteUserByEmail deletes a user by their email
func (uc *userUseCase) DeleteUserByEmail(ctx context.Context, email string) error {
	return uc.userRepo.DeleteUserByEmail(ctx, email)
}

// UndeleteUserByEmail restores a soft-deleted user by email
func (uc *userUseCase) UndeleteUserByEmail(ctx context.Context, email string) error {
	return uc.userRepo.UndeleteUserByEmail(ctx, email)
}

// GetUserByID retrieves a user by their ID
func (uc *userUseCase) GetUserByID(id string) (*entity.User, error) {
	return uc.userRepo.GetUserByID(id)
}

// UpdateUserRole updates a user's role
func (uc *userUseCase) UpdateUserRole(ctx context.Context, userID string, role entity.UserRole) error {
	// Validate that the role is valid
	validRoles := []entity.UserRole{
		entity.UserRoleOwner,
		entity.UserRoleAdmin,
		entity.UserRoleManager,
		entity.UserRoleSupport,
		entity.UserRoleUser,
	}

	isValidRole := false
	for _, validRole := range validRoles {
		if role == validRole {
			isValidRole = true
			break
		}
	}

	if !isValidRole {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Check if user exists
	user, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Update the user's role
	return uc.userRepo.UpdateUserRole(ctx, userID, role)
}

// GetUsersByRole retrieves all users with a specific role
func (uc *userUseCase) GetUsersByRole(ctx context.Context, role entity.UserRole, limit, offset int) ([]*entity.User, error) {
	return uc.userRepo.GetUsersByRole(ctx, role, limit, offset)
}

// CountUsersByRole counts the number of users with a specific role
func (uc *userUseCase) CountUsersByRole(ctx context.Context, role entity.UserRole) (int, error) {
	return uc.userRepo.CountUsersByRole(ctx, role)
}

// GetUserProfile retrieves user profile with wallet information
func (uc *userUseCase) GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileResponse, error) {
	// Get user by ID
	user, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Note: Wallet information can be added here when wallet service is implemented
	// For now, return user profile without wallet information
	response := dto.BuildUserProfileResponse(user, 0.0, "")
	return &response, nil
}

// ForgotPassword initiates password reset process by sending reset email
func (uc *userUseCase) ForgotPassword(ctx context.Context, emailOrPhone string) error {
	// Find user by email or phone
	user, err := uc.GetUserByEmailOrPhone(emailOrPhone)
	if err != nil {
		// Don't reveal if user exists or not for security reasons
		return nil // Return success even if user doesn't exist
	}
	if user == nil {
		// Don't reveal if user exists or not for security reasons
		return nil // Return success even if user doesn't exist
	}

	// Generate secure reset token
	token := utils.GenerateSecureToken(32)

	// Set token expiry to 1 hour from now
	expiry := time.Now().Add(1 * time.Hour)

	// Save reset token to database
	err = uc.userRepo.SetPasswordResetToken(ctx, user.ID, token, expiry)
	if err != nil {
		return fmt.Errorf("failed to set password reset token: %w", err)
	}

	// Send password reset email
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("FRONTEND_URL"), token)
	err = uc.emailService.SendPasswordResetEmail(user.Email, user.Name, resetURL)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Error sending password reset email: %v\n", err)
		// Clean up the token if email failed
		_ = uc.userRepo.ClearPasswordResetToken(ctx, user.ID)
		return fmt.Errorf("failed to send password reset email")
	}

	return nil
}

// ResetPassword resets password using reset token
func (uc *userUseCase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate password strength
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Find user by reset token
	user, err := uc.userRepo.GetUserByPasswordResetToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}
	if user == nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if token is expired
	if user.PasswordResetExpires.Valid && user.PasswordResetExpires.Time.Before(time.Now()) {
		// Clean up expired token
		_ = uc.userRepo.ClearPasswordResetToken(ctx, user.ID)
		return errors.New("reset token has expired")
	}

	// Generate a new unique salt for this password
	config := utils.DefaultPasswordConfig()
	saltBytes, err := utils.GenerateSalt(config.SaltLength)
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the new password with the new salt
	hashedPassword, err := utils.HashPassword(newPassword, saltBytes, config)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Encode salt for storage
	encodedSalt := utils.EncodeSalt(saltBytes)

	// Update password and clear reset token
	err = uc.userRepo.UpdatePassword(ctx, user.ID, hashedPassword, encodedSalt)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear the reset token
	err = uc.userRepo.ClearPasswordResetToken(ctx, user.ID)
	if err != nil {
		// Log error but don't fail since password was updated
		fmt.Printf("Error clearing password reset token: %v\n", err)
	}

	// Clean up any expired tokens (housekeeping)
	go func() {
		_ = uc.userRepo.CleanExpiredPasswordResetTokens(ctx)
	}()

	return nil
}
