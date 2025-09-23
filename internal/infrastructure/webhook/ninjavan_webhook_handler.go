package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// NinjaVanWebhookHandler handles webhooks from NinjaVan
type NinjaVanWebhookHandler struct {
	*BaseWebhookHandler
}

// NinjaVanWebhookPayload represents the webhook payload structure from NinjaVan
type NinjaVanWebhookPayload struct {
	TrackingID       string    `json:"tracking_id"`
	OrderID          string    `json:"order_id"`
	Status           string    `json:"status"`
	StatusCode       string    `json:"status_code"`
	StatusMessage    string    `json:"status_message"`
	SubStatus        string    `json:"sub_status"`
	Description      string    `json:"description"`
	Timestamp        time.Time `json:"timestamp"`
	UpdatedAt        time.Time `json:"updated_at"`
	Location         string    `json:"location"`
	Hub              string    `json:"hub"`
	City             string    `json:"city"`
	Country          string    `json:"country"`
	PostalCode       string    `json:"postal_code"`
	WebhookEventID   string    `json:"webhook_event_id"`
	WebhookEventType string    `json:"webhook_event_type"`
	Comments         string    `json:"comments"`
	DriverName       string    `json:"driver_name"`
	DriverPhone      string    `json:"driver_phone"`
	VehicleNumber    string    `json:"vehicle_number"`

	// Order list for handling multiple orders in one webhook
	Orders []struct {
		TrackingID string `json:"tracking_id"`
		OrderID    string `json:"order_id"`
	} `json:"orders"`

	Details struct {
		ServiceType    string  `json:"service_type"`
		ParcelWeight   float64 `json:"parcel_weight"`
		ParcelValue    float64 `json:"parcel_value"`
		CODAmount      float64 `json:"cod_amount"`
		DeliveryType   string  `json:"delivery_type"`
		DeliverySlot   string  `json:"delivery_slot"`
		Priority       string  `json:"priority"`
		SpecialRequest string  `json:"special_request"`
		POD            string  `json:"pod"` // Proof of Delivery
		PODImage       string  `json:"pod_image"`
		RecipientName  string  `json:"recipient_name"`
		RecipientPhone string  `json:"recipient_phone"`
	} `json:"details"`
}

// NewNinjaVanWebhookHandler creates a new NinjaVan webhook handler
func NewNinjaVanWebhookHandler(secretKey string) *NinjaVanWebhookHandler {
	return &NinjaVanWebhookHandler{
		BaseWebhookHandler: NewBaseWebhookHandler("ninjavan", secretKey),
	}
}

// HandleWebhook processes NinjaVan webhook payload
func (h *NinjaVanWebhookHandler) HandleWebhook(ctx context.Context, payload []byte) ([]*domainservice.TrackingUpdate, error) {
	logger.Info("processing_ninjavan_webhook",
		"Processing NinjaVan webhook payload",
		map[string]interface{}{
			"payload_size": len(payload),
		})

	var nvPayload NinjaVanWebhookPayload
	if err := json.Unmarshal(payload, &nvPayload); err != nil {
		logger.Error("failed_parse_ninjavan_webhook",
			"Failed to parse NinjaVan webhook payload",
			map[string]interface{}{
				"error":   err.Error(),
				"payload": string(payload),
			})
		return nil, fmt.Errorf("failed to parse NinjaVan webhook payload: %w", err)
	}

	logger.Debug("parsed_ninjavan_webhook",
		"Successfully parsed NinjaVan webhook payload",
		map[string]interface{}{
			"tracking_id":  nvPayload.TrackingID,
			"order_id":     nvPayload.OrderID,
			"status":       nvPayload.Status,
			"status_code":  nvPayload.StatusCode,
			"sub_status":   nvPayload.SubStatus,
			"description":  nvPayload.Description,
			"location":     nvPayload.Location,
			"event_type":   nvPayload.WebhookEventType,
			"orders_count": len(nvPayload.Orders),
		})

	var updates []*domainservice.TrackingUpdate

	// If orders array is provided, process each order
	if len(nvPayload.Orders) > 0 {
		for _, order := range nvPayload.Orders {
			trackingNumber := order.TrackingID
			if trackingNumber == "" {
				trackingNumber = order.OrderID
			}

			if trackingNumber == "" {
				logger.Warn("missing_tracking_number_in_order",
					"NinjaVan webhook order missing tracking number",
					map[string]interface{}{
						"order": order,
					})
				continue
			}

			update := h.createTrackingUpdate(nvPayload, trackingNumber)
			updates = append(updates, update)
		}
	} else {
		// Single order webhook
		trackingNumber := nvPayload.TrackingID
		if trackingNumber == "" {
			trackingNumber = nvPayload.OrderID
		}

		if trackingNumber == "" {
			logger.Error("missing_tracking_number",
				"NinjaVan webhook missing tracking number",
				map[string]interface{}{
					"payload": nvPayload,
				})
			return nil, fmt.Errorf("missing tracking number in NinjaVan webhook")
		}

		update := h.createTrackingUpdate(nvPayload, trackingNumber)
		updates = append(updates, update)
	}

	logger.Info("created_ninjavan_tracking_updates",
		"Created tracking updates from NinjaVan webhook",
		map[string]interface{}{
			"updates_count": len(updates),
		})

	return updates, nil
}

