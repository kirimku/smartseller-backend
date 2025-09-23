package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestVerifyTokenClaims parses the test token and outputs its claims
func TestVerifyTokenClaims(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping token verification test in short mode")
	}

	// Get token from environment
	token := os.Getenv("TEST_TOKEN")
	if token == "" {
		t.Log("TEST_TOKEN not found in environment, using auth helper")
		token = GetAuthToken(t)
	}
	require.NotEmpty(t, token, "Could not get a valid test token")

	t.Logf("Verifying token: %s", token)

	// Use new utility function to extract claims
	claims, err := ExtractJWTClaims(token)
	require.NoError(t, err, "Failed to extract token claims")

	// Print all claims
	t.Log("Token claims:")
	for key, value := range claims {
		t.Logf("  %s: %v", key, value)
	}

	// Check specifically for user_id
	if userID, exists := claims["user_id"]; exists {
		t.Logf("✓ Found user_id claim: %v", userID)
	} else {
		t.Logf("✗ user_id claim not found in token")
	}

	// Check for roles or permissions
	if roles, exists := claims["roles"]; exists {
		t.Logf("✓ Found roles claim: %v", roles)
	} else {
		t.Logf("Roles claim not found in token")
	}

	if permissions, exists := claims["permissions"]; exists {
		t.Logf("✓ Found permissions claim: %v", permissions)
	} else {
		t.Logf("Permissions claim not found in token")
	}

	// Check for expiration
	if exp, exists := claims["exp"]; exists {
		t.Logf("✓ Token expires at: %v", exp)
	} else {
		t.Logf("✗ Token has no expiration claim")
	}

	// Check if token is expired
	isExpired, err := IsTokenExpired(token)
	require.NoError(t, err, "Failed to check token expiration")
	if isExpired {
		t.Logf("✗ Token is expired")
	} else {
		t.Logf("✓ Token is valid and not expired")
	}
}
