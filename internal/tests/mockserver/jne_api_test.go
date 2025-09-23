package mockserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJNEAPIMock tests the JNE API mock endpoints directly
func TestJNEAPIMock(t *testing.T) {
	mockServerURL := "http://localhost:3001"

	t.Run("JNE Price Calculation", func(t *testing.T) {
		// Create request payload
		requestBody := map[string]interface{}{
			"username": "testuser",
			"api_key":  "testkey",
			"from":     "JAKARTA",
			"thru":     "BANDUNG",
			"weight":   1,
		}
		
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Make request
		resp, err := http.Post(mockServerURL+"/tracing/api/pricedev", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for price array
		priceArray, ok := result["price"].([]interface{})
		assert.True(t, ok, "Price array not found in response")
		assert.NotEmpty(t, priceArray, "Price array is empty")
		
		// Print result
		fmt.Printf("JNE Price Calculation Response: %+v\n", result)
	})

	t.Run("JNE Booking", func(t *testing.T) {
		// Create request payload
		requestBody := map[string]interface{}{
			"username":      "testuser",
			"api_key":       "testkey",
			"SHIPPER_NAME":  "Sender",
			"SHIPPER_ADDR1": "Test Address",
			"SHIPPER_CITY":  "JAKARTA",
			"RECEIVER_NAME": "Receiver",
			"RECEIVER_ADDR1": "Test Address",
			"RECEIVER_CITY":  "BANDUNG",
			"SERVICE":        "REG",
			"WEIGHT":         1,
		}
		
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Make request
		resp, err := http.Post(mockServerURL+"/tracing/api/generatecnote", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for booking status
		status, ok := result["status"].(bool)
		assert.True(t, ok, "Status field not found in response")
		assert.True(t, status, "Booking status is not true")
		
		// Check for cnote (AWB number)
		cnote, ok := result["cnote"].(string)
		assert.True(t, ok, "Cnote field not found in response")
		assert.NotEmpty(t, cnote, "Cnote is empty")
		
		// Print result
		fmt.Printf("JNE Booking Response: %+v\n", result)
	})

	t.Run("JNE Cancellation", func(t *testing.T) {
		// Create request payload
		requestBody := map[string]interface{}{
			"username": "testuser",
			"api_key":  "testkey",
			"awb":      "JNE0123456789",
		}
		
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Make request
		resp, err := http.Post(mockServerURL+"/tracing/api/cancelcnote", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for cancellation status
		status, ok := result["status"].(bool)
		assert.True(t, ok, "Status field not found in response")
		assert.True(t, status, "Cancellation status is not true")
		
		// Print result
		fmt.Printf("JNE Cancellation Response: %+v\n", result)
	})
	
	t.Run("JNE Error Handling - Invalid Weight", func(t *testing.T) {
		// Create request payload with invalid weight (0)
		requestBody := map[string]interface{}{
			"username": "testuser",
			"api_key":  "testkey",
			"from":     "JAKARTA",
			"thru":     "BANDUNG",
			"weight":   0,
		}
		
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Make request
		resp, err := http.Post(mockServerURL+"/tracing/api/pricedev", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response (expecting 400 Bad Request)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		
		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		
		// Print result
		fmt.Printf("JNE Error Response (Invalid Weight): %+v\n", result)
	})
}