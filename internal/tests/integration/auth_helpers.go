package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// API endpoint URL for authentication
const (
	AuthBaseURL = "http://localhost:8080/api/v1"
)

// GetBaseURL returns the base URL for API endpoints
func GetBaseURL() string {
	return AuthBaseURL
}

// GetAuthToken obtains an authentication token for testing
// It attempts to login with test credentials, and if that fails,
// generates a test token signed with the same secret key as the server.
// This is an exported version of getAuthToken for use in other packages.
func GetAuthToken(t *testing.T) string {
	// First, check if there's a test token in the environment
	testToken := os.Getenv("TEST_TOKEN")
	if testToken != "" {
		t.Log("Using TEST_TOKEN from environment variables")
		return testToken
	}

	// Try to get a real token by logging in
	loginURL := fmt.Sprintf("%s/auth/login", AuthBaseURL)

	// Check if environment variables are set for test user credentials
	email := os.Getenv("TEST_USER_EMAIL")
	password := os.Getenv("TEST_USER_PASSWORD")

	// Default test credentials if not set in environment
	if email == "" {
		email = "test@example.com"
	}
	if password == "" {
		password = "password123"
	}

	t.Logf("Attempting login with email: %s", email)

	loginPayload := map[string]string{
		"email_or_phone": email,
		"password":       password,
	}

	jsonPayload, err := json.Marshal(loginPayload)
	require.NoError(t, err, "Failed to marshal login payload")

	// Create a client with a timeout
	client := CreateHTTPClient()

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(jsonPayload))
	require.NoError(t, err, "Failed to create login request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Could not connect to auth service: %v", err)
		return generateTestToken(t)
	}
	defer resp.Body.Close()

	// If login fails
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errorResp)
		t.Logf("Login failed with status code %d: %v", resp.StatusCode, errorResp)
		return generateTestToken(t)
	}

	var loginResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	require.NoError(t, err, "Failed to decode login response")

	// Check if the response contains the expected structure
	if loginResponse["success"] != true || loginResponse["data"] == nil {
		t.Logf("Unexpected login response format: %v", loginResponse)
		return generateTestToken(t)
	}

	// Extract token from response
	data, ok := loginResponse["data"].(map[string]interface{})
	if !ok {
		t.Logf("Data field is not an object: %v", loginResponse["data"])
		return generateTestToken(t)
	}

	token, ok := data["access_token"].(string)
	if !ok {
		t.Logf("Could not extract access_token from response: %v", data)
		return generateTestToken(t)
	}

	t.Logf("Successfully logged in and retrieved authentication token")
	return token
}

// ExtractJWTClaims extracts claims from a JWT token without verifying the signature
func ExtractJWTClaims(token string) (jwt.MapClaims, error) {
	// Parse the token without verification (just to extract claims)
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Skip verification for this function
		return []byte("dummy-key-for-extraction-only"), nil
	})

	if err != nil && !strings.Contains(err.Error(), "signature is invalid") {
		// If there's an error other than signature validation, return it
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("could not extract claims from token")
}

// IsTokenExpired checks if a JWT token is expired
func IsTokenExpired(token string) (bool, error) {
	claims, err := ExtractJWTClaims(token)
	if err != nil {
		return false, err
	}

	// Check for exp claim
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false, fmt.Errorf("token does not contain expiration claim")
	}

	// Compare with current time
	expiryTime := time.Unix(int64(exp), 0)
	return time.Now().After(expiryTime), nil
}

