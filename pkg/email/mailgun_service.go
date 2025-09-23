package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/pkg/logger"
	"github.com/mailgun/mailgun-go/v4"
)

// MailgunService handles sending emails using Mailgun
type MailgunService struct {
	mg          mailgun.Mailgun
	domain      string
	fromEmail   string
	fromName    string
	templates   map[string]*template.Template
	templateDir string
	mu          sync.RWMutex // Protects templates map
}

// NewMailgunService creates a new email service using Mailgun
func NewMailgunService() *MailgunService {
	// Set default template directory with project root detection
	templateDir := "templates/emails"
	execPath, err := os.Executable()
	if err == nil {
		// Try to locate templates relative to executable
		dir := filepath.Dir(execPath)
		possiblePath := filepath.Join(dir, templateDir)
		if _, err := os.Stat(possiblePath); !os.IsNotExist(err) {
			templateDir = possiblePath
		} else {
			// If not found, use working directory
			if workingDir, err := os.Getwd(); err == nil {
				templateDir = filepath.Join(workingDir, templateDir)
			}
		}
	}

	// Get config from environment variables directly
	domain := os.Getenv("MAILGUN_DOMAIN")
	apiKey := os.Getenv("MAILGUN_API_KEY")
	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "SmartSeller Team" // Default sender name
	}

	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@smartseller.com" // Default sender email
	}

	// Log configuration
	log.Printf("MAILGUN INIT - Creating service with domain=%s, from_name=%s, from_email=%s", domain, fromName, fromEmail)

	// Mask API key for secure logging
	maskedAPIKey := "not set"
	if apiKey != "" {
		if len(apiKey) > 10 {
			maskedAPIKey = apiKey[:6] + "..." + apiKey[len(apiKey)-4:]
		} else {
			maskedAPIKey = "[set but too short]"
		}
	}
	log.Printf("MAILGUN INIT - Using API Key: %s", maskedAPIKey)

	// Also log using the structured logger
	logger.Info("mailgun_service_init", map[string]interface{}{
		"domain":      domain,
		"from_email":  fromEmail,
		"from_name":   fromName,
		"api_key_set": apiKey != "",
	})

	// Create Mailgun client
	mg := mailgun.NewMailgun(domain, apiKey)

	// Set API base - comment out or adjust as needed for your region
	// mg.SetAPIBase(mailgun.APIBaseEU) // Use EU endpoint if your domain is in EU

	service := &MailgunService{
		mg:          mg,
		domain:      domain,
		fromEmail:   fromEmail,
		fromName:    fromName,
		templates:   make(map[string]*template.Template),
		templateDir: templateDir,
	}

	// Pre-load templates - reusing the same template system
	service.loadTemplates()

	return service
}

// loadTemplates loads all email templates from the template directory
// Note: This is the same as in EmailService, could be refactored to avoid duplication
func (s *MailgunService) loadTemplates() {
	// Define template functions
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("02 Jan 2006")
		},
		"year": func() int {
			return time.Now().Year()
		},
	}

	// Check if directory exists
	if _, err := os.Stat(s.templateDir); os.IsNotExist(err) {
		logger.Error("email_template_dir_not_found", "Email template directory not found", err, map[string]interface{}{
			"directory": s.templateDir,
		})
		return
	}

	// Find all HTML files in the template directory
	files, err := ioutil.ReadDir(s.templateDir)
	if err != nil {
		logger.Error("email_template_read_dir_failed", "Failed to read email template directory", err, nil)
		return
	}

	// Load each template
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".html" {
			templateName := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]
			templatePath := filepath.Join(s.templateDir, file.Name())

			// Parse the template with custom functions
			tmpl, err := template.New(file.Name()).Funcs(funcMap).ParseFiles(templatePath)
			if err != nil {
				logger.Error("email_template_parse_failed", "Failed to parse email template", err, map[string]interface{}{
					"template": templateName,
					"path":     templatePath,
				})
				continue
			}

			// Store the parsed template
			s.mu.Lock()
			s.templates[templateName] = tmpl
			s.mu.Unlock()

			logger.Info("email_template_loaded", map[string]interface{}{
				"template": templateName,
				"path":     templatePath,
			})
		}
	}
}

// getTemplate retrieves a template by name, reloading if necessary
// Note: This is the same as in EmailService, could be refactored to avoid duplication
func (s *MailgunService) getTemplate(name string) (*template.Template, error) {
	s.mu.RLock()
	tmpl, exists := s.templates[name]
	s.mu.RUnlock()

	if !exists {
		// Template not found, try to load it
		templatePath := filepath.Join(s.templateDir, name+".html")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("email template %s not found", name)
		}

		var err error
		funcMap := template.FuncMap{
			"formatDate": func(t time.Time) string {
				return t.Format("02 Jan 2006")
			},
			"year": func() int {
				return time.Now().Year()
			},
		}

		tmpl, err = template.New(name + ".html").Funcs(funcMap).ParseFiles(templatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse email template %s: %w", name, err)
		}

		// Store the template for future use
		s.mu.Lock()
		s.templates[name] = tmpl
		s.mu.Unlock()
	}

	return tmpl, nil
}

// CreateWelcomeEmail creates a welcome email
// Note: This is the same as in EmailService, could be refactored to avoid duplication
func (s *MailgunService) CreateWelcomeEmail(name string) (string, string, error) {
	// Generate tracking ID
	trackingID := uuid.New().String()

	// Set email subject
	subject := "Welcome to SmartSeller - Your E-commerce Success Starts Here!"

	// Get the welcome template
	tmpl, err := s.getTemplate("welcome")
	if err != nil {
		return "", "", err
	}

	// Prepare template data
	data := map[string]interface{}{
		"Name":       name,
		"TrackingID": trackingID,
		"UserID":     "u-" + trackingID[:8], // Shortened version as placeholder
		"Date":       time.Now().Format("2006-01-02"),
	}

	// Execute template with data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return "", "", fmt.Errorf("error executing email template: %w", err)
	}

	return subject, body.String(), nil
}

