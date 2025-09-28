package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/email"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// CustomerPasswordResetService handles password reset functionality for customers
type CustomerPasswordResetService struct {
	customerRepo repository.CustomerRepository
	emailService *email.MailgunService
}

// NewCustomerPasswordResetService creates a new password reset service
func NewCustomerPasswordResetService(
	customerRepo repository.CustomerRepository,
	emailService *email.MailgunService,
) *CustomerPasswordResetService {
	return &CustomerPasswordResetService{
		customerRepo: customerRepo,
		emailService: emailService,
	}
}

// RequestPasswordReset initiates a password reset request for a customer
func (s *CustomerPasswordResetService) RequestPasswordReset(ctx context.Context, storefrontID uuid.UUID, email string) error {
	// Get customer by email
	customer, err := s.customerRepo.GetByEmail(ctx, storefrontID, email)
	if err != nil {
		if errors.IsNotFoundError(err) {
			// Don't reveal if email exists or not for security
			return nil
		}
		return err
	}

	// Generate password reset token
	token, err := s.generateSecureToken()
	if err != nil {
		return errors.NewInternalError("Failed to generate reset token", err)
	}

	// Set password reset token in database (with expiration)
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour
	if err := s.customerRepo.SetPasswordResetToken(ctx, storefrontID, customer.ID, token, expiresAt); err != nil {
		return errors.NewInternalError("Failed to set reset token", err)
	}

	// Send password reset email
	if err := s.sendPasswordResetEmail(customer, token); err != nil {
		return errors.NewInternalError("Failed to send reset email", err)
	}

	return nil
}

// ResetPassword resets a customer's password using the reset token
func (s *CustomerPasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate token format
	if len(token) != 64 { // 32 bytes = 64 hex characters
		return errors.NewValidationError("Invalid reset token format", nil)
	}

	// Validate password strength
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Get customer by reset token
	customer, err := s.customerRepo.GetByPasswordResetToken(ctx, token)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewValidationError("Invalid or expired reset token", nil)
		}
		return err
	}

	// Hash the new password
	config := utils.DefaultPasswordConfig()
	salt, err := utils.GenerateSalt(config.SaltLength)
	if err != nil {
		return errors.NewInternalError("Failed to generate salt", err)
	}
	
	hashedPassword, err := utils.HashPassword(newPassword, salt, config)
	if err != nil {
		return errors.NewInternalError("Failed to hash password", err)
	}

	// Update password in database
	if err := s.customerRepo.UpdatePassword(ctx, customer.StorefrontID, customer.ID, hashedPassword); err != nil {
		return errors.NewInternalError("Failed to update password", err)
	}

	// Clear the reset token
	if err := s.customerRepo.ClearPasswordResetToken(ctx, customer.StorefrontID, customer.ID); err != nil {
		return errors.NewInternalError("Failed to clear reset token", err)
	}

	return nil
}

// ValidateResetToken validates if a reset token is valid and not expired
func (s *CustomerPasswordResetService) ValidateResetToken(ctx context.Context, token string) error {
	// Validate token format
	if len(token) != 64 { // 32 bytes = 64 hex characters
		return errors.NewValidationError("Invalid reset token format", nil)
	}

	// Check if token exists and is valid
	_, err := s.customerRepo.GetByPasswordResetToken(ctx, token)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewValidationError("Invalid or expired reset token", nil)
		}
		return err
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func (s *CustomerPasswordResetService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validatePassword validates password strength requirements
func (s *CustomerPasswordResetService) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.NewValidationError("Password must be at least 8 characters long", nil)
	}
	
	if len(password) > 128 {
		return errors.NewValidationError("Password must be less than 128 characters", nil)
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 32 && char <= 126: // printable ASCII characters
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	if !hasUpper {
		return errors.NewValidationError("Password must contain at least one uppercase letter", nil)
	}
	if !hasLower {
		return errors.NewValidationError("Password must contain at least one lowercase letter", nil)
	}
	if !hasDigit {
		return errors.NewValidationError("Password must contain at least one digit", nil)
	}
	if !hasSpecial {
		return errors.NewValidationError("Password must contain at least one special character", nil)
	}

	return nil
}

// sendPasswordResetEmail sends the password reset email to the customer
func (s *CustomerPasswordResetService) sendPasswordResetEmail(customer *entity.Customer, token string) error {
	if customer.Email == nil {
		return errors.NewValidationError("Customer email is not set", nil)
	}

	// Create reset URL (this should be configurable)
	resetURL := fmt.Sprintf("https://app.smartseller.id/reset-password?token=%s", token)

	subject := "Reset Your Password - SmartSeller"
	
	// Create HTML email body
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #dc3545; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 4px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            <p>We received a request to reset your password for your SmartSeller account. If you made this request, click the button below to reset your password:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
            <p><a href="%s">%s</a></p>
            <div class="warning">
                <strong>Security Notice:</strong>
                <ul>
                    <li>This reset link will expire in 1 hour for security reasons</li>
                    <li>If you didn't request this password reset, please ignore this email</li>
                    <li>Your password will remain unchanged until you create a new one</li>
                </ul>
            </div>
            <p>For security reasons, we recommend choosing a strong password that:</p>
            <ul>
                <li>Is at least 8 characters long</li>
                <li>Contains uppercase and lowercase letters</li>
                <li>Includes numbers and special characters</li>
                <li>Is unique to your SmartSeller account</li>
            </ul>
        </div>
        <div class="footer">
            <p>Best regards,<br>The SmartSeller Security Team</p>
        </div>
    </div>
</body>
</html>`, customer.FirstName, resetURL, resetURL, resetURL)

	// Send email using the email service
	return s.emailService.SendEmail(*customer.Email, subject, htmlBody)
}