// GetAuthTokenWithCredentials obtains an authentication token for testing with custom credentials
func GetAuthTokenWithCredentials(t *testing.T, email, password string) string {
	// First, check if there's a test token in the environment
	testToken := os.Getenv("TEST_TOKEN")
	if testToken != "" {
		t.Log("Using TEST_TOKEN from environment variables")
		return testToken
	}

	// Try to get a real token by logging in
	loginURL := fmt.Sprintf("%s/auth/login", AuthBaseURL)

	t.Logf("Attempting login with email: %s", email)

	loginPayload := map[string]string{
		"email_or_phone": email,
		"password":       password,
	}

	jsonPayload, err := json.Marshal(loginPayload)
	require.NoError(t, err, "Failed to marshal login payload")

	// Create a client with a timeout
	client := CreateHTTPClient()

	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(jsonPayload))
	require.NoError(t, err, "Failed to create login request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Could not connect to auth service: %v", err)
		return generateTestToken(t)
	}
	defer resp.Body.Close()

	// If login fails
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errorResp)
		t.Logf("Login failed with status code %d: %v", resp.StatusCode, errorResp)
		return generateTestToken(t)
	}

	var loginResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	require.NoError(t, err, "Failed to decode login response")

	// Check if the response contains the expected structure
	if loginResponse["success"] != true || loginResponse["data"] == nil {
		t.Logf("Unexpected login response format: %v", loginResponse)
		return generateTestToken(t)
	}

	// Extract token from response
	data, ok := loginResponse["data"].(map[string]interface{})
	if !ok {
		t.Logf("Data field is not an object: %v", loginResponse["data"])
		return generateTestToken(t)
	}

	token, ok := data["access_token"].(string)
	if !ok {
		t.Logf("Could not extract access_token from response: %v", data)
		return generateTestToken(t)
	}

	t.Logf("Successfully logged in and retrieved authentication token")
	return token
}

// generateTestToken creates a test token for integration testing
func generateTestToken(t *testing.T) string {
	// Get the SESSION_KEY from environment variable
	secretKey := os.Getenv("SESSION_KEY")
	if secretKey == "" {
		t.Log("SESSION_KEY not set, using default development secret key")
		// This should match the server's default development key from .env
		secretKey = "development-session-key-change-this-in-production"
	}

	// Generate a token that matches the server's expected format
	now := time.Now()
	expiryTime := now.Add(24 * time.Hour)

	// Create token with all required claims
	claims := jwt.MapClaims{
		"user_id": "test_user_id",
		"email":   "test@example.com",
		"phone":   "08123456789",
		"name":    "Test User",
		"exp":     expiryTime.Unix(),
		"iat":     now.Unix(),
	}

	// Debug the claims
	t.Logf("Token claims: %+v", claims)

	// Use the HS256 algorithm as specified in auth_middleware.go
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		t.Logf("Failed to sign token: %v", err)
		t.Skip("Cannot create valid test token")
		return ""
	}

	// Log token details for debugging
	t.Logf("Generated token using secret key: %s", secretKey)
	t.Logf("Token expiry set to: %v", expiryTime)

	// Display token segments
	parts := strings.Split(tokenString, ".")
	if len(parts) == 3 {
		t.Logf("Token header: %s", parts[0])
		t.Logf("Token payload: %s", parts[1])
	}

	return tokenString
}

// GenerateTestTokenWithClaims creates a test token with custom claims for integration testing
func GenerateTestTokenWithClaims(t *testing.T, customClaims jwt.MapClaims) string {
	// Get the SESSION_KEY from environment variable
	secretKey := os.Getenv("SESSION_KEY")
	if secretKey == "" {
		t.Log("SESSION_KEY not set, using default development secret key")
		// This should match the server's default development key from .env
		secretKey = "development-session-key-change-this-in-production"
	}

	// Generate a token that matches the server's expected format
	now := time.Now()
	expiryTime := now.Add(24 * time.Hour)

	// Create base claims
	claims := jwt.MapClaims{
		"exp": expiryTime.Unix(),
		"iat": now.Unix(),
	}

	// Add custom claims, overriding defaults if provided
	for key, value := range customClaims {
		claims[key] = value
	}

	// Debug the claims
	t.Logf("Token claims: %+v", claims)

	// Use the HS256 algorithm as specified in auth_middleware.go
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		t.Logf("Failed to sign token: %v", err)
		t.Skip("Cannot create valid test token")
		return ""
	}

	// Log token details for debugging
	t.Logf("Generated token using secret key: %s", secretKey)
	t.Logf("Token expiry set to: %v", expiryTime)

	return tokenString
}

// CreateHTTPClient returns a properly configured HTTP client for integration tests
func CreateHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}
