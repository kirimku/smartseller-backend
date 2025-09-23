package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	parentpkg "github.com/kirimku/smartseller-backend/internal/tests/integration"
)

// TestLoginWithCustomCredentials demonstrates using custom credentials for login
func TestLoginWithCustomCredentials(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping custom login test in short mode")
	}

	// Use custom credentials
	token := parentpkg.GetAuthTokenWithCredentials(t, "test@example.com", "password123")
	assert.NotEmpty(t, token, "Failed to get token with custom credentials")

	// Extract and verify token claims
	claims, err := parentpkg.ExtractJWTClaims(token)
	assert.NoError(t, err, "Failed to extract claims from token")

	// Check for essential claims
	t.Logf("Token claims: %v", claims)
	assert.Contains(t, claims, "user_id", "Token should contain user_id claim")
	assert.Contains(t, claims, "exp", "Token should contain expiration claim")
	assert.Contains(t, claims, "iat", "Token should contain issued-at claim")

	// Check token expiration
	isExpired, err := parentpkg.IsTokenExpired(token)
	assert.NoError(t, err, "Failed to check token expiration")
	assert.False(t, isExpired, "Token should not be expired")

	t.Log("Successfully validated token obtained with custom credentials")
}
