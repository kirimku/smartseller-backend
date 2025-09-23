package email

import (
	"log"
)

// NewEmailService returns the appropriate email service based on configuration
func NewEmailService() EmailSender {
	// Check if Mailgun is configured
	mailgunDomain := GetEnvWithDefault("MAILGUN_DOMAIN", "")
	mailgunAPIKey := GetEnvWithDefault("MAILGUN_API_KEY", "")

	if mailgunDomain != "" && mailgunAPIKey != "" {
		log.Printf("EMAIL SERVICE - Using Mailgun with domain: %s", mailgunDomain)
		
		// Mask API key for secure logging
		maskedAPIKey := "not set"
		if mailgunAPIKey != "" {
			if len(mailgunAPIKey) > 10 {
				maskedAPIKey = mailgunAPIKey[:6] + "..." + mailgunAPIKey[len(mailgunAPIKey)-4:]
			} else {
				maskedAPIKey = "[set but too short]"
			}
		}
		log.Printf("EMAIL SERVICE - Mailgun API Key: %s", maskedAPIKey)
		
		return NewMailgunService()
	}

	// Fallback to a mock service if in development or testing
	log.Printf("EMAIL SERVICE - WARNING: No email service configured (domain=%s, api_key=%v)", 
		mailgunDomain, mailgunAPIKey != "")
	log.Printf("EMAIL SERVICE - Using mock email service that logs emails but doesn't send them")
	return NewMockEmailService()
}