// SendWelcomeEmail sends a welcome email to a newly registered user
func (s *MailgunService) SendWelcomeEmail(name, email string) error {
	subject, body, err := s.CreateWelcomeEmail(name)
	if err != nil {
		return err
	}

	// Send the email
	return s.SendEmail(email, subject, body)
}

// SendEmail sends an email using Mailgun
func (s *MailgunService) SendEmail(to, subject, htmlBody string) error {
	// Start time for performance tracking
	startTime := time.Now()

	// Standard log for clearer console output
	log.Printf("MAILGUN EMAIL - Sending to: %s, from: %s <%s>, subject: %s",
		to, s.fromName, s.fromEmail, subject)

	// Structured logger
	logger.Info("email_send_attempt", map[string]interface{}{
		"to":       to,
		"subject":  subject,
		"from":     s.fromEmail,
		"provider": "mailgun",
		"domain":   s.domain,
	})

	// Create message
	sender := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	message := s.mg.NewMessage(sender, subject, "", to)
	message.SetHtml(htmlBody)

	// Add tracking options (Mailgun handles open and click tracking automatically)
	message.SetTracking(true)
	message.SetTrackingClicks(true)
	message.SetTrackingOpens(true)

	// Set custom variables for your tracking/analytics
	message.AddVariable("user_type", "customer")
	message.AddVariable("email_type", "welcome")

	// Set message options
	message.SetDeliveryTime(time.Now().Add(30 * time.Second)) // slight delay for better delivery

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send the message
	resp, id, err := s.mg.Send(ctx, message)

	// Log result
	elapsed := time.Since(startTime)
	if err != nil {
		logger.Error("email_send_failed", "Failed to send email via Mailgun", err, map[string]interface{}{
			"to":         to,
			"subject":    subject,
			"elapsed_ms": elapsed.Milliseconds(),
			"provider":   "mailgun",
		})
		return fmt.Errorf("failed to send email via Mailgun: %w", err)
	}

	logger.Info("email_send_success", map[string]interface{}{
		"to":         to,
		"subject":    subject,
		"elapsed_ms": elapsed.Milliseconds(),
		"provider":   "mailgun",
		"message_id": id,
		"response":   resp,
	})

	return nil
}

// SendTestEmail sends a test email
func (s *MailgunService) SendTestEmail(to, subject, htmlBody string) error {
	return s.SendEmail(to, subject, htmlBody)
}

// SendPasswordResetEmail sends a password reset email
func (s *MailgunService) SendPasswordResetEmail(email, name, resetURL string) error {
	// Set email subject
	subject := "Reset Your SmartSeller Password"

	// Try to get the password reset template
	tmpl, err := s.getTemplate("password_reset")
	if err != nil {
		// If template doesn't exist, use a simple HTML fallback
		htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">Reset Your Password</h2>
        <p>Hello %s,</p>
        <p>You have requested to reset your password for your SmartSeller account. Click the button below to reset your password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </div>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #7f8c8d;"><a href="%s">%s</a></p>
        <p><strong>This link will expire in 1 hour.</strong></p>
        <p>If you didn't request this password reset, please ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="font-size: 12px; color: #7f8c8d;">
            Best regards,<br>
            The SmartSeller Team
        </p>
    </div>
</body>
</html>`, name, resetURL, resetURL, resetURL)

		return s.SendEmail(email, subject, htmlBody)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Name":     name,
		"ResetURL": resetURL,
		"Date":     time.Now(),
	}

	// Execute template with data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("error executing password reset email template: %w", err)
	}

	return s.sendPasswordResetEmailWithoutTracking(email, subject, body.String())
}

// sendPasswordResetEmailWithoutTracking sends a password reset email without click tracking to prevent URL wrapping
func (s *MailgunService) sendPasswordResetEmailWithoutTracking(to, subject, htmlBody string) error {
	// Start time for performance tracking
	startTime := time.Now()

	// Standard log for clearer console output
	log.Printf("MAILGUN EMAIL - Sending password reset to: %s, from: %s <%s>, subject: %s",
		to, s.fromName, s.fromEmail, subject)

	// Structured logger
	logger.Info("email_send_attempt", map[string]interface{}{
		"to":       to,
		"subject":  subject,
		"from":     s.fromEmail,
		"provider": "mailgun",
		"domain":   s.domain,
		"type":     "password_reset",
	})

	// Create message
	sender := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	message := s.mg.NewMessage(sender, subject, "", to)
	message.SetHtml(htmlBody)

	// Disable click tracking for password reset emails to prevent URL wrapping
	message.SetTracking(true)
	message.SetTrackingClicks(false) // This prevents URL wrapping
	message.SetTrackingOpens(true)   // Keep open tracking for analytics

	// Send message
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	resp, id, err := s.mg.Send(ctx, message)
	if err != nil {
		// Structured error logger
		logger.Error("email_send_failed", map[string]interface{}{
			"error":    err.Error(),
			"to":       to,
			"subject":  subject,
			"provider": "mailgun",
		})
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	// Standard log for clearer console output
	log.Printf("MAILGUN SUCCESS - Message ID: %s, Response: %s, Time: %v", id, resp, elapsed)

	// Structured success logger
	logger.Info("email_send_success", map[string]interface{}{
		"message_id": id,
		"response":   resp,
		"elapsed_ms": elapsed.Milliseconds(),
		"to":         to,
		"subject":    subject,
		"provider":   "mailgun",
		"type":       "password_reset",
	})

	return nil
}
