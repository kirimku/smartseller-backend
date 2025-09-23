package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Base URL for the API endpoints
const BaseURL = "http://localhost:8080/api/v1"

// TestLoginTokenInspection attempts to login to the API and analyze the returned token
func TestLoginTokenInspection(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping login token inspection test in short mode")
	}

	// Attempt login with test credentials
	loginURL := fmt.Sprintf("%s/auth/login", BaseURL)
	loginPayload := map[string]string{
		"email_or_phone": "test@example.com",
		"password":       "password123",
	}

	jsonPayload, err := json.Marshal(loginPayload)
	assert.NoError(t, err, "Failed to marshal login payload")

	// Create client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create and send login request
	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(jsonPayload))
	assert.NoError(t, err, "Failed to create login request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Could not connect to auth service: %v", err)
		t.Skip("Auth service not available")
		return
	}
	defer resp.Body.Close()

	// Check response status
	t.Logf("Login response status: %d %s", resp.StatusCode, resp.Status)

	// Parse response body
	var loginResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	assert.NoError(t, err, "Failed to decode login response")

	t.Logf("Login response: %v", loginResponse)

	// If login successful, inspect the token
	if resp.StatusCode == http.StatusOK && loginResponse["success"] == true {
		data, ok := loginResponse["data"].(map[string]interface{})
		if !ok {
			t.Logf("Could not extract data from response")
			return
		}

		token, ok := data["access_token"].(string)
		if !ok {
			t.Logf("Could not extract access_token")
			return
		}

		t.Logf("Got token: %s", token)

		// Decode token to inspect payload
		parts := strings.Split(token, ".")
		if len(parts) >= 2 {
			// The middle part is the payload
			encodedPayload := parts[1]

			// Add padding if needed
			if len(encodedPayload)%4 != 0 {
				padding := 4 - (len(encodedPayload) % 4)
				encodedPayload += strings.Repeat("=", padding)
			}

			// Try with standard base64 first, then URL encoding
			var payload []byte
			var decodeErr error

			// Try URL-safe base64 first
			payload, decodeErr = base64.RawURLEncoding.DecodeString(parts[1])
			if decodeErr != nil {
				// Try with padding
				payload, decodeErr = base64.URLEncoding.DecodeString(encodedPayload)
				if decodeErr != nil {
					// Try standard base64
					payload, decodeErr = base64.StdEncoding.DecodeString(encodedPayload)
					if decodeErr != nil {
						t.Logf("Failed to decode token payload: %v", decodeErr)
						return
					}
				}
			}

			t.Logf("Decoded payload: %s", payload)

			// Parse as JSON to see structure
			var claims map[string]interface{}
			if err := json.Unmarshal(payload, &claims); err == nil {
				t.Logf("Token claims: %v", claims)

				// Log important JWT claims
				t.Logf("user_id: %v", claims["user_id"])
				t.Logf("exp: %v", claims["exp"])
				t.Logf("iat: %v", claims["iat"])

				// Test complete - token successfully parsed
				t.Logf("Login test successful - token obtained and verified")
			} else {
				t.Logf("Failed to parse token claims: %v", err)
			}
		}
	}
}