// createTrackingUpdate creates a TrackingUpdate from NinjaVan payload
func (h *NinjaVanWebhookHandler) createTrackingUpdate(nvPayload NinjaVanWebhookPayload, trackingNumber string) *domainservice.TrackingUpdate {
	// Normalize status
	normalizedStatus := h.normalizeNinjaVanStatus(nvPayload.Status, nvPayload.StatusCode)

	// Build location string
	location := h.buildLocationString(nvPayload)

	// Parse timestamp
	timestamp := nvPayload.Timestamp
	if timestamp.IsZero() && !nvPayload.UpdatedAt.IsZero() {
		timestamp = nvPayload.UpdatedAt
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Extract metadata including description
	metadata := h.extractMetadata(nvPayload)

	// Use StatusMessage if available, otherwise use description
	statusText := nvPayload.StatusMessage
	if statusText == "" {
		statusText = nvPayload.Description
	}
	if statusText == "" {
		statusText = nvPayload.Status
	}

	return &domainservice.TrackingUpdate{
		TrackingNumber: trackingNumber,
		CourierCode:    "ninjavan",
		Status:         normalizedStatus,
		StatusText:     statusText,
		Location:       location,
		Timestamp:      timestamp,
		Metadata:       metadata,
	}
}

// ValidateSignature validates the NinjaVan webhook signature
func (h *NinjaVanWebhookHandler) ValidateSignature(payload []byte, signature string) error {
	// NinjaVan uses HMAC SHA256 with their secret key
	return h.ValidateHMACSignature(payload, signature)
}

// normalizeNinjaVanStatus converts NinjaVan status codes to our standard tracking states
func (h *NinjaVanWebhookHandler) normalizeNinjaVanStatus(status, statusCode string) domainservice.TrackingState {
	// Use status code for more precise mapping if available
	if statusCode != "" {
		switch strings.ToUpper(statusCode) {
		case "PENDING", "PENDING_PICKUP", "CREATED", "BOOKED":
			return domainservice.TrackingStatePickupPending
		case "PICKED_UP", "PICKUP_DONE", "COLLECTED", "MANIFEST":
			return domainservice.TrackingStatePickedUp
		case "IN_TRANSIT", "ROUTING", "ARRIVED_AT_ORIGIN", "DEPARTED_FROM_ORIGIN",
			"ARRIVED_AT_DESTINATION", "SORTING", "ON_VEHICLE", "TRANSIT":
			return domainservice.TrackingStateInTransit
		case "OUT_FOR_DELIVERY", "ON_VEHICLE_FOR_DELIVERY", "DELIVERY_PENDING", "DELIVERING":
			return domainservice.TrackingStateOutForDelivery
		case "DELIVERED", "COMPLETED", "POD_RECEIVED", "DELIVERY_SUCCESS":
			return domainservice.TrackingStateDelivered
		case "FAILED_DELIVERY", "DELIVERY_FAIL", "RECIPIENT_NOT_AVAILABLE", "DELIVERY_FAILED":
			return domainservice.TrackingStateDeliveryFailed
		case "RETURNING", "RETURN_TO_SENDER", "RTO_PENDING":
			return domainservice.TrackingStateReturning
		case "RETURNED", "RTO_DELIVERED", "CANCELLED":
			return domainservice.TrackingStateReturned
		case "EXCEPTION", "DAMAGED", "LOST", "VOID":
			return domainservice.TrackingStateException
		}
	}

	// Fallback to status text mapping
	statusLower := strings.ToLower(strings.TrimSpace(status))
	switch {
	case strings.Contains(statusLower, "pending"), strings.Contains(statusLower, "created"),
		strings.Contains(statusLower, "booked"):
		return domainservice.TrackingStatePickupPending
	case strings.Contains(statusLower, "picked"), strings.Contains(statusLower, "collected"),
		strings.Contains(statusLower, "manifest"):
		return domainservice.TrackingStatePickedUp
	case strings.Contains(statusLower, "transit"), strings.Contains(statusLower, "routing"),
		strings.Contains(statusLower, "sorting"), strings.Contains(statusLower, "vehicle"):
		return domainservice.TrackingStateInTransit
	case strings.Contains(statusLower, "delivery") && !strings.Contains(statusLower, "delivered"):
		return domainservice.TrackingStateOutForDelivery
	case strings.Contains(statusLower, "delivered"), strings.Contains(statusLower, "completed"),
		strings.Contains(statusLower, "pod"):
		return domainservice.TrackingStateDelivered
	case strings.Contains(statusLower, "failed"), strings.Contains(statusLower, "unsuccessful"):
		return domainservice.TrackingStateDeliveryFailed
	case strings.Contains(statusLower, "returning"), strings.Contains(statusLower, "return"):
		if strings.Contains(statusLower, "returned") {
			return domainservice.TrackingStateReturned
		}
		return domainservice.TrackingStateReturning
	case strings.Contains(statusLower, "cancelled"), strings.Contains(statusLower, "void"),
		strings.Contains(statusLower, "exception"):
		return domainservice.TrackingStateException
	default:
		return domainservice.TrackingStateUnknown
	}
}

// buildLocationString constructs a location string from NinjaVan location data
func (h *NinjaVanWebhookHandler) buildLocationString(payload NinjaVanWebhookPayload) string {
	var parts []string

	if payload.Location != "" {
		parts = append(parts, payload.Location)
	}

	if payload.Hub != "" {
		parts = append(parts, payload.Hub)
	}

	if payload.City != "" {
		parts = append(parts, payload.City)
	}

	if payload.Country != "" {
		parts = append(parts, payload.Country)
	}

	if payload.PostalCode != "" {
		parts = append(parts, payload.PostalCode)
	}

	return strings.Join(parts, ", ")
}

// extractMetadata extracts metadata from NinjaVan payload
func (h *NinjaVanWebhookHandler) extractMetadata(payload NinjaVanWebhookPayload) map[string]interface{} {
	metadata := map[string]interface{}{
		"courier":            "ninjavan",
		"original_status":    payload.Status,
		"status_code":        payload.StatusCode,
		"sub_status":         payload.SubStatus,
		"webhook_event_id":   payload.WebhookEventID,
		"webhook_event_type": payload.WebhookEventType,
		"hub":                payload.Hub,
		"city":               payload.City,
		"country":            payload.Country,
		"postal_code":        payload.PostalCode,
		"driver_name":        payload.DriverName,
		"driver_phone":       payload.DriverPhone,
		"vehicle_number":     payload.VehicleNumber,
		"comments":           payload.Comments,
	}

	// Add description if available
	if payload.Description != "" {
		metadata["description"] = payload.Description
	}

	// Add delivery details if available
	if payload.Details.ServiceType != "" {
		metadata["service_type"] = payload.Details.ServiceType
		metadata["delivery_type"] = payload.Details.DeliveryType
		metadata["delivery_slot"] = payload.Details.DeliverySlot
		metadata["priority"] = payload.Details.Priority
		metadata["special_request"] = payload.Details.SpecialRequest
		metadata["pod"] = payload.Details.POD
		metadata["pod_image"] = payload.Details.PODImage
		metadata["recipient_name"] = payload.Details.RecipientName
		metadata["recipient_phone"] = payload.Details.RecipientPhone

		if payload.Details.ParcelWeight > 0 {
			metadata["parcel_weight"] = payload.Details.ParcelWeight
		}
		if payload.Details.ParcelValue > 0 {
			metadata["parcel_value"] = payload.Details.ParcelValue
		}
		if payload.Details.CODAmount > 0 {
			metadata["cod_amount"] = payload.Details.CODAmount
		}
	}

	return metadata
}
