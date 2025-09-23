package errors

import (
	"errors"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	msg := "invalid input"
	err := errors.New("validation failed")
	appErr := NewValidationError(msg, err)

	if appErr.Type != ErrorTypeValidation {
		t.Errorf("Expected error type %v, got %v", ErrorTypeValidation, appErr.Type)
	}
	if appErr.Code != 400 {
		t.Errorf("Expected status code 400, got %d", appErr.Code)
	}
	if appErr.Message != msg {
		t.Errorf("Expected message %s, got %s", msg, appErr.Message)
	}
}

func TestNewAuthorizationError(t *testing.T) {
	msg := "unauthorized access"
	appErr := NewAuthorizationError(msg)

	if appErr.Type != ErrorTypeAuthorization {
		t.Errorf("Expected error type %v, got %v", ErrorTypeAuthorization, appErr.Type)
	}
	if appErr.Code != 401 {
		t.Errorf("Expected status code 401, got %d", appErr.Code)
	}
	if appErr.Message != msg {
		t.Errorf("Expected message %s, got %s", msg, appErr.Message)
	}
}

func TestNewNotFoundError(t *testing.T) {
	msg := "resource not found"
	appErr := NewNotFoundError(msg)

	if appErr.Type != ErrorTypeNotFound {
		t.Errorf("Expected error type %v, got %v", ErrorTypeNotFound, appErr.Type)
	}
	if appErr.Code != 404 {
		t.Errorf("Expected status code 404, got %d", appErr.Code)
	}
	if appErr.Message != msg {
		t.Errorf("Expected message %s, got %s", msg, appErr.Message)
	}
}

func TestNewInternalError(t *testing.T) {
	msg := "internal server error"
	err := errors.New("database connection failed")
	appErr := NewInternalError(msg, err)

	if appErr.Type != ErrorTypeInternal {
		t.Errorf("Expected error type %v, got %v", ErrorTypeInternal, appErr.Type)
	}
	if appErr.Code != 500 {
		t.Errorf("Expected status code 500, got %d", appErr.Code)
	}
	if appErr.Message != msg {
		t.Errorf("Expected message %s, got %s", msg, appErr.Message)
	}
}

func TestNewRateLimitError(t *testing.T) {
	appErr := NewRateLimitError()

	if appErr.Type != ErrorTypeRateLimit {
		t.Errorf("Expected error type %v, got %v", ErrorTypeRateLimit, appErr.Type)
	}
	if appErr.Code != 429 {
		t.Errorf("Expected status code 429, got %d", appErr.Code)
	}
	if appErr.Message != "Rate limit exceeded" {
		t.Errorf("Expected message 'Rate limit exceeded', got %s", appErr.Message)
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "with underlying error",
			appErr: &AppError{
				Type:    ErrorTypeValidation,
				Message: "invalid input",
				Err:     errors.New("validation failed"),
			},
			expected: "VALIDATION_ERROR: invalid input (validation failed)",
		},
		{
			name: "without underlying error",
			appErr: &AppError{
				Type:    ErrorTypeNotFound,
				Message: "resource not found",
			},
			expected: "NOT_FOUND: resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.Error(); got != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}
