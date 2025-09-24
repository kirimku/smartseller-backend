package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	// domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// SiCepatWebhookHandler handles webhooks from SiCepat
type SiCepatWebhookHandler struct {
	*BaseWebhookHandler
}

// History represents a tracking history entry (matching Ruby reference exactly)
type History struct {
	Position string `json:"position"`
	Status   string `json:"status"`
	Note     string `json:"note"`
	Time     string `json:"time"`
}

// AdditionalData represents extra data from Shipping struct
type AdditionalData struct {
	Flag string `json:"flag"`
}

// CourierDriver represents driver information
type CourierDriver struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// Shipping represents the webhook payload structure exactly matching Ruby reference
type Shipping struct {
	AirwaybillNumber  string         `json:"airwaybill_number"`
	CourierName       string         `json:"courier_name"`
	CourierService    string         `json:"courier_service"`
	CourierDriver     CourierDriver  `json:"courier_driver"`
	ActualShippingFee int            `json:"actual_shipping_fee"`
	ActualWeight      int            `json:"actual_weight"`
	ShipmentDate      string         `json:"shipment_date"`
	ShipperName       string         `json:"shipper_name"`
	ShipperAddress    string         `json:"shipper_address"`
	ReceiverName      string         `json:"receiver_name"`
	ReceiverAddress   string         `json:"receiver_address"`
	SummaryStatus     string         `json:"summary_status"`
	LastStatus        string         `json:"last_status"`
	LastUpdateAt      string         `json:"last_update_at"`
	ShipmentHistories []History      `json:"shipment_histories"`
	AdditionalData    AdditionalData `json:"additional_data"`
	Note              string         `json:"note"`
	ErrorTxt          string         `json:"error_txt"`
	ResiStatus        int            `json:"resi_status"`
	Version           string         `json:"version"`
	TimeReceived      float64        `json:"time_received"` // in Unix seconds
	LatencyCount      float64        `json:"latency_count"` // in seconds
	SentAt            string         `json:"sent_at"`
}

// HTTPResponse represents the response structure for webhooks
type HTTPResponse struct {
	Success      bool   `json:"Success"`
	ErrorMessage string `json:"ErrorMessage"`
}

// NewSiCepatWebhookHandler creates a new SiCepat webhook handler
func NewSiCepatWebhookHandler(secretKey string) *SiCepatWebhookHandler {
	return &SiCepatWebhookHandler{
		BaseWebhookHandler: NewBaseWebhookHandler("sicepat", secretKey),
	}
}

// CountLatency calculates latency from partner to replayer (matching Ruby reference exactly)
func (h *SiCepatWebhookHandler) CountLatency(timeUpdated string) float64 {
	tr, err := time.Parse("2006-01-02T15:04:05-07:00", timeUpdated)
	if err != nil {
		logger.Error("sicepat_invalid_timestamp", "Date is invalid", map[string]interface{}{"error": err.Error(), "time": timeUpdated})
		return -1
	}
	return time.Now().Sub(tr).Seconds()
}

// SetLatency sets latency from partner to replayer (matching Ruby reference exactly)
func (s *Shipping) SetLatency(h *SiCepatWebhookHandler, timeUpdated string) {
	s.LatencyCount = h.CountLatency(timeUpdated)
	s.TimeReceived = float64(time.Now().UnixNano()) / 1000000000
}

// publish3plToInLatency publishes latency metrics from 3PL to our service
func (h *SiCepatWebhookHandler) publish3plToInLatency(lastUpdateAt string) bool {
	lastUpdateTime, timeErr := time.Parse("2006-01-02T15:04:05-07:00", lastUpdateAt)
	if timeErr == nil {
		// In production, this would send to DataDog or similar monitoring system
		latency := float64(time.Now().Unix() - lastUpdateTime.Unix())
		logger.Info("sicepat_3pl_to_in_latency",
			"SiCepat 3PL to In latency",
			map[string]interface{}{
				"latency_seconds": latency,
				"courier":         "sicepat",
			})
		return true
	}
	return false
}

// publishInToOutLatency publishes internal processing latency metrics
func (h *SiCepatWebhookHandler) publishInToOutLatency(timeReceived float64) bool {
	if timeReceived > 0 {
		elapse := float64(time.Now().Unix()) - timeReceived
		logger.Info("sicepat_in_to_out_latency",
			"SiCepat In to Out latency",
			map[string]interface{}{
				"latency_seconds": elapse,
				"courier":         "sicepat",
			})
		return true
	}
	return false
}

