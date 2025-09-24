package usecase

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// ErrorCode represents a unique error code for business logic errors
type ErrorCode string

const (
	// Product errors
	ErrCodeProductNotFound      ErrorCode = "PRODUCT_NOT_FOUND"
	ErrCodeProductAlreadyExists ErrorCode = "PRODUCT_ALREADY_EXISTS"
	ErrCodeProductInvalidSKU    ErrorCode = "PRODUCT_INVALID_SKU"
	ErrCodeProductInvalidPrice  ErrorCode = "PRODUCT_INVALID_PRICE"
	ErrCodeProductInvalidStatus ErrorCode = "PRODUCT_INVALID_STATUS"
	ErrCodeProductInvalidStock  ErrorCode = "PRODUCT_INVALID_STOCK"

	// Category errors
	ErrCodeCategoryNotFound      ErrorCode = "CATEGORY_NOT_FOUND"
	ErrCodeCategoryAlreadyExists ErrorCode = "CATEGORY_ALREADY_EXISTS"
	ErrCodeCategoryInvalidPath   ErrorCode = "CATEGORY_INVALID_PATH"
	ErrCodeCategoryHasProducts   ErrorCode = "CATEGORY_HAS_PRODUCTS"
	ErrCodeCategoryHasChildren   ErrorCode = "CATEGORY_HAS_CHILDREN"
	ErrCodeCategoryCircularRef   ErrorCode = "CATEGORY_CIRCULAR_REFERENCE"

	// Variant errors
	ErrCodeVariantNotFound      ErrorCode = "VARIANT_NOT_FOUND"
	ErrCodeVariantAlreadyExists ErrorCode = "VARIANT_ALREADY_EXISTS"
	ErrCodeVariantInvalidSKU    ErrorCode = "VARIANT_INVALID_SKU"
	ErrCodeVariantInvalidPrice  ErrorCode = "VARIANT_INVALID_PRICE"
	ErrCodeVariantInvalidStock  ErrorCode = "VARIANT_INVALID_STOCK"
	ErrCodeVariantOptionExists  ErrorCode = "VARIANT_OPTION_EXISTS"
	ErrCodeVariantOptionInvalid ErrorCode = "VARIANT_OPTION_INVALID"

	// Image errors
	ErrCodeImageNotFound      ErrorCode = "IMAGE_NOT_FOUND"
	ErrCodeImageInvalidURL    ErrorCode = "IMAGE_INVALID_URL"
	ErrCodeImageTooLarge      ErrorCode = "IMAGE_TOO_LARGE"
	ErrCodeImageLimitExceeded ErrorCode = "IMAGE_LIMIT_EXCEEDED"
	ErrCodeImagePrimaryExists ErrorCode = "IMAGE_PRIMARY_EXISTS"
	ErrCodeImageInvalidFormat ErrorCode = "IMAGE_INVALID_FORMAT"

	// Business logic errors
	ErrCodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"
	ErrCodeInsufficientStock     ErrorCode = "INSUFFICIENT_STOCK"
	ErrCodeInvalidOperation      ErrorCode = "INVALID_OPERATION"
	ErrCodeDataConsistency       ErrorCode = "DATA_CONSISTENCY_ERROR"
	ErrCodeValidationFailed      ErrorCode = "VALIDATION_FAILED"

	// Authorization errors
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"

	// System errors
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrCodeTimeout       ErrorCode = "TIMEOUT"
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"
)

// UseCaseError represents a standardized error from the use case layer
type UseCaseError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Cause       error                  `json:"-"`
	HTTPStatus  int                    `json:"-"`
	UserMessage string                 `json:"user_message,omitempty"`
}

// Error implements the error interface
func (e *UseCaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *UseCaseError) Unwrap() error {
	return e.Cause
}

// WithDetail adds a detail to the error
func (e *UseCaseError) WithDetail(key string, value interface{}) *UseCaseError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithCause sets the underlying cause of the error
func (e *UseCaseError) WithCause(cause error) *UseCaseError {
	e.Cause = cause
	return e
}

