package email

import "os"

// GetEnvWithDefault gets an environment variable or returns the default if not set
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEmailConfig returns a standardized set of email configuration values
func GetEmailConfig() (fromName, fromEmail string) {
	fromName = GetEnvWithDefault("SMTP_FROM_NAME", "SmartSeller Team")
	fromEmail = GetEnvWithDefault("SMTP_FROM_EMAIL", "noreply@smartseller.com")
	return fromName, fromEmail
}
