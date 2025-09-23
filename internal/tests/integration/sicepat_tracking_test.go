package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kirimku/smartseller-backend/internal/config"
	domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external/sicepat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiCepatTracking_Integration(t *testing.T) {
	// Set environment variables for testing
	oldAPIURL := os.Getenv("SICEPAT_API_URL")
	oldAPIKey := os.Getenv("SICEPAT_API_KEY")

	// Start the mock server
	mockServerURL := mock_server.StartMockServer(8083)
	defer mock_server.StopMockServer()

	// Override environment variables to point to our mock server
	os.Setenv("SICEPAT_API_URL", mockServerURL)
	os.Setenv("SICEPAT_API_KEY", "test_api_key")

	// Restore environment variables after the test
	defer func() {
		os.Setenv("SICEPAT_API_URL", oldAPIURL)
		os.Setenv("SICEPAT_API_KEY", oldAPIKey)
	}()

	// Setup mock responses for tracking
	setupSiCepatTrackingMockResponses(t)

	// Load config
	err := config.LoadConfig()
	assert.NoError(t, err)

	// Create SiCepat client
	sicepatClient := sicepat.NewClient(sicepat.ClientConfig{
		BaseURL:            mockServerURL,
		APIKey:             "test_api_key",
		APIKeyBukasend:     "test_bukasend_key",
		APIKeyWhiteLabel:   "test_whitelabel_key",
		Timeout:            30 * time.Second,
		UseReceiptNumber:   false,
		UseToggleCondition: false,
	})

	t.Run("Track Success - Standard Flow", func(t *testing.T) {
		ctx := context.Background()
		trackingNumber := "000123456789"

		// Test Track method
		tracking, err := sicepatClient.Track(ctx, trackingNumber)

		assert.NoError(t, err)
		require.NotNil(t, tracking)

		// Verify basic tracking information
		assert.Equal(t, trackingNumber, tracking.TrackingNumber)
		assert.Equal(t, "SICEPAT", tracking.CourierName)
		assert.Equal(t, "Regular", tracking.ServiceType)
		assert.Equal(t, "Test Sender", tracking.SenderName)
		assert.Equal(t, "Test Receiver", tracking.ReceiverName)
		assert.Equal(t, "Jl. Test No. 123, Jakarta", tracking.Destination)

		// Verify weight conversion
		assert.Equal(t, 2.5, tracking.Weight) // 2500 grams converted to kg

		// Verify status normalization
		assert.Equal(t, domainservice.TrackingStateDelivered, tracking.Status)

		// Verify events
		assert.Len(t, tracking.Events, 4)

		// First event should be pickup/manifested
		firstEvent := tracking.Events[0]
		assert.Equal(t, domainservice.TrackingStatePickedUp, firstEvent.Status)
		assert.Equal(t, "MANIFESTED", firstEvent.Description)
		assert.Equal(t, "Jakarta Selatan", firstEvent.Location)

		// Last event should be delivered
		lastEvent := tracking.Events[len(tracking.Events)-1]
		assert.Equal(t, domainservice.TrackingStateDelivered, lastEvent.Status)
		assert.Equal(t, "DELIVERED", lastEvent.Description)

		// Verify delivery information
		assert.NotNil(t, tracking.DeliveryTime)
		assert.Equal(t, "Test Receiver", tracking.PODReceiver)

		// Verify timestamps are reasonable
		assert.False(t, tracking.UpdatedAt.IsZero())
		assert.False(t, tracking.DeliveryTime.IsZero())
	})

	t.Run("TrackToModel Success", func(t *testing.T) {
		ctx := context.Background()
		trackingNumber := "000123456789"

		// Test TrackToModel method
		trackingInfo, err := sicepatClient.TrackToModel(ctx, trackingNumber)

		assert.NoError(t, err)
		require.NotNil(t, trackingInfo)

		// Verify basic information following Ruby structure
		assert.Equal(t, trackingNumber, trackingInfo.AirwaybillNumber)
		assert.Equal(t, "SICEPAT", trackingInfo.CourierName)
		assert.Equal(t, "Regular", trackingInfo.CourierService)

		// Verify addresses and names
		assert.Equal(t, "Test Sender", trackingInfo.ShipperName)
		assert.Equal(t, "Test Receiver", trackingInfo.ReceiverName)
		assert.Equal(t, "Jl. Test No. 123, Jakarta", trackingInfo.ReceiverAddress)

		// Verify status information
		assert.Equal(t, "DELIVERED", trackingInfo.LastStatus)
		assert.Equal(t, "DELIVERED", trackingInfo.SummaryStatus)

		// Verify weight and fee (following Ruby conversion logic)
		assert.Equal(t, 2500, trackingInfo.ActualWeight)       // 2.5 kg = 2500 grams
		assert.Equal(t, 15000, trackingInfo.ActualShippingFee) // Original fee in smallest unit

		// Verify tracking history
		assert.Len(t, trackingInfo.ShipmentHistories, 4)

		// Check first history entry
		firstHistory := trackingInfo.ShipmentHistories[0]
		assert.Equal(t, "MANIFESTED", firstHistory.Status)
		assert.Equal(t, "Jakarta Selatan", firstHistory.Position)
		assert.False(t, firstHistory.Time.IsZero())

		// Check last history entry
		lastHistory := trackingInfo.ShipmentHistories[len(trackingInfo.ShipmentHistories)-1]
		assert.Equal(t, "DELIVERED", lastHistory.Status)
		assert.Equal(t, "Jakarta Timur", lastHistory.Position)

		// Verify additional data
		assert.Equal(t, "DELIVERED", trackingInfo.AdditionalData.Flag)

		// Verify status indicators
		assert.Equal(t, 1, trackingInfo.ResiStatus) // 1 for delivered
		assert.Empty(t, trackingInfo.ErrorTxt)      // No error for successful tracking

		// Verify timestamps
		assert.False(t, trackingInfo.ShipmentDate.IsZero())
		assert.False(t, trackingInfo.LastUpdateAt.IsZero())
	})

	t.Run("Track Not Found", func(t *testing.T) {
		ctx := context.Background()
		trackingNumber := "NOTFOUND123"

		// Test Track method with not found tracking number
		tracking, err := sicepatClient.Track(ctx, trackingNumber)

		// Should return error for not found
		assert.Error(t, err)
		assert.Nil(t, tracking)
		assert.Contains(t, err.Error(), "tracking failed with status 400")
	})

	t.Run("TrackToModel Not Found", func(t *testing.T) {
		ctx := context.Background()
		trackingNumber := "NOTFOUND123"

		// Test TrackToModel method with not found tracking number
		trackingInfo, err := sicepatClient.TrackToModel(ctx, trackingNumber)

		// Should return error
		assert.Error(t, err)
		assert.Nil(t, trackingInfo)
	})

	t.Run("Track In Transit Status", func(t *testing.T) {
		ctx := context.Background()
		trackingNumber := "INTRANSIT123"

		// Test Track method with in-transit package
		tracking, err := sicepatClient.Track(ctx, trackingNumber)

		assert.NoError(t, err)
		require.NotNil(t, tracking)

		// Verify status is in transit
		assert.Equal(t, domainservice.TrackingStateInTransit, tracking.Status)

		// Should not have delivery time for in-transit package
		assert.Nil(t, tracking.DeliveryTime)
		assert.Empty(t, tracking.PODReceiver)

		// Should have at least pickup event
		assert.NotEmpty(t, tracking.Events)
		firstEvent := tracking.Events[0]
		assert.Equal(t, domainservice.TrackingStatePickedUp, firstEvent.Status)
	})

	t.Run("Track Failed/Returned Status", func(t *testing.T) {
		ctx := context.Background()
		trackingNumber := "FAILED123"

		// Test Track method with failed package
		tracking, err := sicepatClient.Track(ctx, trackingNumber)

		assert.NoError(t, err)
		require.NotNil(t, tracking)

		// Verify status is failed
		assert.Equal(t, domainservice.TrackingStateFailed, tracking.Status)

		// Should not have delivery time for failed package
		assert.Nil(t, tracking.DeliveryTime)
		assert.Empty(t, tracking.PODReceiver)
	})
}

