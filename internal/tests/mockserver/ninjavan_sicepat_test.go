package mockserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNinjaVanAndSiCepatAPIMock tests the NinjaVan and SiCepat API mock endpoints
func TestNinjaVanAndSiCepatAPIMock(t *testing.T) {
	mockServerURL := "http://localhost:1080"

	t.Run("NinjaVan OAuth Token", func(t *testing.T) {
		// Create request payload
		requestBody := map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"grant_type":    "client_credentials",
		}
		
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Make request
		resp, err := http.Post(mockServerURL+"/oauth/token", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for token
		token, ok := result["access_token"].(string)
		assert.True(t, ok, "Access token not found in response")
		assert.NotEmpty(t, token, "Access token is empty")
		
		// Print result
		fmt.Printf("NinjaVan OAuth Response: %+v\n", result)
	})

	t.Run("NinjaVan Pricing", func(t *testing.T) {
		// Create request payload
		requestBody := map[string]interface{}{
			"service_level": "Standard",
			"weight": 1000,
			"from": map[string]interface{}{
				"l1_tier_code": "JKT",
				"l2_tier_code": "JKT10",
			},
			"to": map[string]interface{}{
				"l1_tier_code": "BDG",
				"l2_tier_code": "BDG01",
			},
		}
		
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Make request
		resp, err := http.Post(mockServerURL+"/pricing", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for price data
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok, "Data field not found in response")
		assert.NotNil(t, data["total_fee"], "Total fee not found in response")
		
		// Print result
		fmt.Printf("NinjaVan Pricing Response: %+v\n", result)
	})

	t.Run("NinjaVan Orders", func(t *testing.T) {
		// Create request payload
		requestBody := map[string]interface{}{
			"reference_id": "TRX12345",
			"from": map[string]interface{}{
				"name": "Sender Name",
				"phone": "081234567890",
				"address": "Test Address",
			},
			"to": map[string]interface{}{
				"name": "Receiver Name",
				"phone": "089876543210",
				"address": "Test Address",
			},
			"parcel_job": map[string]interface{}{
				"service_level": "Standard",
				"pickup_time": "2023-05-03T10:00:00Z",
				"items": []map[string]interface{}{
					{
						"name": "Test Item",
						"quantity": 1,
						"price": 50000,
					},
				},
			},
		}
		
		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Make request
		resp, err := http.Post(mockServerURL+"/orders", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for tracking number
		data, ok := result["data"].(map[string]interface{})
		assert.True(t, ok, "Data field not found in response")
		assert.NotEmpty(t, data["tracking_number"], "Tracking number not found in response")
		
		// Print result
		fmt.Printf("NinjaVan Orders Response: %+v\n", result)
	})

	t.Run("SiCepat Tariff", func(t *testing.T) {
		// Build URL with query parameters
		baseURL := mockServerURL + "/customer/tariff"
		params := url.Values{}
		params.Add("api-key", "test_api_key")
		params.Add("origin", "JAKARTA")
		params.Add("destination", "BANDUNG")
		params.Add("weight", "1")
		
		// Make request
		resp, err := http.Get(baseURL + "?" + params.Encode())
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Check for results
		sicepat, ok := result["sicepat"].(map[string]interface{})
		assert.True(t, ok, "SiCepat field not found in response")
		
		status, ok := sicepat["status"].(map[string]interface{})
		assert.True(t, ok, "Status field not found in response")
		assert.Equal(t, "200", status["code"], "Status code is not 200")
		
		results, ok := sicepat["results"].([]interface{})
		assert.True(t, ok, "Results array not found in response")
		assert.NotEmpty(t, results, "Results array is empty")
		
		// Print result
		fmt.Printf("SiCepat Tariff Response: %+v\n", result)
	})
}