// WithUserMessage sets a user-friendly message
func (e *UseCaseError) WithUserMessage(message string) *UseCaseError {
	e.UserMessage = message
	return e
}

// NewUseCaseError creates a new use case error
func NewUseCaseError(code ErrorCode, message string) *UseCaseError {
	return &UseCaseError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatusForErrorCode(code),
	}
}

// getHTTPStatusForErrorCode maps error codes to HTTP status codes
func getHTTPStatusForErrorCode(code ErrorCode) int {
	switch code {
	case ErrCodeProductNotFound, ErrCodeCategoryNotFound, ErrCodeVariantNotFound, ErrCodeImageNotFound:
		return http.StatusNotFound
	case ErrCodeProductAlreadyExists, ErrCodeCategoryAlreadyExists, ErrCodeVariantAlreadyExists:
		return http.StatusConflict
	case ErrCodeProductInvalidSKU, ErrCodeProductInvalidPrice, ErrCodeProductInvalidStatus, ErrCodeProductInvalidStock,
		ErrCodeCategoryInvalidPath, ErrCodeVariantInvalidSKU, ErrCodeVariantInvalidPrice, ErrCodeVariantInvalidStock,
		ErrCodeImageInvalidURL, ErrCodeImageInvalidFormat, ErrCodeValidationFailed:
		return http.StatusBadRequest
	case ErrCodeImageTooLarge, ErrCodeImageLimitExceeded:
		return http.StatusRequestEntityTooLarge
	case ErrCodeBusinessRuleViolation, ErrCodeInsufficientStock, ErrCodeInvalidOperation:
		return http.StatusUnprocessableEntity
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	case ErrCodeInternalError, ErrCodeDatabaseError, ErrCodeDataConsistency:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Predefined error constructors for common use cases

// Product errors
func ErrProductNotFound(productID uuid.UUID) *UseCaseError {
	return NewUseCaseError(ErrCodeProductNotFound, "Product not found").
		WithDetail("product_id", productID.String()).
		WithUserMessage("The requested product could not be found")
}

func ErrProductAlreadyExists(sku string) *UseCaseError {
	return NewUseCaseError(ErrCodeProductAlreadyExists, "Product with SKU already exists").
		WithDetail("sku", sku).
		WithUserMessage("A product with this SKU already exists")
}

func ErrProductInvalidSKU(sku string, reason string) *UseCaseError {
	return NewUseCaseError(ErrCodeProductInvalidSKU, "Invalid product SKU").
		WithDetail("sku", sku).
		WithDetail("reason", reason).
		WithUserMessage("The product SKU format is invalid")
}

func ErrProductInvalidPrice(price interface{}, reason string) *UseCaseError {
	return NewUseCaseError(ErrCodeProductInvalidPrice, "Invalid product price").
		WithDetail("price", price).
		WithDetail("reason", reason).
		WithUserMessage("The product price is invalid")
}

func ErrProductInvalidStock(stock int, reason string) *UseCaseError {
	return NewUseCaseError(ErrCodeProductInvalidStock, "Invalid stock quantity").
		WithDetail("stock", stock).
		WithDetail("reason", reason).
		WithUserMessage("The stock quantity is invalid")
}

// Category errors
func ErrCategoryNotFound(categoryID uuid.UUID) *UseCaseError {
	return NewUseCaseError(ErrCodeCategoryNotFound, "Category not found").
		WithDetail("category_id", categoryID.String()).
		WithUserMessage("The requested category could not be found")
}

func ErrCategoryAlreadyExists(name string) *UseCaseError {
	return NewUseCaseError(ErrCodeCategoryAlreadyExists, "Category already exists").
		WithDetail("name", name).
		WithUserMessage("A category with this name already exists")
}

func ErrCategoryHasProducts(categoryID uuid.UUID, productCount int) *UseCaseError {
	return NewUseCaseError(ErrCodeCategoryHasProducts, "Category has associated products").
		WithDetail("category_id", categoryID.String()).
		WithDetail("product_count", productCount).
		WithUserMessage("Cannot delete category that has associated products")
}

func ErrCategoryCircularReference(categoryID uuid.UUID, parentID uuid.UUID) *UseCaseError {
	return NewUseCaseError(ErrCodeCategoryCircularRef, "Circular reference in category hierarchy").
		WithDetail("category_id", categoryID.String()).
		WithDetail("parent_id", parentID.String()).
		WithUserMessage("Cannot create circular reference in category hierarchy")
}

// Variant errors
func ErrVariantNotFound(variantID uuid.UUID) *UseCaseError {
	return NewUseCaseError(ErrCodeVariantNotFound, "Product variant not found").
		WithDetail("variant_id", variantID.String()).
		WithUserMessage("The requested product variant could not be found")
}

func ErrVariantAlreadyExists(productID uuid.UUID, sku string) *UseCaseError {
	return NewUseCaseError(ErrCodeVariantAlreadyExists, "Variant with SKU already exists").
		WithDetail("product_id", productID.String()).
		WithDetail("sku", sku).
		WithUserMessage("A variant with this SKU already exists for this product")
}

func ErrVariantOptionInvalid(optionName string, reason string) *UseCaseError {
	return NewUseCaseError(ErrCodeVariantOptionInvalid, "Invalid variant option").
		WithDetail("option_name", optionName).
		WithDetail("reason", reason).
		WithUserMessage("The variant option configuration is invalid")
}

// Image errors
func ErrImageNotFound(imageID uuid.UUID) *UseCaseError {
	return NewUseCaseError(ErrCodeImageNotFound, "Product image not found").
		WithDetail("image_id", imageID.String()).
		WithUserMessage("The requested product image could not be found")
}

func ErrImageInvalidURL(url string, reason string) *UseCaseError {
	return NewUseCaseError(ErrCodeImageInvalidURL, "Invalid image URL").
		WithDetail("url", url).
		WithDetail("reason", reason).
		WithUserMessage("The image URL is invalid or cannot be accessed")
}

func ErrImageTooLarge(size int64, maxSize int64) *UseCaseError {
	return NewUseCaseError(ErrCodeImageTooLarge, "Image file too large").
		WithDetail("size", size).
		WithDetail("max_size", maxSize).
		WithUserMessage(fmt.Sprintf("Image file is too large. Maximum size allowed is %d bytes", maxSize))
}

func ErrImageLimitExceeded(current int, limit int) *UseCaseError {
	return NewUseCaseError(ErrCodeImageLimitExceeded, "Image limit exceeded").
		WithDetail("current", current).
		WithDetail("limit", limit).
		WithUserMessage(fmt.Sprintf("Maximum of %d images allowed per product", limit))
}

// Business logic errors
func ErrBusinessRuleViolation(rule string, details map[string]interface{}) *UseCaseError {
	err := NewUseCaseError(ErrCodeBusinessRuleViolation, fmt.Sprintf("Business rule violation: %s", rule)).
		WithUserMessage("Operation violates business rules")

	for k, v := range details {
		err.WithDetail(k, v)
	}

	return err
}

func ErrInsufficientStock(available int, required int) *UseCaseError {
	return NewUseCaseError(ErrCodeInsufficientStock, "Insufficient stock").
		WithDetail("available", available).
		WithDetail("required", required).
		WithUserMessage(fmt.Sprintf("Insufficient stock. Available: %d, Required: %d", available, required))
}

func ErrInvalidOperation(operation string, reason string) *UseCaseError {
	return NewUseCaseError(ErrCodeInvalidOperation, fmt.Sprintf("Invalid operation: %s", operation)).
		WithDetail("operation", operation).
		WithDetail("reason", reason).
		WithUserMessage("The requested operation cannot be performed")
}

func ErrDataConsistency(entity string, details map[string]interface{}) *UseCaseError {
	err := NewUseCaseError(ErrCodeDataConsistency, fmt.Sprintf("Data consistency error for %s", entity)).
		WithUserMessage("Data consistency violation detected")

	for k, v := range details {
		err.WithDetail(k, v)
	}

	return err
}

// Authorization errors
func ErrUnauthorized(action string) *UseCaseError {
	return NewUseCaseError(ErrCodeUnauthorized, "Unauthorized access").
		WithDetail("action", action).
		WithUserMessage("You are not authorized to perform this action")
}

func ErrForbidden(resource string, action string) *UseCaseError {
	return NewUseCaseError(ErrCodeForbidden, "Access forbidden").
		WithDetail("resource", resource).
		WithDetail("action", action).
		WithUserMessage("You do not have permission to access this resource")
}

// System errors
func ErrInternalError(component string, cause error) *UseCaseError {
	return NewUseCaseError(ErrCodeInternalError, "Internal system error").
		WithDetail("component", component).
		WithCause(cause).
		WithUserMessage("An internal error occurred. Please try again later")
}

func ErrDatabaseError(operation string, cause error) *UseCaseError {
	return NewUseCaseError(ErrCodeDatabaseError, "Database error").
		WithDetail("operation", operation).
		WithCause(cause).
		WithUserMessage("A database error occurred. Please try again later")
}

func ErrTimeout(operation string, duration string) *UseCaseError {
	return NewUseCaseError(ErrCodeTimeout, "Operation timeout").
		WithDetail("operation", operation).
		WithDetail("duration", duration).
		WithUserMessage("The operation timed out. Please try again")
}

// Validation errors
func ErrValidationFailed(field string, value interface{}, rule string) *UseCaseError {
	return NewUseCaseError(ErrCodeValidationFailed, "Validation failed").
		WithDetail("field", field).
		WithDetail("value", value).
		WithDetail("rule", rule).
		WithUserMessage(fmt.Sprintf("Validation failed for field '%s'", field))
}

// Error checking utilities

// IsUseCaseError checks if an error is a UseCaseError
func IsUseCaseError(err error) bool {
	var useCaseErr *UseCaseError
	return errors.As(err, &useCaseErr)
}

// AsUseCaseError converts an error to a UseCaseError if possible
func AsUseCaseError(err error) (*UseCaseError, bool) {
	var useCaseErr *UseCaseError
	ok := errors.As(err, &useCaseErr)
	return useCaseErr, ok
}

// HasErrorCode checks if an error has a specific error code
func HasErrorCode(err error, code ErrorCode) bool {
	if useCaseErr, ok := AsUseCaseError(err); ok {
		return useCaseErr.Code == code
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) (ErrorCode, bool) {
	if useCaseErr, ok := AsUseCaseError(err); ok {
		return useCaseErr.Code, true
	}
	return "", false
}

// GetHTTPStatus extracts the HTTP status code from an error
func GetHTTPStatus(err error) int {
	if useCaseErr, ok := AsUseCaseError(err); ok {
		return useCaseErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// WrapRepositoryError wraps a repository error as a use case error
func WrapRepositoryError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check for common repository error patterns
	errMsg := err.Error()

	// Not found errors
	if containsAny(errMsg, []string{"not found", "no rows", "record not found"}) {
		return NewUseCaseError(ErrCodeProductNotFound, "Resource not found").
			WithDetail("operation", operation).
			WithCause(err)
	}

	// Constraint violations
	if containsAny(errMsg, []string{"duplicate", "unique constraint", "already exists"}) {
		return NewUseCaseError(ErrCodeProductAlreadyExists, "Resource already exists").
			WithDetail("operation", operation).
			WithCause(err)
	}

	// Foreign key violations
	if containsAny(errMsg, []string{"foreign key", "violates foreign key constraint"}) {
		return NewUseCaseError(ErrCodeDataConsistency, "Data consistency violation").
			WithDetail("operation", operation).
			WithCause(err)
	}

	// Default to database error
	return ErrDatabaseError(operation, err)
}

// containsAny checks if a string contains any of the given substrings
func containsAny(str string, substrings []string) bool {
	for _, substring := range substrings {
		if len(str) >= len(substring) {
			for i := 0; i <= len(str)-len(substring); i++ {
				if str[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}
