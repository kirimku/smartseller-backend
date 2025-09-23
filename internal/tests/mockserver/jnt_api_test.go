package mockserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJNTAPIMock tests the JNT API mock endpoints directly
func TestJNTAPIMock(t *testing.T) {
	mockServerURL := "http://localhost:3001"

	t.Run("JNT Tariff Inquiry", func(t *testing.T) {
		// Create form data
		formData := url.Values{}
		formData.Add("serviceType", "REG")
		formData.Add("weight", "1000")
		formData.Add("srcCity", "Jakarta")
		formData.Add("destCity", "Bandung")
		formData.Add("signature", "test_signature") // Add required signature

		// Make request
		resp, err := http.PostForm(mockServerURL+"/jandt_track/inquiry.action", formData)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for success
		assert.True(t, result["is_success"].(bool))

		// Print result
		fmt.Printf("JNT Tariff Inquiry Response: %+v\n", result)
	})

	t.Run("JNT Booking", func(t *testing.T) {
		// Create form data
		formData := url.Values{}
		formData.Add("apiKey", "test_api_key")
		formData.Add("pickupName", "Sender")
		formData.Add("pickupPhone", "081234567890")
		formData.Add("receiverName", "Receiver")
		formData.Add("receiverPhone", "081234567891")
		formData.Add("receiverCity", "Bandung")
		formData.Add("weight", "1000")
		formData.Add("signature", "test_signature") // Add required signature

		// Make request
		resp, err := http.PostForm(mockServerURL+"/jandt_track/order.action", formData)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for success
		assert.True(t, result["is_success"].(bool))

		// Print result
		fmt.Printf("JNT Booking Response: %+v\n", result)
	})

	t.Run("JNT Cancellation", func(t *testing.T) {
		// Create form data
		formData := url.Values{}
		formData.Add("apiKey", "test_api_key")
		formData.Add("awb", "JNT12345678")
		formData.Add("signature", "test_signature") // Add required signature

		// Make request
		resp, err := http.PostForm(mockServerURL+"/jandt_track/update.action", formData)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for success
		assert.True(t, result["is_success"].(bool))

		// Print result
		fmt.Printf("JNT Cancellation Response: %+v\n", result)
	})
}