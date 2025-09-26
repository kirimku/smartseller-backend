package errors

import "fmt"

// ErrorType represents the type of error
type ErrorType string

const (
	// Error types
	ErrorTypeValidation    ErrorType = "VALIDATION_ERROR"
	ErrorTypeAuthorization ErrorType = "AUTHORIZATION_ERROR"
	ErrorTypeNotFound      ErrorType = "NOT_FOUND"
	ErrorTypeInternal      ErrorType = "INTERNAL_ERROR"
	ErrorTypeRateLimit     ErrorType = "RATE_LIMIT_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
	Code    int       `json:"-"` // HTTP status code
	Err     error     `json:"-"` // Original error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    400,
		Err:     err,
	}
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeAuthorization,
		Message: message,
		Code:    401,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Code:    404,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    500,
		Err:     err,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError() *AppError {
	return &AppError{
		Type:    ErrorTypeRateLimit,
		Message: "Rate limit exceeded",
		Code:    429,
	}
}

// NewBusinessError creates a new business logic error
func NewBusinessError(baseErr error, message string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    500,
		Err:     baseErr,
	}
}

// Common error variables
var (
	ErrInvalidTenant = fmt.Errorf("invalid tenant context")
)
