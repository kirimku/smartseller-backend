package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// JNE Delivered Status Codes based on Ruby reference implementation
// DELIVERED_CODES = %w[D01 D02 D03 D04 D05 D06 D07 D08 D09 D10 D11 D12 DB1]
var JNE_DELIVERED_CODES = []string{
	"D01", "D02", "D03", "D04", "D05", "D06",
	"D07", "D08", "D09", "D10", "D11", "D12", "DB1",
}

// JNEWebhookHandler handles webhooks from JNE
type JNEWebhookHandler struct {
	*BaseWebhookHandler
}

// JNEWebhookPayload represents the actual webhook payload structure from JNE
type JNEWebhookPayload struct {
	// Primary tracking number field (actual JNE webhook format)
	AirwaybillNumber string `json:"airwaybill_number"`

	// Legacy fields (for backward compatibility)
	CNoteNo string `json:"cnote_no"`
	AWB     string `json:"awb"`

	// Courier information
	CourierName    string `json:"courier_name"`
	CourierService string `json:"courier_service"`
	CourierDriver  struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Phone string `json:"phone"`
	} `json:"courier_driver"`

	// Shipment details
	ActualShippingFee float64   `json:"actual_shipping_fee"`
	ActualWeight      float64   `json:"actual_weight"`
	ShipmentDate      time.Time `json:"shipment_date"`

	// Address information
	ShipperName     string `json:"shipper_name"`
	ShipperAddress  string `json:"shipper_address"`
	ReceiverName    string `json:"receiver_name"`
	ReceiverAddress string `json:"receiver_address"`

	// Status information
	SummaryStatus string    `json:"summary_status"`
	LastStatus    string    `json:"last_status"`
	LastUpdateAt  time.Time `json:"last_update_at"`

	// Tracking history
	ShipmentHistories []struct {
		Position string    `json:"position"`
		Status   string    `json:"status"`
		Time     time.Time `json:"time"`
	} `json:"shipment_histories"`

	// Additional data
	AdditionalData struct {
		Flag string `json:"flag"`
	} `json:"additional_data"`

	// Legacy status fields (for backward compatibility)
	Status      string `json:"status"`
	StatusCode  string `json:"status_code"`
	StatusDesc  string `json:"status_desc"`
	Description string `json:"description"`
	Location    string `json:"location"`
	City        string `json:"city"`
	Office      string `json:"office"`

	// Legacy timestamp fields
	Timestamp time.Time `json:"timestamp"`
	EventDate string    `json:"event_date"`
	EventTime string    `json:"event_time"`

	// Metadata
	WebhookID    string  `json:"webhook_id"`
	Reference    string  `json:"reference"`
	Note         string  `json:"note"`
	ErrorTxt     string  `json:"error_txt"`
	ResiStatus   int     `json:"resi_status"`
	Version      string  `json:"version"`
	SentAt       string  `json:"sent_at"`
	TimeReceived float64 `json:"time_received"` // in Unix seconds; for in-to-out latency calculation
	LatencyCount float64 `json:"latency_count"` // in seconds

	// Legacy details (for backward compatibility)
	Details struct {
		Weight       float64 `json:"weight"`
		ServiceType  string  `json:"service_type"`
		ServiceName  string  `json:"service_name"`
		ReceiverName string  `json:"receiver_name"`
		SenderName   string  `json:"sender_name"`
		POD          string  `json:"pod"`
		PODPhoto     string  `json:"pod_photo"`
		DeliveryDate string  `json:"delivery_date"`
		DeliveryTime string  `json:"delivery_time"`
	} `json:"details"`
}

// NewJNEWebhookHandler creates a new JNE webhook handler
func NewJNEWebhookHandler(secretKey string) *JNEWebhookHandler {
	return &JNEWebhookHandler{
		BaseWebhookHandler: NewBaseWebhookHandler("jne", secretKey),
	}
}

// CountLatency calculates the latency between the event time and current time
func CountLatency(timeUpdated string) float64 {
	// JNE typically uses ISO8601 format, try multiple formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05+07:00",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		tr, err := time.Parse(format, timeUpdated)
		if err == nil {
			return time.Now().Sub(tr).Seconds()
		}
	}

	// Return a negative value to indicate error
	logger.Error("jne_invalid_timestamp", "Invalid timestamp format", nil)
	return -1
}

