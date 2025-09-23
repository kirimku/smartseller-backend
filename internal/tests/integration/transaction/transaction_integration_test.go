package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
)

// Basic API endpoint URL
const (
	BaseURL = "http://localhost:8080/api/v1"
)

// getAuthToken obtains an authentication token for testing
func getAuthToken(t *testing.T) string {
	// Try to get a real token by logging in
	loginURL := fmt.Sprintf("%s/auth/login", BaseURL)

	// Default test credentials
	email := "test@example.com"
	password := "password123"

	t.Logf("Attempting login with email: %s", email)

	loginPayload := map[string]string{
		"email_or_phone": email,
		"password":       password,
	}

	jsonPayload, err := json.Marshal(loginPayload)
	require.NoError(t, err, "Failed to marshal login payload")

	// Create a client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(jsonPayload))
	require.NoError(t, err, "Failed to create login request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err, "Could not connect to auth service")
	defer resp.Body.Close()

	// If login fails
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status code %d", resp.StatusCode)
	}

	var loginResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	require.NoError(t, err, "Failed to decode login response")

	// Extract token from response
	data, ok := loginResponse["data"].(map[string]interface{})
	require.True(t, ok, "Data field is not an object")

	token, ok := data["access_token"].(string)
	require.True(t, ok, "Could not extract access_token from response")

	t.Logf("Successfully logged in and retrieved authentication token")
	return token
}

// TestCreateTransaction tests the transaction creation endpoint
func TestCreateTransaction(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get authentication token
	authToken := getAuthToken(t)

	// Create a base request for tests
	baseRequest := createSampleTransactionRequest()

	// Test cases
	t.Run("Successful Transaction Creation", func(t *testing.T) {
		testSuccessfulTransactionCreation(t, baseRequest, authToken)
	})

	t.Run("Missing Required Field", func(t *testing.T) {
		testMissingRequiredField(t, baseRequest, authToken)
	})

	t.Run("Invalid Weight", func(t *testing.T) {
		testInvalidWeight(t, baseRequest, authToken)
	})

	t.Run("Invalid Courier", func(t *testing.T) {
		testInvalidCourier(t, baseRequest, authToken)
	})

	t.Run("COD Transaction", func(t *testing.T) {
		testCODTransaction(t, baseRequest, authToken)
	})

	t.Run("Invalid Address", func(t *testing.T) {
		testInvalidAddress(t, baseRequest, authToken)
	})

	t.Run("Idempotent Creation", func(t *testing.T) {
		testIdempotentCreation(t, baseRequest, authToken)
	})
}

// createSampleTransactionRequest creates a sample transaction request
func createSampleTransactionRequest() dto.TransactionCreateRequest {
	return dto.TransactionCreateRequest{
		OrderName:          "Test Order",
		OrderPrice:         50000,
		Weight:             1000, // 1kg in grams
		Length:             20,
		Width:              15,
		Height:             10,
		Notes:              "Handle with care",
		Courier:            "jne",
		CourierServiceType: "reg",
		Platform:           "web",
		COD:                false,
		UniqueID:           fmt.Sprintf("test-txn-%d", time.Now().UnixNano()),
		From: dto.TransactionAddressDTO{
			Name:     "John Sender",
			Phone:    "08123456789",
			Email:    "john@example.com",
			Province: "DKI Jakarta",
			City:     "Jakarta Selatan",
			Area:     "Kebayoran Baru",
			Address:  "Jl. Test No. 123",
			PostCode: "12190",
		},
		To: dto.TransactionAddressDTO{
			Name:     "Jane Receiver",
			Phone:    "08987654321",
			Email:    "jane@example.com",
			Province: "Jawa Barat",
			City:     "Bandung",
			Area:     "Cicendo",
			Address:  "Jl. Testing No. 456",
			PostCode: "40172",
		},
	}
}

