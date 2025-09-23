package email

import (
	"fmt"
	"log"
)

// MockEmailService is a mock implementation that just logs emails
type MockEmailService struct{}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

// SendEmail logs the email details but doesn't actually send
func (s *MockEmailService) SendEmail(to, subject, htmlBody string) error {
	log.Printf("MOCK EMAIL: To: %s, Subject: %s", to, subject)
	log.Printf("MOCK EMAIL BODY: %s", htmlBody)
	return nil
}

// SendWelcomeEmail creates and logs a welcome email
func (s *MockEmailService) SendWelcomeEmail(name, email string) error {
	subject := "Welcome to SmartSeller"
	body := fmt.Sprintf("<h1>Welcome %s!</h1><p>This is a mock welcome email.</p>", name)

	// Use the body variable to avoid the "declared and not used" error
	log.Printf("MOCK WELCOME EMAIL: To: %s, Subject: %s", email, subject)
	log.Printf("MOCK WELCOME EMAIL BODY: %s", body)

	return nil
}

// SendTestEmail logs a test email
func (s *MockEmailService) SendTestEmail(to, subject, htmlBody string) error {
	log.Printf("MOCK TEST EMAIL: To: %s, Subject: %s", to, subject)
	log.Printf("MOCK TEST EMAIL BODY: %s", htmlBody)
	return nil
}

// SendPasswordResetEmail logs a password reset email
func (s *MockEmailService) SendPasswordResetEmail(email, name, resetURL string) error {
	subject := "Reset Your SmartSeller Password"
	body := fmt.Sprintf("<h1>Reset Password</h1><p>Hello %s, click this link to reset your password: %s</p>", name, resetURL)

	log.Printf("MOCK PASSWORD RESET EMAIL: To: %s, Subject: %s", email, subject)
	log.Printf("MOCK PASSWORD RESET EMAIL BODY: %s", body)
	log.Printf("MOCK PASSWORD RESET URL: %s", resetURL)

	return nil
}
