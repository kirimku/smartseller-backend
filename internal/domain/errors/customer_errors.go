package errors

import (
	"fmt"
	"net/http"
)

// DomainError represents a domain-specific error with code and message
type DomainError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string, httpStatus int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Customer-related errors
var (
	ErrCustomerNotFound         = NewDomainError("CUSTOMER_NOT_FOUND", "Customer not found", http.StatusNotFound)
	ErrEmailAlreadyExists       = NewDomainError("EMAIL_ALREADY_EXISTS", "Email address is already registered", http.StatusConflict)
	ErrPhoneAlreadyExists       = NewDomainError("PHONE_ALREADY_EXISTS", "Phone number is already registered", http.StatusConflict)
	ErrInvalidCredentials       = NewDomainError("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrAccountSuspended         = NewDomainError("ACCOUNT_SUSPENDED", "Customer account is suspended", http.StatusForbidden)
	ErrAccountLocked            = NewDomainError("ACCOUNT_LOCKED", "Account is temporarily locked due to too many failed login attempts", http.StatusForbidden)
	ErrEmailNotVerified         = NewDomainError("EMAIL_NOT_VERIFIED", "Email address is not verified", http.StatusForbidden)
	ErrPhoneNotVerified         = NewDomainError("PHONE_NOT_VERIFIED", "Phone number is not verified", http.StatusForbidden)
	ErrInvalidEmailToken        = NewDomainError("INVALID_EMAIL_TOKEN", "Invalid or expired email verification token", http.StatusBadRequest)
	ErrInvalidPhoneToken        = NewDomainError("INVALID_PHONE_TOKEN", "Invalid or expired phone verification token", http.StatusBadRequest)
	ErrInvalidPasswordToken     = NewDomainError("INVALID_PASSWORD_TOKEN", "Invalid or expired password reset token", http.StatusBadRequest)
	ErrWeakPassword            = NewDomainError("WEAK_PASSWORD", "Password does not meet security requirements", http.StatusBadRequest)
	ErrInvalidRefreshToken     = NewDomainError("INVALID_REFRESH_TOKEN", "Invalid or expired refresh token", http.StatusUnauthorized)
	ErrSessionExpired          = NewDomainError("SESSION_EXPIRED", "Session has expired", http.StatusUnauthorized)
	ErrSessionNotFound         = NewDomainError("SESSION_NOT_FOUND", "Session not found", http.StatusUnauthorized)
	ErrTooManyFailedAttempts   = NewDomainError("TOO_MANY_FAILED_ATTEMPTS", "Too many failed login attempts", http.StatusTooManyRequests)
	ErrInvalidCustomerStatus   = NewDomainError("INVALID_CUSTOMER_STATUS", "Invalid customer status", http.StatusBadRequest)
	ErrInvalidCustomerType     = NewDomainError("INVALID_CUSTOMER_TYPE", "Invalid customer type", http.StatusBadRequest)
	ErrCustomerInactive        = NewDomainError("CUSTOMER_INACTIVE", "Customer account is inactive", http.StatusForbidden)
	ErrCustomerBlocked         = NewDomainError("CUSTOMER_BLOCKED", "Customer account is blocked", http.StatusForbidden)
	ErrInvalidEmailFormat      = NewDomainError("INVALID_EMAIL_FORMAT", "Invalid email format", http.StatusBadRequest)
	ErrInvalidPhoneFormat      = NewDomainError("INVALID_PHONE_FORMAT", "Invalid phone number format", http.StatusBadRequest)
	ErrCustomerDataRequired    = NewDomainError("CUSTOMER_DATA_REQUIRED", "Customer data is required", http.StatusBadRequest)
	ErrDuplicateCustomerData   = NewDomainError("DUPLICATE_CUSTOMER_DATA", "Customer data already exists", http.StatusConflict)
)

// Storefront-related errors
var (
	ErrStorefrontNotFound       = NewDomainError("STOREFRONT_NOT_FOUND", "Storefront not found", http.StatusNotFound)
	ErrStorefrontInactive       = NewDomainError("STOREFRONT_INACTIVE", "Storefront is not active", http.StatusForbidden)
	ErrStorefrontSuspended      = NewDomainError("STOREFRONT_SUSPENDED", "Storefront is suspended", http.StatusForbidden)
	ErrSlugAlreadyExists        = NewDomainError("SLUG_ALREADY_EXISTS", "Storefront slug already exists", http.StatusConflict)
	ErrDomainAlreadyExists      = NewDomainError("DOMAIN_ALREADY_EXISTS", "Domain already exists", http.StatusConflict)
	ErrSubdomainAlreadyExists   = NewDomainError("SUBDOMAIN_ALREADY_EXISTS", "Subdomain already exists", http.StatusConflict)
	ErrUnauthorizedStorefront   = NewDomainError("UNAUTHORIZED_STOREFRONT", "Not authorized to access this storefront", http.StatusForbidden)
	ErrInvalidStorefrontSlug    = NewDomainError("INVALID_STOREFRONT_SLUG", "Invalid storefront slug format", http.StatusBadRequest)
	ErrInvalidStorefrontDomain  = NewDomainError("INVALID_STOREFRONT_DOMAIN", "Invalid domain format", http.StatusBadRequest)
	ErrInvalidStorefrontStatus  = NewDomainError("INVALID_STOREFRONT_STATUS", "Invalid storefront status", http.StatusBadRequest)
	ErrStorefrontNameRequired   = NewDomainError("STOREFRONT_NAME_REQUIRED", "Storefront name is required", http.StatusBadRequest)
	ErrStorefrontOwnerMismatch  = NewDomainError("STOREFRONT_OWNER_MISMATCH", "Storefront owner mismatch", http.StatusForbidden)
)

// Address-related errors
var (
	ErrAddressNotFound         = NewDomainError("ADDRESS_NOT_FOUND", "Address not found", http.StatusNotFound)
	ErrInvalidAddressType      = NewDomainError("INVALID_ADDRESS_TYPE", "Invalid address type", http.StatusBadRequest)
	ErrCannotDeleteDefaultAddr = NewDomainError("CANNOT_DELETE_DEFAULT_ADDRESS", "Cannot delete default address", http.StatusBadRequest)
	ErrDuplicateAddress        = NewDomainError("DUPLICATE_ADDRESS", "Address already exists for this customer", http.StatusConflict)
	ErrAddressRequired         = NewDomainError("ADDRESS_REQUIRED", "Address information is required", http.StatusBadRequest)
	ErrInvalidAddressData      = NewDomainError("INVALID_ADDRESS_DATA", "Invalid address data", http.StatusBadRequest)
	ErrAddressNotActive        = NewDomainError("ADDRESS_NOT_ACTIVE", "Address is not active", http.StatusBadRequest)
	ErrInvalidCoordinates      = NewDomainError("INVALID_COORDINATES", "Invalid latitude/longitude coordinates", http.StatusBadRequest)
	ErrAddressValidationFailed = NewDomainError("ADDRESS_VALIDATION_FAILED", "Address validation failed", http.StatusBadRequest)
)

// Multi-tenancy and tenant-related errors
var (
	ErrTenantMismatch          = NewDomainError("TENANT_MISMATCH", "Resource does not belong to this tenant", http.StatusForbidden)
	ErrTenantNotFound          = NewDomainError("TENANT_NOT_FOUND", "Tenant not found", http.StatusNotFound)
	ErrTenantAccessDenied      = NewDomainError("TENANT_ACCESS_DENIED", "Access denied for this tenant", http.StatusForbidden)
	ErrInvalidTenantContext    = NewDomainError("INVALID_TENANT_CONTEXT", "Invalid tenant context", http.StatusBadRequest)
	ErrTenantLimitExceeded     = NewDomainError("TENANT_LIMIT_EXCEEDED", "Tenant limit exceeded", http.StatusPaymentRequired)
	ErrTenantStorageExceeded   = NewDomainError("TENANT_STORAGE_EXCEEDED", "Tenant storage limit exceeded", http.StatusPaymentRequired)
	ErrTenantFeatureDisabled   = NewDomainError("TENANT_FEATURE_DISABLED", "Feature is disabled for this tenant", http.StatusForbidden)
	ErrCrossTenantAccess       = NewDomainError("CROSS_TENANT_ACCESS", "Cross-tenant access is not allowed", http.StatusForbidden)
)

// Authentication and authorization errors
var (
	ErrUnauthorized            = NewDomainError("UNAUTHORIZED", "Unauthorized access", http.StatusUnauthorized)
	ErrInsufficientPermissions = NewDomainError("INSUFFICIENT_PERMISSIONS", "Insufficient permissions", http.StatusForbidden)
	ErrInvalidToken            = NewDomainError("INVALID_TOKEN", "Invalid authentication token", http.StatusUnauthorized)
	ErrTokenExpired            = NewDomainError("TOKEN_EXPIRED", "Authentication token expired", http.StatusUnauthorized)
	ErrInvalidAPIKey           = NewDomainError("INVALID_API_KEY", "Invalid API key", http.StatusUnauthorized)
	ErrRateLimitExceeded       = NewDomainError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded", http.StatusTooManyRequests)
	ErrIPBlocked               = NewDomainError("IP_BLOCKED", "IP address is blocked", http.StatusForbidden)
	ErrSuspiciousActivity      = NewDomainError("SUSPICIOUS_ACTIVITY", "Suspicious activity detected", http.StatusForbidden)
)

// Validation and business logic errors
var (
	ErrValidationFailed        = NewDomainError("VALIDATION_FAILED", "Validation failed", http.StatusBadRequest)
	ErrRequiredFieldMissing    = NewDomainError("REQUIRED_FIELD_MISSING", "Required field is missing", http.StatusBadRequest)
	ErrInvalidFieldValue       = NewDomainError("INVALID_FIELD_VALUE", "Invalid field value", http.StatusBadRequest)
	ErrInvalidDateRange        = NewDomainError("INVALID_DATE_RANGE", "Invalid date range", http.StatusBadRequest)
	ErrInvalidPaginationParams = NewDomainError("INVALID_PAGINATION_PARAMS", "Invalid pagination parameters", http.StatusBadRequest)
	ErrInvalidSortParams       = NewDomainError("INVALID_SORT_PARAMS", "Invalid sort parameters", http.StatusBadRequest)
	ErrInvalidSearchQuery      = NewDomainError("INVALID_SEARCH_QUERY", "Invalid search query", http.StatusBadRequest)
	ErrBusinessRuleViolation   = NewDomainError("BUSINESS_RULE_VIOLATION", "Business rule violation", http.StatusBadRequest)
	ErrOperationNotAllowed     = NewDomainError("OPERATION_NOT_ALLOWED", "Operation not allowed", http.StatusForbidden)
	ErrConflictingState        = NewDomainError("CONFLICTING_STATE", "Resource is in conflicting state", http.StatusConflict)
)

// System and infrastructure errors
var (
	ErrDatabaseConnection      = NewDomainError("DATABASE_CONNECTION_ERROR", "Database connection error", http.StatusInternalServerError)
	ErrDatabaseQuery           = NewDomainError("DATABASE_QUERY_ERROR", "Database query error", http.StatusInternalServerError)
	ErrExternalServiceUnavail  = NewDomainError("EXTERNAL_SERVICE_UNAVAILABLE", "External service unavailable", http.StatusServiceUnavailable)
	ErrCacheUnavailable        = NewDomainError("CACHE_UNAVAILABLE", "Cache service unavailable", http.StatusServiceUnavailable)
	ErrConfigurationError      = NewDomainError("CONFIGURATION_ERROR", "Configuration error", http.StatusInternalServerError)
	ErrInternalServerError     = NewDomainError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrServiceTemporarilyUnavail = NewDomainError("SERVICE_TEMPORARILY_UNAVAILABLE", "Service temporarily unavailable", http.StatusServiceUnavailable)
)

// Helper functions to check specific error types
func IsCustomerNotFound(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "CUSTOMER_NOT_FOUND"
	}
	return false
}

