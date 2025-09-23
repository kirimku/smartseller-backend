package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// TelegramService handles sending messages to Telegram
type TelegramService struct {
	botToken string
	chatIDs  []string
	timeout  time.Duration
	client   *http.Client
}

// NewTelegramService creates a new TelegramService
func NewTelegramService(botToken string, chatIDs []string, timeout time.Duration) *TelegramService {
	return &TelegramService{
		botToken: botToken,
		chatIDs:  chatIDs,
		timeout:  timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SendAlert sends an alert message to all configured Telegram chat IDs
func (t *TelegramService) SendAlert(message string) error {
	if t.botToken == "" || len(t.chatIDs) == 0 {
		logger.Warn("telegram_not_configured", "Telegram bot token or chat IDs not configured")
		return nil // Don't fail if not configured
	}

	for _, chatID := range t.chatIDs {
		if err := t.sendMessage(chatID, message); err != nil {
			logger.Error("telegram_send_failed", "Failed to send Telegram message", map[string]interface{}{
				"chat_id": chatID,
				"error":   err.Error(),
			})
			// Continue with other chat IDs even if one fails
		}
	}

	return nil
}

// sendMessage sends a message to a specific chat ID
func (t *TelegramService) sendMessage(chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
		// Temporarily remove parse_mode to avoid Markdown parsing errors
		// "parse_mode": "Markdown",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for better error debugging
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, bodyStr)
	}

	return nil
}

// TestConnection tests the connection to Telegram API
func (t *TelegramService) TestConnection() error {
	if t.botToken == "" {
		return fmt.Errorf("bot token not configured")
	}

	testMessage := "🧪 Telegram Alert Test\n\nThis is a test message to verify Telegram integration is working correctly."
	return t.SendAlert(testMessage)
}