// HandleWebhook processes SiCepat webhook payload following Ruby HandleRequest logic exactly
func (h *SiCepatWebhookHandler) HandleWebhook(ctx context.Context, payload []byte) ([]*domainservice.TrackingUpdate, error) {
	logger.Info("processing_sicepat_webhook",
		"Processing SiCepat webhook payload",
		map[string]interface{}{
			"payload_size": len(payload),
		})

	var shipping Shipping
	if err := json.Unmarshal(payload, &shipping); err != nil {
		logger.Error("sicepat_webhook_decode_error", "Error decode payload", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("failed to parse SiCepat webhook payload: %w", err)
	}

	logger.Info("sicepat_webhook_payload", "SiCepat Payload received", map[string]interface{}{"payload": shipping})

	// Set latency following Ruby HandleRequest logic exactly
	histories := shipping.ShipmentHistories
	if histories != nil && len(histories) > 0 {
		shipping.SetLatency(h, histories[len(histories)-1].Time)
	}
	h.publish3plToInLatency(shipping.LastUpdateAt)

	// Validate required fields
	trackingNumber := shipping.AirwaybillNumber
	if trackingNumber == "" {
		return nil, fmt.Errorf("missing airwaybill_number in SiCepat webhook")
	}

	// Parse event timestamp from last update
	var eventTime time.Time
	if shipping.LastUpdateAt != "" {
		if parsedTime, err := time.Parse("2006-01-02T15:04:05-07:00", shipping.LastUpdateAt); err == nil {
			eventTime = parsedTime
		} else {
			eventTime = time.Now()
		}
	} else {
		eventTime = time.Now()
	}

	// Normalize status based on summary and last status
	normalizedStatus := h.normalizeSiCepatStatus(shipping.SummaryStatus, shipping.LastStatus)

	// Get status text - use last status if available
	statusText := shipping.LastStatus
	if statusText == "" {
		statusText = shipping.SummaryStatus
	}

	// Build location from shipping histories if available
	location := ""
	if len(shipping.ShipmentHistories) > 0 {
		lastHistory := shipping.ShipmentHistories[len(shipping.ShipmentHistories)-1]
		location = lastHistory.Position
		if lastHistory.Note != "" {
			location = fmt.Sprintf("%s - %s", location, lastHistory.Note)
		}
	}

	// Create tracking update with Ruby reference metadata structure
	update := &domainservice.TrackingUpdate{
		TrackingNumber: trackingNumber,
		CourierCode:    "sicepat",
		Status:         normalizedStatus,
		StatusText:     statusText,
		Location:       location,
		Timestamp:      eventTime,
		Metadata: map[string]interface{}{
			"courier_name":        shipping.CourierName,
			"courier_service":     shipping.CourierService,
			"actual_shipping_fee": shipping.ActualShippingFee,
			"actual_weight":       shipping.ActualWeight,
			"shipment_date":       shipping.ShipmentDate,
			"shipper_name":        shipping.ShipperName,
			"shipper_address":     shipping.ShipperAddress,
			"receiver_name":       shipping.ReceiverName,
			"receiver_address":    shipping.ReceiverAddress,
			"summary_status":      shipping.SummaryStatus,
			"last_status":         shipping.LastStatus,
			"last_update_at":      shipping.LastUpdateAt,
			"note":                shipping.Note,
			"error_txt":           shipping.ErrorTxt,
			"resi_status":         shipping.ResiStatus,
			"version":             shipping.Version,
			"source":              "sicepat_webhook",

			// Latency tracking (matching Ruby structure exactly)
			"latency_count": shipping.LatencyCount,
			"time_received": shipping.TimeReceived,
		},
	}

	// Validate history events
	updates := []*domainservice.TrackingUpdate{update}

	logger.Info("sicepat_webhook_processed",
		"SiCepat webhook processed successfully",
		map[string]interface{}{
			"tracking_number": update.TrackingNumber,
			"status":          string(update.Status),
			"summary_status":  shipping.SummaryStatus,
			"last_status":     shipping.LastStatus,
			"location":        update.Location,
			"latency_seconds": shipping.LatencyCount,
		})

	return updates, nil
}

// ParsePayload processes payload from pubsub following Ruby reference exactly
func (h *SiCepatWebhookHandler) ParsePayload(payload []byte) (*Shipping, error) {
	var shipping Shipping
	err := json.Unmarshal(payload, &shipping)

	// Publish internal processing latency
	h.publishInToOutLatency(shipping.TimeReceived)

	// Get first history entry for business logic
	var history History
	if len(shipping.ShipmentHistories) > 0 {
		history = shipping.ShipmentHistories[0]
	}

	// Apply business logic from Ruby reference: set ShipperAddress only when status is IN (manifested)
	matcher, _ := regexp.Compile("manifested")
	if !(matcher.MatchString(strings.ToLower(history.Position)) && history.Status == "IN") {
		shipping.ShipperAddress = ""
	}

	// Set sent timestamp
	shipping.SentAt = time.Now().Format("2006-01-02T15:04:05.000Z")

	return &shipping, err
}

// ValidateSignature validates SiCepat webhook signature
func (h *SiCepatWebhookHandler) ValidateSignature(payload []byte, signature string) error {
	// SiCepat typically uses HMAC SHA256 with a specific format
	// Remove any "sicepat-signature=" prefix if present
	cleanSignature := strings.TrimPrefix(signature, "sicepat-signature=")
	cleanSignature = strings.TrimPrefix(cleanSignature, "sha256=")
	return h.ValidateHMACSignature(payload, cleanSignature)
}

// CreateErrorResponse returns error result when something happened from HandleRequest
func (h *SiCepatWebhookHandler) CreateErrorResponse() []byte {
	respBody := HTTPResponse{
		Success:      false,
		ErrorMessage: "ERROR",
	}
	b, _ := json.Marshal(respBody)
	return b
}

// CreateSuccessResponse returns success response
func (h *SiCepatWebhookHandler) CreateSuccessResponse() []byte {
	respBody := HTTPResponse{
		Success:      true,
		ErrorMessage: "",
	}
	b, _ := json.Marshal(respBody)
	return b
}

// normalizeSiCepatStatus converts SiCepat status codes to our standard tracking states
func (h *SiCepatWebhookHandler) normalizeSiCepatStatus(summaryStatus, lastStatus string) domainservice.TrackingState {
	// Check both summary and last status for comprehensive mapping
	status := strings.ToLower(strings.TrimSpace(summaryStatus + " " + lastStatus))

	switch {
	case strings.Contains(status, "booking"), strings.Contains(status, "created"),
		strings.Contains(status, "new"), strings.Contains(status, "pending"):
		return domainservice.TrackingStatePickupPending
	case strings.Contains(status, "pickup"), strings.Contains(status, "picked"),
		strings.Contains(status, "manifest"), strings.Contains(status, "collected"):
		return domainservice.TrackingStatePickedUp
	case strings.Contains(status, "transit"), strings.Contains(status, "processing"),
		strings.Contains(status, "sorting"), strings.Contains(status, "shipment"):
		return domainservice.TrackingStateInTransit
	case strings.Contains(status, "delivering"), strings.Contains(status, "out for delivery"),
		strings.Contains(status, "with courier"), strings.Contains(status, "on delivery"):
		return domainservice.TrackingStateOutForDelivery
	case strings.Contains(status, "delivered"), strings.Contains(status, "pod"),
		strings.Contains(status, "success"), strings.Contains(status, "complete"):
		return domainservice.TrackingStateDelivered
	case strings.Contains(status, "failed"), strings.Contains(status, "unsuccessful"),
		strings.Contains(status, "problem"), strings.Contains(status, "exception"):
		return domainservice.TrackingStateDeliveryFailed
	case strings.Contains(status, "returning"), strings.Contains(status, "return"):
		if strings.Contains(status, "returned") || strings.Contains(status, "complete") {
			return domainservice.TrackingStateReturned
		}
		return domainservice.TrackingStateReturning
	case strings.Contains(status, "cancelled"), strings.Contains(status, "void"),
		strings.Contains(status, "cancel"):
		return domainservice.TrackingStateException
	default:
		return domainservice.TrackingStateUnknown
	}
}

// validateHistoryEvents filters out history entries with 00:00:00 timestamps
// Based on the reference implementation's ValidateHistoryParam function
func (h *SiCepatWebhookHandler) validateHistoryEvents(events []*domainservice.TrackingUpdate) []*domainservice.TrackingUpdate {
	if len(events) == 0 {
		return events
	}

	validEvents := make([]*domainservice.TrackingUpdate, 0, len(events))
	for _, event := range events {
		if !event.Timestamp.IsZero() {
			// Check if time is 00:00:00
			hour, minute, second := event.Timestamp.Hour(), event.Timestamp.Minute(), event.Timestamp.Second()
			if hour == 0 && minute == 0 && second == 0 {
				logger.Info("sicepat_omitting_zero_time_event",
					"Omitting event with 00:00:00 timestamp",
					map[string]interface{}{
						"tracking_number": event.TrackingNumber,
						"status":          string(event.Status),
					})
				continue // Skip this event
			}
		}
		validEvents = append(validEvents, event)
	}

	return validEvents
}