// publish3plToInLatency logs/publishes the latency from JNE to our service
func publish3plToInLatency(lastUpdateAt string) bool {
	latency := CountLatency(lastUpdateAt)
	if latency > 0 {
		// In a production environment, this could send to a metrics system
		logger.Info("jne_webhook_latency",
			"JNE webhook 3PL-to-In latency",
			map[string]interface{}{"latency_seconds": latency})
		return true
	}
	return false
}

// publishInToOutLatency logs/publishes the latency for our internal processing
func publishInToOutLatency(timeReceived float64) bool {
	if timeReceived > 0 {
		elapse := float64(time.Now().Unix()) - timeReceived
		logger.Info("jne_internal_processing_latency",
			"JNE webhook internal processing latency",
			map[string]interface{}{"latency_seconds": elapse})
		return true
	}
	return false
}

// HandleWebhook processes JNE webhook payload
func (h *JNEWebhookHandler) HandleWebhook(ctx context.Context, payload []byte) ([]*TrackingUpdate, error) {
	logger.Info("processing_jne_webhook",
		"Processing JNE webhook payload",
		map[string]interface{}{
			"payload_size": len(payload),
		})

	// Debug: Log the raw payload to understand the structure
	fmt.Printf("\n==== JNE WEBHOOK RAW PAYLOAD DEBUG ====\n")
	fmt.Printf("Raw Payload: %s\n", string(payload))
	fmt.Printf("========================================\n\n")

	var jnePayload JNEWebhookPayload
	if err := json.Unmarshal(payload, &jnePayload); err != nil {
		fmt.Printf("\n==== JNE WEBHOOK PARSE ERROR ====\n")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("=================================\n\n")
		return nil, fmt.Errorf("failed to parse JNE webhook payload: %w", err)
	}

	// Debug: Log the parsed payload structure
	fmt.Printf("\n==== JNE WEBHOOK PARSED PAYLOAD DEBUG ====\n")
	fmt.Printf("AirwaybillNumber: '%s'\n", jnePayload.AirwaybillNumber)
	fmt.Printf("CNoteNo: '%s'\n", jnePayload.CNoteNo)
	fmt.Printf("AWB: '%s'\n", jnePayload.AWB)
	fmt.Printf("CourierName: '%s'\n", jnePayload.CourierName)
	fmt.Printf("SummaryStatus: '%s'\n", jnePayload.SummaryStatus)
	fmt.Printf("LastStatus: '%s'\n", jnePayload.LastStatus)
	fmt.Printf("Status: '%s'\n", jnePayload.Status)
	fmt.Printf("StatusCode: '%s'\n", jnePayload.StatusCode)
	fmt.Printf("StatusDesc: '%s'\n", jnePayload.StatusDesc)
	fmt.Printf("Description: '%s'\n", jnePayload.Description)
	fmt.Printf("==========================================\n\n")

	// Store receipt time for latency calculation
	jnePayload.TimeReceived = float64(time.Now().Unix())

	// Validate required fields - JNE now uses airwaybill_number primarily
	trackingNumber := jnePayload.AirwaybillNumber
	if trackingNumber == "" {
		// Fallback to legacy fields for backward compatibility
		trackingNumber = jnePayload.CNoteNo
		if trackingNumber == "" {
			trackingNumber = jnePayload.AWB
		}
	}

	if trackingNumber == "" {
		// Debug: Log all fields to see what's actually available
		fmt.Printf("\n==== JNE WEBHOOK MISSING TRACKING NUMBER DEBUG ====\n")
		fmt.Printf("AirwaybillNumber (empty): '%s'\n", jnePayload.AirwaybillNumber)
		fmt.Printf("CNoteNo (empty): '%s'\n", jnePayload.CNoteNo)
		fmt.Printf("AWB (empty): '%s'\n", jnePayload.AWB)
		fmt.Printf("WebhookID: '%s'\n", jnePayload.WebhookID)
		fmt.Printf("Reference: '%s'\n", jnePayload.Reference)

		// Try to parse as generic JSON to see all available fields
		var genericPayload map[string]interface{}
		if err := json.Unmarshal(payload, &genericPayload); err == nil {
			fmt.Printf("All available fields:\n")
			for key, value := range genericPayload {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
		fmt.Printf("===================================================\n\n")

		return nil, fmt.Errorf("missing tracking number in JNE webhook")
	}

	// Parse event timestamp - handle both new and legacy formats
	var eventTime time.Time
	var timestampString string

	// Try new webhook format first
	if !jnePayload.LastUpdateAt.IsZero() {
		eventTime = jnePayload.LastUpdateAt
		timestampString = jnePayload.LastUpdateAt.Format(time.RFC3339)
	} else if jnePayload.EventDate != "" && jnePayload.EventTime != "" {
		// Legacy format: separate date and time fields
		eventTimeStr := fmt.Sprintf("%s %s", jnePayload.EventDate, jnePayload.EventTime)
		eventTime = normalizeTimestamp(eventTimeStr)
		timestampString = eventTimeStr
	} else if !jnePayload.Timestamp.IsZero() {
		// Legacy timestamp field
		eventTime = jnePayload.Timestamp
		timestampString = jnePayload.Timestamp.Format(time.RFC3339)
	} else {
		// Default to current time
		eventTime = time.Now()
		timestampString = eventTime.Format(time.RFC3339)
	}

	// Calculate and record latency
	publish3plToInLatency(timestampString)
	jnePayload.LatencyCount = CountLatency(timestampString)

	// Normalize JNE status to our standard tracking states
	// Use new webhook format status fields first, then fallback to legacy
	statusCode := jnePayload.LastStatus
	statusText := jnePayload.SummaryStatus

	// Fallback to legacy status fields if new ones are empty
	if statusCode == "" {
		statusCode = jnePayload.StatusCode
		if statusCode == "" {
			statusCode = jnePayload.Status
		}
	}

	if statusText == "" {
		statusText = jnePayload.StatusDesc
		if statusText == "" {
			statusText = jnePayload.Description
		}
	}

	// Use note field if no other status text is available
	if statusText == "" && jnePayload.Note != "" {
		statusText = jnePayload.Note
	}

	normalizedStatus := h.normalizeJNEStatus(statusCode, statusText)

	// Create tracking update
	update := &TrackingUpdate{
		TrackingNumber: trackingNumber,
		CourierCode:    "jne",
		Status:         normalizedStatus,
		StatusText:     statusText,
		Location:       h.buildLocation(jnePayload.Location, jnePayload.City, jnePayload.Office),
		Timestamp:      eventTime,
		Metadata: map[string]interface{}{
			// New webhook format fields
			"airwaybill_number":   jnePayload.AirwaybillNumber,
			"courier_name":        jnePayload.CourierName,
			"courier_service":     jnePayload.CourierService,
			"actual_shipping_fee": jnePayload.ActualShippingFee,
			"actual_weight":       jnePayload.ActualWeight,
			"shipment_date":       jnePayload.ShipmentDate,
			"shipper_name":        jnePayload.ShipperName,
			"shipper_address":     jnePayload.ShipperAddress,
			"receiver_name":       jnePayload.ReceiverName,
			"receiver_address":    jnePayload.ReceiverAddress,
			"summary_status":      jnePayload.SummaryStatus,
			"last_status":         jnePayload.LastStatus,
			"last_update_at":      jnePayload.LastUpdateAt,
			"note":                jnePayload.Note,
			"version":             jnePayload.Version,

			// Legacy fields for backward compatibility
			"webhook_id":    jnePayload.WebhookID,
			"reference":     jnePayload.Reference,
			"status_code":   jnePayload.StatusCode,
			"office":        jnePayload.Office,
			"weight":        jnePayload.Details.Weight,
			"service_type":  jnePayload.Details.ServiceType,
			"service_name":  jnePayload.Details.ServiceName,
			"sender_name":   jnePayload.Details.SenderName,
			"pod":           jnePayload.Details.POD,
			"pod_photo":     jnePayload.Details.PODPhoto,
			"delivery_date": jnePayload.Details.DeliveryDate,
			"delivery_time": jnePayload.Details.DeliveryTime,

			// Metadata
			"source":        "jne_webhook",
			"latency_count": jnePayload.LatencyCount,
			"time_received": jnePayload.TimeReceived,
		},
	}

	// Validate history events and publish internal processing latency
	updates := h.validateHistoryEvents([]*TrackingUpdate{update})
	publishInToOutLatency(jnePayload.TimeReceived)

	logger.Info("jne_webhook_processed",
		"JNE webhook processed successfully",
		map[string]interface{}{
			"tracking_number": update.TrackingNumber,
			"status":          string(update.Status),
			"status_code":     jnePayload.StatusCode,
			"location":        update.Location,
			"latency_seconds": jnePayload.LatencyCount,
		})

	return updates, nil
}

// ValidateSignature validates JNE webhook signature
func (h *JNEWebhookHandler) ValidateSignature(payload []byte, signature string) error {
	// JNE typically uses HMAC SHA256 with a specific format
	// Remove any "jne-signature=" prefix if present
	cleanSignature := strings.TrimPrefix(signature, "jne-signature=")
	cleanSignature = strings.TrimPrefix(cleanSignature, "sha256=")
	return h.ValidateHMACSignature(payload, cleanSignature)
}

// normalizeJNEStatus converts JNE status codes to our standard tracking states
// Based on Ruby reference implementation with delivered codes D01-D12, DB1
// Updated to handle new webhook format status values
func (h *JNEWebhookHandler) normalizeJNEStatus(status, statusCode string) TrackingStateTrackingState {
	// Check if status code indicates delivered status first
	if h.isDeliveredStatus(statusCode) {
		return TrackingStateTrackingStateDelivered
	}

	// Use status code for more precise mapping if available
	if statusCode != "" {
		switch strings.ToUpper(statusCode) {
		// New webhook format status codes
		case "MANIFESTED", "PICKUPED", "PICKUP_COMPLETED":
			return TrackingStateTrackingStatePickedUp
		case "IN_TRANSIT", "INTRANSIT", "ON_TRANSIT":
			return TrackingStateTrackingStateInTransit
		case "OUT_FOR_DELIVERY", "WITH_DELIVERY_COURIER":
			return TrackingStateTrackingStateOutForDelivery
		case "DELIVERED", "DELIVERY_COMPLETED":
			return TrackingStateTrackingStateDelivered
		case "DELIVERY_FAILED", "FAILED":
			return TrackingStateTrackingStateDeliveryFailed
		case "RETURN_TO_ORIGIN", "RETURNING":
			return TrackingStateTrackingStateReturning
		case "RETURNED", "RTO":
			return TrackingStateTrackingStateReturned
		case "CANCELLED", "VOID":
			return TrackingStateTrackingStateException

		// Legacy status codes
		case "BOOKING", "BOOKED", "CREATED", "B01":
			return TrackingStateTrackingStatePickupPending
		case "PICKUP", "PICKED", "MANIFEST", "M01", "M02":
			return TrackingStateTrackingStatePickedUp
		case "TRANSIT", "SORTING", "T01", "T02", "T03":
			return TrackingStateTrackingStateInTransit
		case "DELIVERING", "O01", "O02":
			return TrackingStateTrackingStateOutForDelivery
		case "UNSUCCESSFUL", "F01", "F02", "F03":
			return TrackingStateTrackingStateDeliveryFailed
		case "RETURN", "R01", "R02":
			return TrackingStateTrackingStateReturning
		case "R03":
			return TrackingStateTrackingStateReturned
		case "CANCEL", "C01":
			return TrackingStateTrackingStateException
		}
	}

	// Also check status text for new webhook format
	statusLower := strings.ToLower(strings.TrimSpace(status))
	switch {
	// New webhook format status mappings
	case strings.Contains(statusLower, "in_transit"), strings.Contains(statusLower, "in transit"):
		return TrackingStateTrackingStateInTransit
	case strings.Contains(statusLower, "manifested"):
		return TrackingStateTrackingStatePickedUp
	case strings.Contains(statusLower, "pickup_completed"), strings.Contains(statusLower, "pickup completed"):
		return TrackingStateTrackingStatePickedUp
	case strings.Contains(statusLower, "out_for_delivery"), strings.Contains(statusLower, "out for delivery"):
		return TrackingStateTrackingStateOutForDelivery
	case strings.Contains(statusLower, "delivery_completed"), strings.Contains(statusLower, "delivery completed"):
		return TrackingStateTrackingStateDelivered
	case strings.Contains(statusLower, "delivery_failed"), strings.Contains(statusLower, "delivery failed"):
		return TrackingStateTrackingStateDeliveryFailed
	case strings.Contains(statusLower, "return_to_origin"), strings.Contains(statusLower, "return to origin"):
		return TrackingStateTrackingStateReturning

	// Legacy status text mappings
	case strings.Contains(statusLower, "delivered"), strings.Contains(statusLower, "terkirim"),
		strings.Contains(statusLower, "selesai"), strings.Contains(statusLower, "diterima"):
		return TrackingStateTrackingStateDelivered
	case strings.Contains(statusLower, "booking"), strings.Contains(statusLower, "created"),
		strings.Contains(statusLower, "dibuat"):
		return TrackingStateTrackingStatePickupPending
	case strings.Contains(statusLower, "pickup"), strings.Contains(statusLower, "picked"),
		strings.Contains(statusLower, "manifest"), strings.Contains(statusLower, "diambil"):
		return TrackingStateTrackingStatePickedUp
	case strings.Contains(statusLower, "transit"), strings.Contains(statusLower, "sorting"),
		strings.Contains(statusLower, "perjalanan"):
		return TrackingStateTrackingStateInTransit
	case strings.Contains(statusLower, "delivering"), strings.Contains(statusLower, "pengiriman"):
		return TrackingStateTrackingStateOutForDelivery
	case strings.Contains(statusLower, "failed"), strings.Contains(statusLower, "unsuccessful"),
		strings.Contains(statusLower, "gagal"):
		return TrackingStateTrackingStateDeliveryFailed
	case strings.Contains(statusLower, "returning"), strings.Contains(statusLower, "return"):
		if strings.Contains(statusLower, "returned") || strings.Contains(statusLower, "dikembalikan") {
			return TrackingStateTrackingStateReturned
		}
		return TrackingStateTrackingStateReturning
	case strings.Contains(statusLower, "cancelled"), strings.Contains(statusLower, "void"),
		strings.Contains(statusLower, "dibatalkan"):
		return TrackingStateTrackingStateException
	default:
		return TrackingStateTrackingStateUnknown
	}
}

// isDeliveredStatus checks if the status code indicates a delivered status
// Based on Ruby reference: DELIVERED_CODES = %w[D01 D02 D03 D04 D05 D06 D07 D08 D09 D10 D11 D12 DB1]
func (h *JNEWebhookHandler) isDeliveredStatus(statusCode string) bool {
	if statusCode == "" {
		return false
	}

	statusUpper := strings.ToUpper(strings.TrimSpace(statusCode))

	// Check against delivered codes
	for _, deliveredCode := range JNE_DELIVERED_CODES {
		if statusUpper == deliveredCode {
			return true
		}
	}

	return false
}

// buildLocation combines location, city, and office information
func (h *JNEWebhookHandler) buildLocation(location, city, office string) string {
	var parts []string

	if location != "" {
		parts = append(parts, location)
	}
	if city != "" && city != location {
		parts = append(parts, city)
	}
	if office != "" && office != location && office != city {
		parts = append(parts, fmt.Sprintf("(%s)", office))
	}

	return strings.Join(parts, ", ")
}

// validateHistoryEvents filters out history entries with 00:00:00 timestamps
// Based on the reference implementation's ValidateHistoryParam function
func (h *JNEWebhookHandler) validateHistoryEvents(events []*TrackingUpdate) []*TrackingUpdate {
	if len(events) == 0 {
		return events
	}

	validEvents := make([]*TrackingUpdate, 0, len(events))
	for _, event := range events {
		if !event.Timestamp.IsZero() {
			// Check if time is 00:00:00
			hour, minute, second := event.Timestamp.Hour(), event.Timestamp.Minute(), event.Timestamp.Second()
			if hour == 0 && minute == 0 && second == 0 {
				logger.Info("jne_omitting_zero_time_event",
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

// TestNormalizeJNEStatus exposes the normalizeJNEStatus method for testing
func (h *JNEWebhookHandler) TestNormalizeJNEStatus(status, statusCode string) string {
	result := h.normalizeJNEStatus(status, statusCode)
	return string(result)
}
