package telegram

import (
	"fmt"
	"time"

	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// AlertLevel represents the severity level of an alert
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// TelegramAlertManager manages different types of alerts
type TelegramAlertManager struct {
	service    *TelegramService
	alertLevel AlertLevel
}

// NewTelegramAlertManager creates a new TelegramAlertManager
func NewTelegramAlertManager(service *TelegramService, alertLevel AlertLevel) *TelegramAlertManager {
	return &TelegramAlertManager{
		service:    service,
		alertLevel: alertLevel,
	}
}

// SendSystemError sends a system error alert
func (m *TelegramAlertManager) SendSystemError(service, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage("🚨 System Error", AlertLevelError, service, message, details)
	m.sendAlert(alertMessage, AlertLevelError)
}

// SendDatabaseError sends a database error alert
func (m *TelegramAlertManager) SendDatabaseError(operation, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage("🗄️ Database Error", AlertLevelCritical, operation, message, details)
	m.sendAlert(alertMessage, AlertLevelCritical)
}

// SendPaymentError sends a payment error alert
func (m *TelegramAlertManager) SendPaymentError(provider, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage("💳 Payment Error", AlertLevelError, provider, message, details)
	m.sendAlert(alertMessage, AlertLevelError)
}

// SendCourierError sends a courier error alert
func (m *TelegramAlertManager) SendCourierError(courier, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage("🚚 Courier Error", AlertLevelWarning, courier, message, details)
	m.sendAlert(alertMessage, AlertLevelWarning)
}

// SendPerformanceAlert sends a performance alert
func (m *TelegramAlertManager) SendPerformanceAlert(metric, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage("⚡ Performance Alert", AlertLevelWarning, metric, message, details)
	m.sendAlert(alertMessage, AlertLevelWarning)
}

// SendBusinessAlert sends a business logic alert
func (m *TelegramAlertManager) SendBusinessAlert(category, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage("📊 Business Alert", AlertLevelInfo, category, message, details)
	m.sendAlert(alertMessage, AlertLevelInfo)
}

// SendSecurityAlert sends a security alert
func (m *TelegramAlertManager) SendSecurityAlert(event, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage("🔒 Security Alert", AlertLevelCritical, event, message, details)
	m.sendAlert(alertMessage, AlertLevelCritical)
}

// SendCustomAlert sends a custom alert with specified level and emoji
func (m *TelegramAlertManager) SendCustomAlert(title, emoji string, level AlertLevel, category, message string, details map[string]interface{}) {
	alertMessage := m.formatAlertMessage(fmt.Sprintf("%s %s", emoji, title), level, category, message, details)
	m.sendAlert(alertMessage, level)
}

// formatAlertMessage formats an alert message with consistent structure
func (m *TelegramAlertManager) formatAlertMessage(title string, level AlertLevel, category, message string, details map[string]interface{}) string {
	alertMessage := fmt.Sprintf("%s\n", title)
	alertMessage += fmt.Sprintf("Level: %s\n", string(level))
	alertMessage += fmt.Sprintf("Category: %s\n", category)
	alertMessage += fmt.Sprintf("Message: %s\n", message)

	if len(details) > 0 {
		alertMessage += "\nDetails:\n"
		for key, value := range details {
			alertMessage += fmt.Sprintf("• %s: %v\n", key, value)
		}
	}

	alertMessage += fmt.Sprintf("\nTime: %s", time.Now().Format("2006-01-02 15:04:05"))
	return alertMessage
}

// sendAlert sends the alert through the service if the level meets the threshold
func (m *TelegramAlertManager) sendAlert(message string, level AlertLevel) {
	// Check if this alert level should be sent based on configured level
	if !m.shouldSendAlert(level) {
		return // Skip sending this alert
	}

	// Send alert asynchronously to avoid blocking the main thread
	go func() {
		if err := m.service.SendAlert(message); err != nil {
			logger.Error("telegram_alert_failed", "Failed to send Telegram alert", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()
}

// shouldSendAlert determines if an alert should be sent based on the configured alert level
func (m *TelegramAlertManager) shouldSendAlert(alertLevel AlertLevel) bool {
	// Define level hierarchy (higher number = higher priority)
	levelPriority := map[AlertLevel]int{
		AlertLevelInfo:     1,
		AlertLevelWarning:  2,
		AlertLevelError:    3,
		AlertLevelCritical: 4,
	}

	alertPriority, alertExists := levelPriority[alertLevel]
	configPriority, configExists := levelPriority[m.alertLevel]

	// If either level doesn't exist in our mapping, default to sending
	if !alertExists || !configExists {
		return true
	}

	// Send alert if its priority is >= configured level priority
	return alertPriority >= configPriority
}
