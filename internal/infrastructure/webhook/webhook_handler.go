package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// TrackingUpdate represents a tracking update (temporary placeholder)
type TrackingUpdate struct {
	TrackingNumber string
	Status         string
	Timestamp      time.Time
}

// WebhookHandler defines the interface for handling courier webhooks
type WebhookHandler interface {
	// HandleWebhook processes the webhook payload and returns tracking updates
	HandleWebhook(ctx context.Context, payload []byte) ([]*TrackingUpdate, error)

	// ValidateSignature validates the webhook signature
	ValidateSignature(payload []byte, signature string) error

	// GetCourierCode returns the courier code this handler supports
	GetCourierCode() string
}

// BaseWebhookHandler provides common functionality for webhook handlers
type BaseWebhookHandler struct {
	courierCode string
	secretKey   string
}

// NewBaseWebhookHandler creates a new base webhook handler
func NewBaseWebhookHandler(courierCode, secretKey string) *BaseWebhookHandler {
	return &BaseWebhookHandler{
		courierCode: courierCode,
		secretKey:   secretKey,
	}
}

// GetCourierCode returns the courier code
func (h *BaseWebhookHandler) GetCourierCode() string {
	return h.courierCode
}

// ValidateHMACSignature validates HMAC SHA256 signature (common pattern)
func (h *BaseWebhookHandler) ValidateHMACSignature(payload []byte, signature string) error {
	if h.secretKey == "" {
		// If no secret key configured, skip validation (for testing)
		return nil
	}

	// Create HMAC SHA256 hash
	mac := hmac.New(sha256.New, []byte(h.secretKey))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// Common webhook payload structures that most couriers use
type CommonWebhookPayload struct {
	TrackingNumber string                 `json:"tracking_number"`
	AWB            string                 `json:"awb"`
	Status         string                 `json:"status"`
	StatusText     string                 `json:"status_text"`
	Description    string                 `json:"description"`
	Location       string                 `json:"location"`
	City           string                 `json:"city"`
	Timestamp      time.Time              `json:"timestamp"`
	EventTime      time.Time              `json:"event_time"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// normalizeTimestamp tries to parse different timestamp formats
func normalizeTimestamp(timeStr string) time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05+07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	// If no format matches, return current time
	return time.Now()
}