// testSuccessfulTransactionCreation tests successful transaction creation
func testSuccessfulTransactionCreation(t *testing.T, request dto.TransactionCreateRequest, authToken string) {
	// Debug the token
	t.Logf("Using auth token: %s", authToken)

	// Create a copy for this test
	testRequest := request
	testRequest.UniqueID = fmt.Sprintf("test-txn-%d", time.Now().UnixNano())

	// Serialize the request to JSON
	requestBody, err := json.Marshal(testRequest)
	require.NoError(t, err, "Failed to marshal request")

	// Create HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()

	// Check response status code
	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			t.Logf("Error response: %v", errResponse)
		}
		return
	}

	// Parse the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err, "Failed to unmarshal response")

	// Check response structure
	success, ok := response["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.True(t, success, "Expected success to be true")

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "Response missing 'data' field or not an object")

	// Check transaction data
	assert.Contains(t, data, "id", "Expected transaction to have an ID")
	assert.Equal(t, "pending", data["state"], "Expected initial state to be pending")
	assert.Equal(t, testRequest.OrderName, data["products"].(map[string]interface{})["name"], "Order name mismatch")
	assert.Equal(t, testRequest.Courier, data["delivery"].(map[string]interface{})["courier_code"], "Courier mismatch")

	// Verify shipping fee calculation was performed
	assert.Greater(t, data["total_amount"].(float64), testRequest.OrderPrice, "Expected total amount to include shipping fee")

	// Verify cost components were added
	assert.Contains(t, data, "cost_components", "Expected cost components in response")
	components := data["cost_components"].([]interface{})
	assert.NotEmpty(t, components, "Expected at least one cost component")
}

// testMissingRequiredField tests validation for missing required fields
func testMissingRequiredField(t *testing.T, request dto.TransactionCreateRequest, authToken string) {
	// Create a copy with missing required field
	invalidRequest := request
	invalidRequest.OrderName = "" // OrderName is required
	invalidRequest.UniqueID = fmt.Sprintf("test-txn-%d", time.Now().UnixNano())

	// Serialize the request to JSON
	requestBody, err := json.Marshal(invalidRequest)
	require.NoError(t, err, "Failed to marshal request")

	// Create HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()

	// Check response status code
	if !assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected status code 400") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			t.Logf("Unexpected status code: got %d, want %d. Response: %v", resp.StatusCode, http.StatusBadRequest, errResponse)
		}
		return
	}

	// Parse the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err, "Failed to unmarshal response")

	// Check error message
	success, ok := response["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.False(t, success, "Expected success to be false")

	errorMsg, ok := response["error"].(string)
	if ok {
		assert.Contains(t, errorMsg, "order_name", "Expected error to mention missing field")
	} else {
		errorObj, ok := response["error"].(map[string]interface{})
		if ok {
			t.Logf("Error object: %v", errorObj)
			// Check if there's a message or details field that might contain the field name
			for _, v := range errorObj {
				if str, ok := v.(string); ok {
					assert.Contains(t, str, "order_name", "Expected error to mention missing field")
				}
			}
		} else {
			t.Logf("Unexpected error format: %v", response["error"])
			t.Fail()
		}
	}
}