func IsStorefrontNotFound(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "STOREFRONT_NOT_FOUND"
	}
	return false
}

func IsAddressNotFound(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "ADDRESS_NOT_FOUND"
	}
	return false
}

func IsTenantMismatch(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "TENANT_MISMATCH"
	}
	return false
}

func IsUnauthorized(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "UNAUTHORIZED" || domainErr.Code == "INVALID_CREDENTIALS"
	}
	return false
}

func IsValidationError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == "VALIDATION_FAILED" || 
			   domainErr.Code == "REQUIRED_FIELD_MISSING" ||
			   domainErr.Code == "INVALID_FIELD_VALUE"
	}
	return false
}

func IsConflictError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.HTTPStatus == http.StatusConflict
	}
	return false
}

func IsNotFoundError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.HTTPStatus == http.StatusNotFound
	}
	return false
}

func IsForbiddenError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.HTTPStatus == http.StatusForbidden
	}
	return false
}

func IsInternalServerError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.HTTPStatus == http.StatusInternalServerError
	}
	return false
}

// GetHTTPStatus returns the HTTP status code for the error
func GetHTTPStatus(err error) int {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// WrapError wraps a generic error with a domain error
func WrapError(baseErr *DomainError, message string) *DomainError {
	return &DomainError{
		Code:       baseErr.Code,
		Message:    message,
		HTTPStatus: baseErr.HTTPStatus,
	}
}

// WithCustomMessage creates a new domain error with a custom message
func WithCustomMessage(baseErr *DomainError, message string) *DomainError {
	return &DomainError{
		Code:       baseErr.Code,
		Message:    message,
		HTTPStatus: baseErr.HTTPStatus,
	}
}