func TestSiCepatTracking_ErrorHandling(t *testing.T) {
	// Test with invalid API configuration
	sicepatClient := sicepat.NewClient(sicepat.ClientConfig{
		BaseURL:            "http://invalid-url-that-does-not-exist.com",
		APIKey:             "invalid_key",
		APIKeyBukasend:     "",
		APIKeyWhiteLabel:   "",
		Timeout:            1 * time.Second, // Short timeout
		UseReceiptNumber:   false,
		UseToggleCondition: false,
	})

	t.Run("Network Error", func(t *testing.T) {
		ctx := context.Background()
		trackingNumber := "000123456789"

		// Test should handle network error gracefully
		tracking, err := sicepatClient.Track(ctx, trackingNumber)

		assert.Error(t, err)
		assert.Nil(t, tracking)
		assert.Contains(t, err.Error(), "failed to get SiCepat tracking info")
	})

	t.Run("Timeout Error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		trackingNumber := "000123456789"

		// Test should handle timeout gracefully
		tracking, err := sicepatClient.Track(ctx, trackingNumber)

		assert.Error(t, err)
		assert.Nil(t, tracking)
	})
}

// setupSiCepatTrackingMockResponses sets up mock HTTP responses for SiCepat tracking API
func setupSiCepatTrackingMockResponses(t *testing.T) {
	// Success response for normal tracking
	successResponse := sicepat.SicepatTrackingResponse{
		Sicepat: sicepat.SicepatData{
			Status: sicepat.SicepatStatus{
				Code:        200,
				Description: "OK",
			},
			Result: &sicepat.SicepatResult{
				Service:      "Regular",
				SendDate:     "2024-12-20 10:30",
				Sender:       "Test Sender",
				Weight:       "2.5",
				TotalPrice:   "15000",
				KodeAsal:     "Jakarta Selatan",
				ReceiverName: "Test Receiver",
				ReceiverAddr: "Jl. Test No. 123, Jakarta",
				LastStatus: sicepat.SicepatLastStatus{
					Status:   "DELIVERED",
					DateTime: "2024-12-22 14:30",
				},
				TrackHistory: []sicepat.SicepatTrackingHistory{
					{
						DateTime: "2024-12-20 10:30",
						Status:   "MANIFESTED",
						City:     "Jakarta Selatan",
					},
					{
						DateTime: "2024-12-21 09:15",
						Status:   "IN TRANSIT",
						City:     "Bekasi Hub",
					},
					{
						DateTime: "2024-12-22 08:00",
						Status:   "OUT FOR DELIVERY",
						City:     "Jakarta Timur",
					},
					{
						DateTime: "2024-12-22 14:30",
						Status:   "DELIVERED",
						City:     "Jakarta Timur",
					},
				},
			},
		},
	}

	// In transit response
	inTransitResponse := sicepat.SicepatTrackingResponse{
		Sicepat: sicepat.SicepatData{
			Status: sicepat.SicepatStatus{
				Code:        200,
				Description: "OK",
			},
			Result: &sicepat.SicepatResult{
				Service:      "Regular",
				SendDate:     "2024-12-20 10:30",
				Sender:       "Test Sender",
				Weight:       "1.5",
				TotalPrice:   "12000",
				KodeAsal:     "Jakarta Selatan",
				ReceiverName: "Test Receiver",
				ReceiverAddr: "Jl. Test No. 456, Bandung",
				LastStatus: sicepat.SicepatLastStatus{
					Status:   "IN TRANSIT",
					DateTime: "2024-12-21 09:15",
				},
				TrackHistory: []sicepat.SicepatTrackingHistory{
					{
						DateTime: "2024-12-20 10:30",
						Status:   "MANIFESTED",
						City:     "Jakarta Selatan",
					},
					{
						DateTime: "2024-12-21 09:15",
						Status:   "IN TRANSIT",
						City:     "Bekasi Hub",
					},
				},
			},
		},
	}

	// Failed/returned response
	failedResponse := sicepat.SicepatTrackingResponse{
		Sicepat: sicepat.SicepatData{
			Status: sicepat.SicepatStatus{
				Code:        200,
				Description: "OK",
			},
			Result: &sicepat.SicepatResult{
				Service:      "Regular",
				SendDate:     "2024-12-20 10:30",
				Sender:       "Test Sender",
				Weight:       "1.0",
				TotalPrice:   "10000",
				KodeAsal:     "Jakarta Selatan",
				ReceiverName: "Test Receiver",
				ReceiverAddr: "Jl. Test No. 789, Surabaya",
				LastStatus: sicepat.SicepatLastStatus{
					Status:   "RETURNED TO SENDER",
					DateTime: "2024-12-23 16:00",
				},
				TrackHistory: []sicepat.SicepatTrackingHistory{
					{
						DateTime: "2024-12-20 10:30",
						Status:   "MANIFESTED",
						City:     "Jakarta Selatan",
					},
					{
						DateTime: "2024-12-21 09:15",
						Status:   "IN TRANSIT",
						City:     "Surabaya Hub",
					},
					{
						DateTime: "2024-12-22 14:00",
						Status:   "DELIVERY FAILED",
						City:     "Surabaya",
					},
					{
						DateTime: "2024-12-23 16:00",
						Status:   "RETURNED TO SENDER",
						City:     "Jakarta Selatan",
					},
				},
			},
		},
	}

	// Not found response
	notFoundResponse := sicepat.SicepatTrackingResponse{
		Sicepat: sicepat.SicepatData{
			Status: sicepat.SicepatStatus{
				Code:        400,
				Description: "Data not found",
			},
			Result: nil,
		},
	}

	// Add mock HTTP handlers
	mock_server.AddHandler("GET", "/customer/waybill", func(w http.ResponseWriter, r *http.Request) {
		waybill := r.URL.Query().Get("waybill")
		apiKey := r.URL.Query().Get("api-key")

		// Verify API key
		if apiKey != "test_api_key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Return appropriate response based on waybill
		switch waybill {
		case "000123456789":
			json.NewEncoder(w).Encode(successResponse)
		case "INTRANSIT123":
			json.NewEncoder(w).Encode(inTransitResponse)
		case "FAILED123":
			json.NewEncoder(w).Encode(failedResponse)
		case "NOTFOUND123":
			json.NewEncoder(w).Encode(notFoundResponse)
		default:
			json.NewEncoder(w).Encode(notFoundResponse)
		}
	})
}

// Test SiCepat client tracking service integration
func TestSiCepatClient_TrackingServiceIntegration(t *testing.T) {
	// Create client with mock configuration
	client := sicepat.NewClient(sicepat.ClientConfig{
		BaseURL:            "https://api.sicepat.com",
		APIKey:             "test_api_key",
		APIKeyBukasend:     "test_bukasend_key",
		APIKeyWhiteLabel:   "test_whitelabel_key",
		Timeout:            30 * time.Second,
		UseReceiptNumber:   false,
		UseToggleCondition: false,
	})

	t.Run("Tracking Service Available", func(t *testing.T) {
		trackingService := client.GetTrackingService()
		assert.NotNil(t, trackingService)
	})

	t.Run("Track Method Available", func(t *testing.T) {
		// Just test that the method exists and can be called
		// (will fail with network error but that's expected)
		ctx := context.Background()
		_, err := client.Track(ctx, "test123")

		// Should get a network error, not a method not found error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get SiCepat tracking info")
	})

	t.Run("TrackToModel Method Available", func(t *testing.T) {
		// Just test that the method exists and can be called
		ctx := context.Background()
		_, err := client.TrackToModel(ctx, "test123")

		// Should get a network error, not a method not found error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute tracking request")
	})
}
