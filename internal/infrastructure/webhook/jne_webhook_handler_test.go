package webhook

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	// domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJNEWebhookHandler_isDeliveredStatus(t *testing.T) {
	handler := NewJNEWebhookHandler("test-secret")

	tests := []struct {
		name       string
		statusCode string
		expected   bool
	}{
		{"D01 delivered", "D01", true},
		{"D02 delivered", "D02", true},
		{"D03 delivered", "D03", true},
		{"D04 delivered", "D04", true},
		{"D05 delivered", "D05", true},
		{"D06 delivered", "D06", true},
		{"D07 delivered", "D07", true},
		{"D08 delivered", "D08", true},
		{"D09 delivered", "D09", true},
		{"D10 delivered", "D10", true},
		{"D11 delivered", "D11", true},
		{"D12 delivered", "D12", true},
		{"DB1 delivered", "DB1", true},
		{"lowercase d01", "d01", true},
		{"mixed case Db1", "Db1", true},
		{"not delivered T01", "T01", false},
		{"not delivered M01", "M01", false},
		{"empty status", "", false},
		{"invalid status", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isDeliveredStatus(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJNEWebhookHandler_normalizeJNEStatus(t *testing.T) {
	handler := NewJNEWebhookHandler("test-secret")

	tests := []struct {
		name        string
		status      string
		statusCode  string
		expected    domainservice.TrackingState
		description string
	}{
		{
			name:        "D01 delivered status",
			status:      "Delivered",
			statusCode:  "D01",
			expected:    domainservice.TrackingStateDelivered,
			description: "Should recognize D01 as delivered",
		},
		{
			name:        "DB1 delivered status",
			status:      "Package delivered",
			statusCode:  "DB1",
			expected:    domainservice.TrackingStateDelivered,
			description: "Should recognize DB1 as delivered",
		},
		{
			name:        "D12 delivered status",
			status:      "Package delivered successfully",
			statusCode:  "D12",
			expected:    domainservice.TrackingStateDelivered,
			description: "Should recognize D12 as delivered",
		},
		{
			name:        "M01 manifest status",
			status:      "Package manifested",
			statusCode:  "M01",
			expected:    domainservice.TrackingStatePickedUp,
			description: "Should recognize M01 as picked up",
		},
		{
			name:        "T01 transit status",
			status:      "In transit",
			statusCode:  "T01",
			expected:    domainservice.TrackingStateInTransit,
			description: "Should recognize T01 as in transit",
		},
		{
			name:        "O01 out for delivery",
			status:      "Out for delivery",
			statusCode:  "O01",
			expected:    domainservice.TrackingStateOutForDelivery,
			description: "Should recognize O01 as out for delivery",
		},
		{
			name:        "F01 delivery failed",
			status:      "Delivery failed",
			statusCode:  "F01",
			expected:    domainservice.TrackingStateDeliveryFailed,
			description: "Should recognize F01 as delivery failed",
		},
		{
			name:        "R01 returning",
			status:      "Returning to sender",
			statusCode:  "R01",
			expected:    domainservice.TrackingStateReturning,
			description: "Should recognize R01 as returning",
		},
		{
			name:        "R03 returned",
			status:      "Returned to sender",
			statusCode:  "R03",
			expected:    domainservice.TrackingStateReturned,
			description: "Should recognize R03 as returned",
		},
		{
			name:        "C01 cancelled",
			status:      "Package cancelled",
			statusCode:  "C01",
			expected:    domainservice.TrackingStateException,
			description: "Should recognize C01 as exception",
		},
		{
			name:        "B01 booking",
			status:      "Package booked",
			statusCode:  "B01",
			expected:    domainservice.TrackingStatePickupPending,
			description: "Should recognize B01 as pickup pending",
		},
		{
			name:        "Indonesian delivered text without code",
			status:      "Paket telah diterima",
			statusCode:  "",
			expected:    domainservice.TrackingStateDelivered,
			description: "Should recognize Indonesian delivered text",
		},
		{
			name:        "Indonesian transit text without code",
			status:      "Paket dalam perjalanan",
			statusCode:  "",
			expected:    domainservice.TrackingStateInTransit,
			description: "Should recognize Indonesian transit text",
		},
		{
			name:        "Indonesian pickup text without code",
			status:      "Paket telah diambil",
			statusCode:  "",
			expected:    domainservice.TrackingStatePickedUp,
			description: "Should recognize Indonesian pickup text",
		},
		{
			name:        "Indonesian delivery text without code",
			status:      "Sedang dalam pengiriman",
			statusCode:  "",
			expected:    domainservice.TrackingStateOutForDelivery,
			description: "Should recognize Indonesian delivery text",
		},
		{
			name:        "Indonesian failed text without code",
			status:      "Pengiriman gagal",
			statusCode:  "",
			expected:    domainservice.TrackingStateDeliveryFailed,
			description: "Should recognize Indonesian failed text",
		},
		{
			name:        "Indonesian cancelled text without code",
			status:      "Paket dibatalkan",
			statusCode:  "",
			expected:    domainservice.TrackingStateException,
			description: "Should recognize Indonesian cancelled text",
		},
		{
			name:        "Indonesian returned text without code",
			status:      "Paket dikembalikan",
			statusCode:  "",
			expected:    domainservice.TrackingStateReturned,
			description: "Should recognize Indonesian returned text",
		},
		{
			name:        "Unknown status",
			status:      "Some unknown status",
			statusCode:  "UNKNOWN",
			expected:    domainservice.TrackingStateUnknown,
			description: "Should return unknown for unrecognized status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.normalizeJNEStatus(tt.status, tt.statusCode)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestJNEWebhookHandler_HandleWebhook_Success(t *testing.T) {
	handler := NewJNEWebhookHandler("test-secret")

	// Create test payload
	payload := JNEWebhookPayload{
		CNoteNo:     "JNE123456789",
		AWB:         "JNE123456789",
		Status:      "Package delivered",
		StatusCode:  "D01",
		StatusDesc:  "Delivered successfully",
		Description: "Package has been delivered to recipient",
		Location:    "Jakarta Selatan",
		City:        "Jakarta",
		Office:      "JNE Jakarta Selatan",
		Timestamp:   time.Now(),
		EventDate:   "2024-01-20",
		EventTime:   "15:30:00",
		WebhookID:   "webhook-123",
		Reference:   "ref-456",
		Details: struct {
			Weight       float64 `json:"weight"`
			ServiceType  string  `json:"service_type"`
			ServiceName  string  `json:"service_name"`
			ReceiverName string  `json:"receiver_name"`
			SenderName   string  `json:"sender_name"`
			POD          string  `json:"pod"`
			PODPhoto     string  `json:"pod_photo"`
			DeliveryDate string  `json:"delivery_date"`
			DeliveryTime string  `json:"delivery_time"`
		}{
			Weight:       1.5,
			ServiceType:  "REG",
			ServiceName:  "JNE REG",
			ReceiverName: "John Doe",
			SenderName:   "Jane Smith",
			POD:          "John Doe",
			PODPhoto:     "https://example.com/pod.jpg",
			DeliveryDate: "2024-01-20",
			DeliveryTime: "15:30:00",
		},
	}

	// Marshal to JSON
	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	// Handle webhook
	ctx := context.Background()
	updates, err := handler.HandleWebhook(ctx, jsonPayload)

	// Assertions
	require.NoError(t, err)
	require.Len(t, updates, 1)

	update := updates[0]
	assert.Equal(t, "JNE123456789", update.TrackingNumber)
	assert.Equal(t, "jne", update.CourierCode)
	assert.Equal(t, domainservice.TrackingStateDelivered, update.Status)
	assert.Equal(t, "Delivered successfully", update.StatusText)
	assert.Equal(t, "Jakarta Selatan, Jakarta, (JNE Jakarta Selatan)", update.Location)

	// Check metadata
	assert.Equal(t, "webhook-123", update.Metadata["webhook_id"])
	assert.Equal(t, "ref-456", update.Metadata["reference"])
	assert.Equal(t, "D01", update.Metadata["status_code"])
	assert.Equal(t, 1.5, update.Metadata["weight"])
	assert.Equal(t, "REG", update.Metadata["service_type"])
	assert.Equal(t, "JNE REG", update.Metadata["service_name"])
	assert.Equal(t, "John Doe", update.Metadata["receiver_name"])
	assert.Equal(t, "Jane Smith", update.Metadata["sender_name"])
	assert.Equal(t, "John Doe", update.Metadata["pod"])
	assert.Equal(t, "jne_webhook", update.Metadata["source"])
}

func TestJNEWebhookHandler_HandleWebhook_WithDifferentDeliveredCodes(t *testing.T) {
	handler := NewJNEWebhookHandler("test-secret")

	deliveredCodes := []string{"D01", "D02", "D03", "D04", "D05", "D06", "D07", "D08", "D09", "D10", "D11", "D12", "DB1"}

	for _, code := range deliveredCodes {
		t.Run("Delivered code "+code, func(t *testing.T) {
			payload := JNEWebhookPayload{
				CNoteNo:     "JNE123456789",
				AWB:         "JNE123456789",
				Status:      "Package delivered",
				StatusCode:  code,
				StatusDesc:  "Delivered with code " + code,
				Description: "Package has been delivered",
				Location:    "Jakarta",
				EventDate:   "2024-01-20",
				EventTime:   "15:30:00",
			}

			jsonPayload, err := json.Marshal(payload)
			require.NoError(t, err)

			ctx := context.Background()
			updates, err := handler.HandleWebhook(ctx, jsonPayload)

			require.NoError(t, err)
			require.Len(t, updates, 1)

			update := updates[0]
			assert.Equal(t, domainservice.TrackingStateDelivered, update.Status, "Code %s should be recognized as delivered", code)
		})
	}
}

func TestJNEWebhookHandler_HandleWebhook_MissingTrackingNumber(t *testing.T) {
	handler := NewJNEWebhookHandler("test-secret")

	// Create payload without tracking number
	payload := JNEWebhookPayload{
		Status:      "Package delivered",
		StatusCode:  "D01",
		StatusDesc:  "Delivered successfully",
		Description: "Package has been delivered to recipient",
		Location:    "Jakarta",
		EventDate:   "2024-01-20",
		EventTime:   "15:30:00",
	}

	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	ctx := context.Background()
	updates, err := handler.HandleWebhook(ctx, jsonPayload)

	// Should return error for missing tracking number
	assert.Error(t, err)
	assert.Nil(t, updates)
	assert.Contains(t, err.Error(), "missing tracking number")
}

func TestJNEWebhookHandler_HandleWebhook_InvalidJSON(t *testing.T) {
	handler := NewJNEWebhookHandler("test-secret")

	invalidJSON := []byte(`{"invalid": json}`)

	ctx := context.Background()
	updates, err := handler.HandleWebhook(ctx, invalidJSON)

	assert.Error(t, err)
	assert.Nil(t, updates)
	assert.Contains(t, err.Error(), "failed to parse JNE webhook payload")
}

func TestJNEWebhookHandler_buildLocation(t *testing.T) {
	handler := NewJNEWebhookHandler("test-secret")

	tests := []struct {
		name     string
		location string
		city     string
		office   string
		expected string
	}{
		{
			name:     "All fields present",
			location: "Jakarta Selatan",
			city:     "Jakarta",
			office:   "JNE Jakarta Selatan",
			expected: "Jakarta Selatan, Jakarta, (JNE Jakarta Selatan)",
		},
		{
			name:     "Location and city same",
			location: "Jakarta",
			city:     "Jakarta",
			office:   "JNE Jakarta",
			expected: "Jakarta, (JNE Jakarta)",
		},
		{
			name:     "Location and office same",
			location: "Jakarta Hub",
			city:     "Jakarta",
			office:   "Jakarta Hub",
			expected: "Jakarta Hub, Jakarta",
		},
		{
			name:     "Only location",
			location: "Jakarta",
			city:     "",
			office:   "",
			expected: "Jakarta",
		},
		{
			name:     "Only city",
			location: "",
			city:     "Jakarta",
			office:   "",
			expected: "Jakarta",
		},
		{
			name:     "Location and office, no city",
			location: "Jakarta Hub",
			city:     "",
			office:   "JNE Jakarta",
			expected: "Jakarta Hub, (JNE Jakarta)",
		},
		{
			name:     "Empty fields",
			location: "",
			city:     "",
			office:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.buildLocation(tt.location, tt.city, tt.office)
			assert.Equal(t, tt.expected, result)
		})
	}
}
