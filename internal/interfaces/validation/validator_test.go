package validation

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestValidateAndSanitize(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   string
		expectedError bool
		expectedName  string
		expectedEmail string
	}{
		{
			name:          "valid input",
			requestBody:   `{"name": "John Doe", "email": "john@example.com", "refresh_token": "valid-token-123"}`,
			expectedError: false,
			expectedName:  "John Doe",
			expectedEmail: "john@example.com",
		},
		{
			name:          "invalid email",
			requestBody:   `{"name": "John Doe", "email": "invalid-email", "refresh_token": "valid-token-123"}`,
			expectedError: true,
		},
		{
			name:          "name too short",
			requestBody:   `{"name": "J", "email": "john@example.com", "refresh_token": "valid-token-123"}`,
			expectedError: true,
		},
		{
			name:          "missing required refresh token",
			requestBody:   `{"name": "John Doe", "email": "john@example.com"}`,
			expectedError: true,
		},
		{
			name:          "sanitize HTML in input",
			requestBody:   `{"name": "<script>alert('xss')</script>John", "email": "john@example.com", "refresh_token": "valid-token-123"}`,
			expectedError: false,
			expectedName:  "alert(xss)John",
			expectedEmail: "john@example.com",
		},
		{
			name:          "sanitize SQL injection attempt",
			requestBody:   `{"name": "John' OR '1'='1", "email": "john@example.com", "refresh_token": "valid-token-123"}`,
			expectedError: false,
			expectedName:  "John OR 1=1",
			expectedEmail: "john@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.requestBody))

			// Test validation
			var body RequestBody
			err := ValidateAndSanitize(req, &body)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// If we expect success, check sanitized values
			if !tt.expectedError {
				if body.Name != tt.expectedName {
					t.Errorf("Expected name %q, got %q", tt.expectedName, body.Name)
				}
				if body.Email != tt.expectedEmail {
					t.Errorf("Expected email %q, got %q", tt.expectedEmail, body.Email)
				}
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			input:    "<script>alert('xss')</script>",
			expected: "alert(xss)",
		},
		{
			input:    "Robert'; DROP TABLE Students;--",
			expected: "Robert DROP TABLE Students",
		},
		{
			input:    `"quoted"text'here';`,
			expected: "quotedtexthere",
		},
	}

	for _, tt := range tests {
		result := sanitizeString(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeString(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