// testInvalidWeight tests validation for invalid weight
func testInvalidWeight(t *testing.T, request dto.TransactionCreateRequest, authToken string) {
	// Create a copy with invalid weight
	invalidRequest := request
	invalidRequest.Weight = 0 // Weight must be positive
	invalidRequest.UniqueID = fmt.Sprintf("test-txn-%d", time.Now().UnixNano())

	// Serialize the request to JSON
	requestBody, err := json.Marshal(invalidRequest)
	require.NoError(t, err, "Failed to marshal request")

	// Create HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()

	// Check response status code
	if !assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected status code 400") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			t.Logf("Unexpected status code: got %d, want %d. Response: %v", resp.StatusCode, http.StatusBadRequest, errResponse)
		}
		return
	}

	// Parse the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err, "Failed to unmarshal response")

	// Check error message
	success, ok := response["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.False(t, success, "Expected success to be false")

	// Check if error is a string or an object
	errorField := response["error"]
	if errorStr, ok := errorField.(string); ok {
		assert.Contains(t, errorStr, "weight", "Expected error to mention weight")
	} else if errorObj, ok := errorField.(map[string]interface{}); ok {
		t.Logf("Error object: %v", errorObj)
		found := false
		// Check all parts of the error object for the field name
		for _, v := range errorObj {
			if str, ok := v.(string); ok && strings.Contains(str, "weight") {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error to mention weight")
	} else {
		t.Logf("Unexpected error format: %T %v", errorField, errorField)
		t.Fail()
	}
}

// testInvalidCourier tests validation for invalid courier
func testInvalidCourier(t *testing.T, request dto.TransactionCreateRequest, authToken string) {
	// Create a copy with invalid courier
	invalidRequest := request
	invalidRequest.Courier = "invalid_courier" // Not in the allowed list
	invalidRequest.UniqueID = fmt.Sprintf("test-txn-%d", time.Now().UnixNano())

	// Serialize the request to JSON
	requestBody, err := json.Marshal(invalidRequest)
	require.NoError(t, err, "Failed to marshal request")

	// Create HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()

	// Check response status code
	if !assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected status code 400") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			t.Logf("Unexpected status code: got %d, want %d. Response: %v", resp.StatusCode, http.StatusBadRequest, errResponse)
		}
		return
	}

	// Parse the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err, "Failed to unmarshal response")

	// Check error message
	success, ok := response["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.False(t, success, "Expected success to be false")

	// Check if error is a string or an object
	errorField := response["error"]
	if errorStr, ok := errorField.(string); ok {
		assert.Contains(t, errorStr, "courier", "Expected error to mention courier")
	} else if errorObj, ok := errorField.(map[string]interface{}); ok {
		t.Logf("Error object: %v", errorObj)
		found := false
		// Check all parts of the error object for the field name
		for _, v := range errorObj {
			if str, ok := v.(string); ok && strings.Contains(str, "courier") {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error to mention courier")
	} else {
		t.Logf("Unexpected error format: %T %v", errorField, errorField)
		t.Fail()
	}
}

// testCODTransaction tests creation of COD transaction
func testCODTransaction(t *testing.T, request dto.TransactionCreateRequest, authToken string) {
	// Create a COD transaction request
	codRequest := request
	codRequest.COD = true
	codRequest.CODValue = 50000
	codRequest.CODAdminFeePaidBy = "seller"
	codRequest.UniqueID = fmt.Sprintf("test-cod-txn-%d", time.Now().UnixNano())

	// Serialize the request to JSON
	requestBody, err := json.Marshal(codRequest)
	require.NoError(t, err, "Failed to marshal request")

	// Create HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()

	// Check response status code
	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			t.Logf("Error response: %v", errResponse)
		}
		return
	}

	// Parse the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err, "Failed to unmarshal response")

	// Check response structure
	success, ok := response["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.True(t, success, "Expected success to be true")

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "Response missing 'data' field or not an object")

	// Check COD details in response
	assert.True(t, data["cod"].(bool), "Expected COD to be true")

	codDetail, ok := data["cod_detail"].(map[string]interface{})
	require.True(t, ok, "Missing cod_detail field or not an object")
	assert.Equal(t, 50000.0, codDetail["value"].(float64), "COD value mismatch")
	assert.Equal(t, "seller", codDetail["admin_fee_paid_by"].(string), "COD admin fee paid by mismatch")

	// Check that COD admin fee is included in cost components
	components, ok := data["cost_components"].([]interface{})
	require.True(t, ok, "Missing cost_components field or not an array")
	var foundCODAdminFee bool
	for _, comp := range components {
		component, ok := comp.(map[string]interface{})
		if !ok {
			continue
		}
		compType, ok := component["type"].(string)
		if !ok {
			continue
		}
		if compType == "cod_fee" {
			foundCODAdminFee = true
			break
		}
	}
	assert.True(t, foundCODAdminFee, "Expected COD admin fee in cost components")
}

// testInvalidAddress tests validation for invalid address
func testInvalidAddress(t *testing.T, request dto.TransactionCreateRequest, authToken string) {
	// Create a copy with invalid address
	invalidRequest := request
	invalidRequest.To.City = "" // City is required
	invalidRequest.UniqueID = fmt.Sprintf("test-txn-%d", time.Now().UnixNano())

	// Serialize the request to JSON
	requestBody, err := json.Marshal(invalidRequest)
	require.NoError(t, err, "Failed to marshal request")

	// Create HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send the request
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()

	// Check response status code
	if !assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected status code 400") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResponse); err == nil {
			t.Logf("Unexpected status code: got %d, want %d. Response: %v", resp.StatusCode, http.StatusBadRequest, errResponse)
		}
		return
	}

	// Parse the response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err, "Failed to unmarshal response")

	// Check error message
	success, ok := response["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.False(t, success, "Expected success to be false")

	// Check if error is a string or an object
	errorField := response["error"]
	if errorStr, ok := errorField.(string); ok {
		assert.Contains(t, errorStr, "city", "Expected error to mention city")
	} else if errorObj, ok := errorField.(map[string]interface{}); ok {
		t.Logf("Error object: %v", errorObj)
		found := false
		// Check all parts of the error object for the field name
		for _, v := range errorObj {
			if str, ok := v.(string); ok && strings.Contains(str, "city") {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error to mention city")
	} else {
		t.Logf("Unexpected error format: %T %v", errorField, errorField)
		t.Fail()
	}
}

// testIdempotentCreation tests idempotent transaction creation
func testIdempotentCreation(t *testing.T, request dto.TransactionCreateRequest, authToken string) {
	// Create a unique ID for this test
	idempotentRequest := request
	idempotentRequest.UniqueID = fmt.Sprintf("idempotent-txn-%d", time.Now().UnixNano())

	// Serialize the request to JSON
	requestBody, err := json.Marshal(idempotentRequest)
	require.NoError(t, err, "Failed to marshal request")

	client := &http.Client{}

	// First request should succeed
	req1, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp1, err := client.Do(req1)
	require.NoError(t, err, "Failed to send first request")
	defer resp1.Body.Close()

	if !assert.Equal(t, http.StatusCreated, resp1.StatusCode, "First request should succeed") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp1.Body).Decode(&errResponse); err == nil {
			t.Logf("Error response for first request: %v", errResponse)
		}
		return
	}

	// Extract transaction ID from first response
	var response1 map[string]interface{}
	err = json.NewDecoder(resp1.Body).Decode(&response1)
	require.NoError(t, err, "Failed to unmarshal first response")

	success1, ok := response1["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.True(t, success1, "Expected success to be true")

	data1, ok := response1["data"].(map[string]interface{})
	require.True(t, ok, "First response missing 'data' field or not an object")

	id1, ok := data1["id"].(float64)
	require.True(t, ok, "Transaction ID not found or not a number")
	transactionID1 := int(id1)

	// Second request with same unique ID should return the same transaction
	req2, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/transactions", BaseURL), bytes.NewBuffer(requestBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp2, err := client.Do(req2)
	require.NoError(t, err, "Failed to send second request")
	defer resp2.Body.Close()

	if !assert.Equal(t, http.StatusOK, resp2.StatusCode, "Second request should return 200 OK") {
		// Parse error response for debugging
		var errResponse map[string]interface{}
		if err := json.NewDecoder(resp2.Body).Decode(&errResponse); err == nil {
			t.Logf("Error response for second request: %v", errResponse)
		}
		return
	}

	// Extract transaction ID from second response
	var response2 map[string]interface{}
	err = json.NewDecoder(resp2.Body).Decode(&response2)
	require.NoError(t, err, "Failed to unmarshal second response")

	success2, ok := response2["success"].(bool)
	require.True(t, ok, "Response missing 'success' field or not a boolean")
	assert.True(t, success2, "Expected success to be true")

	data2, ok := response2["data"].(map[string]interface{})
	require.True(t, ok, "Second response missing 'data' field or not an object")

	id2, ok := data2["id"].(float64)
	require.True(t, ok, "Transaction ID not found or not a number")
	transactionID2 := int(id2)

	// Verify both requests returned the same transaction ID
	assert.Equal(t, transactionID1, transactionID2, "Both requests should return the same transaction ID")
}
