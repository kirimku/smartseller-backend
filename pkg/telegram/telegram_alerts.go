package telegram

import (
	"sync"

	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

var (
	alertManager     *TelegramAlertManager
	alertManagerOnce sync.Once
)

// GetAlertManager returns the singleton TelegramAlertManager instance
func GetAlertManager() *TelegramAlertManager {
	alertManagerOnce.Do(func() {
		telegramConfig := config.AppConfig.Telegram
		if telegramConfig.Enabled {
			service := NewTelegramService(
				telegramConfig.BotToken,
				telegramConfig.ChatIDs,
				telegramConfig.Timeout,
			)
			alertManager = NewTelegramAlertManager(service, AlertLevel(telegramConfig.AlertLevel))
		} else {
			// Create a no-op manager if Telegram is disabled
			alertManager = &TelegramAlertManager{
				service: &TelegramService{}, // Empty service that does nothing
			}
		}
	})
	return alertManager
}

// AlertSystemError sends a system error alert
func AlertSystemError(service, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendSystemError(service, message, details)
}

// AlertDatabaseError sends a database error alert
func AlertDatabaseError(operation, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendDatabaseError(operation, message, details)
}

// AlertPaymentError sends a payment error alert
func AlertPaymentError(provider, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendPaymentError(provider, message, details)
}

// AlertCourierError sends a courier error alert
func AlertCourierError(courier, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendCourierError(courier, message, details)
}

// AlertPerformanceAlert sends a performance alert
func AlertPerformanceAlert(metric, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendPerformanceAlert(metric, message, details)
}

// AlertBusinessAlert sends a business logic alert
func AlertBusinessAlert(category, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendBusinessAlert(category, message, details)
}

// AlertSecurityAlert sends a security alert
func AlertSecurityAlert(event, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendSecurityAlert(event, message, details)
}

// AlertCustomAlert sends a custom alert with specified parameters
func AlertCustomAlert(title, emoji string, level AlertLevel, category, message string, details map[string]interface{}) {
	manager := GetAlertManager()
	manager.SendCustomAlert(title, emoji, level, category, message, details)
}

// TestTelegramConnection tests the Telegram connection
func TestTelegramConnection() error {
	manager := GetAlertManager()
	if manager.service.botToken == "" {
		logger.Warn("telegram_test_skipped", "Telegram bot token not configured, skipping test")
		return nil
	}

	return manager.service.TestConnection()
}
