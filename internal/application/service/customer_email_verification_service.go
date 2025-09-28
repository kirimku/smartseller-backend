package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/email"
)

// CustomerEmailVerificationService handles email verification for customers
type CustomerEmailVerificationService struct {
	customerRepo repository.CustomerRepository
	emailService *email.MailgunService
}

// NewCustomerEmailVerificationService creates a new email verification service
func NewCustomerEmailVerificationService(
	customerRepo repository.CustomerRepository,
	emailService *email.MailgunService,
) *CustomerEmailVerificationService {
	return &CustomerEmailVerificationService{
		customerRepo: customerRepo,
		emailService: emailService,
	}
}

// SendVerificationEmail sends an email verification to a customer
func (s *CustomerEmailVerificationService) SendVerificationEmail(ctx context.Context, storefrontID, customerID uuid.UUID) error {
	// Get customer
	customer, err := s.customerRepo.GetByID(ctx, storefrontID, customerID)
	if err != nil {
		return err
	}

	// Verify customer belongs to storefront
	if customer.StorefrontID != storefrontID {
		return errors.NewNotFoundError("Customer not found")
	}

	// Check if email is already verified
	if customer.IsEmailVerified() {
		return errors.NewValidationError("Email is already verified", nil)
	}

	// Generate verification token
	token, err := s.generateSecureToken()
	if err != nil {
		return errors.NewInternalError("Failed to generate verification token", err)
	}

	// Set verification token in database
	if err := s.customerRepo.SetEmailVerificationToken(ctx, storefrontID, customerID, token); err != nil {
		return errors.NewInternalError("Failed to set verification token", err)
	}

	// Send verification email
	if err := s.sendVerificationEmail(customer, token); err != nil {
		return errors.NewInternalError("Failed to send verification email", err)
	}

	return nil
}

// VerifyEmail verifies a customer's email using the verification token
func (s *CustomerEmailVerificationService) VerifyEmail(ctx context.Context, token string) error {
	// Validate token format
	if len(token) != 64 { // 32 bytes = 64 hex characters
		return errors.NewValidationError("Invalid verification token format", nil)
	}

	// Get customer by verification token
	customer, err := s.customerRepo.GetByEmailVerificationToken(ctx, token)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewValidationError("Invalid or expired verification token", nil)
		}
		return err
	}

	// Check if email is already verified
	if customer.IsEmailVerified() {
		return errors.NewValidationError("Email is already verified", nil)
	}

	// Update email verification status
	if err := s.customerRepo.UpdateEmailVerification(ctx, customer.StorefrontID, customer.ID, true); err != nil {
		return errors.NewInternalError("Failed to update email verification status", err)
	}

	return nil
}

// ResendVerificationEmail resends verification email to a customer
func (s *CustomerEmailVerificationService) ResendVerificationEmail(ctx context.Context, storefrontID uuid.UUID, email string) error {
	// Get customer by email
	customer, err := s.customerRepo.GetByEmail(ctx, storefrontID, email)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return errors.NewNotFoundError("Customer not found")
		}
		return err
	}

	// Check if email is already verified
	if customer.IsEmailVerified() {
		return errors.NewValidationError("Email is already verified", nil)
	}

	// Generate new verification token
	token, err := s.generateSecureToken()
	if err != nil {
		return errors.NewInternalError("Failed to generate verification token", err)
	}

	// Set verification token in database
	if err := s.customerRepo.SetEmailVerificationToken(ctx, storefrontID, customer.ID, token); err != nil {
		return errors.NewInternalError("Failed to set verification token", err)
	}

	// Send verification email
	if err := s.sendVerificationEmail(customer, token); err != nil {
		return errors.NewInternalError("Failed to send verification email", err)
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func (s *CustomerEmailVerificationService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// sendVerificationEmail sends the verification email to the customer
func (s *CustomerEmailVerificationService) sendVerificationEmail(customer *entity.Customer, token string) error {
	if customer.Email == nil {
		return errors.NewValidationError("Customer email is not set", nil)
	}

	// Create verification URL (this should be configurable)
	verificationURL := fmt.Sprintf("https://app.smartseller.id/verify-email?token=%s", token)

	subject := "Verify Your Email Address - SmartSeller"
	
	// Create HTML email body
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Email Verification</h1>
        </div>
        <div class="content">
            <h2>Hello %s,</h2>
            <p>Thank you for registering with SmartSeller! To complete your registration, please verify your email address by clicking the button below:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Verify Email Address</a>
            </p>
            <p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
            <p><a href="%s">%s</a></p>
            <p>This verification link will expire in 24 hours for security reasons.</p>
            <p>If you didn't create an account with SmartSeller, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>Best regards,<br>The SmartSeller Team</p>
        </div>
    </div>
</body>
</html>`, customer.FirstName, verificationURL, verificationURL, verificationURL)

	// Send email using the email service
	return s.emailService.SendEmail(*customer.Email, subject, htmlBody)
}