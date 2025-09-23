package email

// EmailSender defines the interface for any email service
type EmailSender interface {
	SendEmail(to, subject, htmlBody string) error
	SendWelcomeEmail(name, email string) error
	SendPasswordResetEmail(email, name, resetURL string) error
	SendTestEmail(to, subject, htmlBody string) error
}

// Ensure our implementation satisfies the interface
var _ EmailSender = (*MailgunService)(nil)
