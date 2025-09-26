package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/kirimku/smartseller-backend/pkg/cache"
)

// BaseService provides common functionality for all services
type BaseService struct {
	tenantResolver tenant.TenantResolver
	cache          cache.Cache
}

// NewBaseService creates a new base service
func NewBaseService(
	tenantResolver tenant.TenantResolver,
	cache cache.Cache,
	_ interface{}, // logger placeholder for compatibility
) *BaseService {
	return &BaseService{
		tenantResolver: tenantResolver,
		cache:          cache,
	}
}

// ServiceContext holds context information for service operations
type ServiceContext struct {
	Context      context.Context
	StorefrontID uuid.UUID
	TenantType   tenant.TenantType
	UserID       *uuid.UUID
	AdminAccess  bool
	RequestID    string
	ClientIP     string
	UserAgent    string
	Timestamp    time.Time
}

// NewServiceContext creates a new service context from HTTP context
func (bs *BaseService) NewServiceContext(ctx context.Context) (*ServiceContext, error) {
	// Extract tenant information
	storefrontID, err := bs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, errors.NewBusinessError(
			errors.ErrInvalidTenant,
			"Failed to resolve tenant context",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}

	// Get tenant type
	tenantType := bs.getTenantTypeFromContext(ctx)

	// Extract user information
	userID := bs.getUserFromContext(ctx)
	adminAccess := bs.hasAdminAccess(ctx)

	// Extract request metadata
	requestID := bs.getRequestID(ctx)
	clientIP := bs.getClientIP(ctx)
	userAgent := bs.getUserAgent(ctx)

	return &ServiceContext{
		Context:      ctx,
		StorefrontID: storefrontID,
		TenantType:   tenantType,
		UserID:       userID,
		AdminAccess:  adminAccess,
		RequestID:    requestID,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
		Timestamp:    time.Now(),
	}, nil
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
	Code    string      `json:"code"`
}

// ValidationResult holds validation results
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(field, message, code string, value interface{}) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
		Code:    code,
	})
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return !vr.Valid
}

// GetError returns the first error as a business error
func (vr *ValidationResult) GetError() error {
	if !vr.HasErrors() {
		return nil
	}

	return errors.NewValidationError(
		"Validation failed",
		vr.Errors,
	)
}

// CacheKey generates a cache key with tenant isolation
func (bs *BaseService) CacheKey(storefrontID uuid.UUID, keyType, identifier string) string {
	return fmt.Sprintf("tenant:%s:%s:%s", storefrontID.String(), keyType, identifier)
}

// LogOperation logs service operations with context
func (bs *BaseService) LogOperation(
	serviceCtx *ServiceContext,
	operation string,
	level logger.Level,
	message string,
	fields map[string]interface{},
) {
	if bs.logger == nil {
		log.Printf("[%s] %s: %s", level, operation, message)
		return
	}

	// Add context fields
	logFields := map[string]interface{}{
		"operation":     operation,
		"storefront_id": serviceCtx.StorefrontID,
		"tenant_type":   serviceCtx.TenantType,
		"request_id":    serviceCtx.RequestID,
		"client_ip":     serviceCtx.ClientIP,
		"timestamp":     serviceCtx.Timestamp,
	}

	if serviceCtx.UserID != nil {
		logFields["user_id"] = *serviceCtx.UserID
	}

	// Merge additional fields
	for k, v := range fields {
		logFields[k] = v
	}

	switch level {
	case logger.DEBUG:
		bs.logger.Debug(message, logFields)
	case logger.INFO:
		bs.logger.Info(message, logFields)
	case logger.WARN:
		bs.logger.Warn(message, logFields)
	case logger.ERROR:
		bs.logger.Error(message, logFields)
	default:
		bs.logger.Info(message, logFields)
	}
}

// HandleServiceError handles and logs service errors
func (bs *BaseService) HandleServiceError(
	serviceCtx *ServiceContext,
	operation string,
	err error,
	additionalFields map[string]interface{},
) error {
	// Log the error
	logFields := map[string]interface{}{
		"error": err.Error(),
	}

	// Add additional fields
	for k, v := range additionalFields {
		logFields[k] = v
	}

	bs.LogOperation(serviceCtx, operation, logger.ERROR, "Service operation failed", logFields)

	// Return the error (could be wrapped or transformed)
	return err
}

// ValidateRequired validates required fields
func (bs *BaseService) ValidateRequired(result *ValidationResult, field string, value interface{}) {
	switch v := value.(type) {
	case string:
		if v == "" {
			result.AddError(field, "Field is required", "required", value)
		}
	case *string:
		if v == nil || *v == "" {
			result.AddError(field, "Field is required", "required", value)
		}
	case uuid.UUID:
		if v == uuid.Nil {
			result.AddError(field, "Field is required", "required", value)
		}
	case *uuid.UUID:
		if v == nil || *v == uuid.Nil {
			result.AddError(field, "Field is required", "required", value)
		}
	case int, int32, int64:
		// Numbers are considered valid if not zero for required fields
		// You might want to adjust this logic based on your needs
	case *int, *int32, *int64:
		if v == nil {
			result.AddError(field, "Field is required", "required", value)
		}
	default:
		if value == nil {
			result.AddError(field, "Field is required", "required", value)
		}
	}
}

// ValidateEmail validates email format
func (bs *BaseService) ValidateEmail(result *ValidationResult, field, email string) {
	if email == "" {
		return // Skip validation for empty emails (use ValidateRequired if needed)
	}

	// Simple email validation - in production you might want more sophisticated validation
	if len(email) < 5 || !containsAt(email) || !containsDot(email) {
		result.AddError(field, "Invalid email format", "invalid_email", email)
	}
}

// ValidateLength validates string length
func (bs *BaseService) ValidateLength(result *ValidationResult, field, value string, min, max int) {
	if len(value) < min {
		result.AddError(field, fmt.Sprintf("Field must be at least %d characters long", min), "min_length", value)
	}
	if max > 0 && len(value) > max {
		result.AddError(field, fmt.Sprintf("Field must be at most %d characters long", max), "max_length", value)
	}
}

// ValidateUUID validates UUID format
func (bs *BaseService) ValidateUUID(result *ValidationResult, field, value string) {
	if value == "" {
		return // Skip validation for empty UUIDs
	}

	if _, err := uuid.Parse(value); err != nil {
		result.AddError(field, "Invalid UUID format", "invalid_uuid", value)
	}
}

// ValidateEnum validates enum values
func (bs *BaseService) ValidateEnum(result *ValidationResult, field, value string, validValues []string) {
	if value == "" {
		return // Skip validation for empty values
	}

	for _, valid := range validValues {
		if value == valid {
			return
		}
	}

	result.AddError(field, fmt.Sprintf("Invalid value. Allowed values: %v", validValues), "invalid_enum", value)
}

// CheckTenantAccess validates that the current context has access to the specified tenant
func (bs *BaseService) CheckTenantAccess(serviceCtx *ServiceContext, targetStorefrontID uuid.UUID) error {
	// Admin users have access to all tenants
	if serviceCtx.AdminAccess {
		return nil
	}

	// Check if the current context matches the target tenant
	if serviceCtx.StorefrontID != targetStorefrontID {
		return errors.NewAuthorizationError(
			"Insufficient privileges to access this tenant",
			map[string]interface{}{
				"current_tenant": serviceCtx.StorefrontID,
				"target_tenant":  targetStorefrontID,
			},
		)
	}

	return nil
}

// GetFromCache retrieves a value from cache with tenant isolation
func (bs *BaseService) GetFromCache(storefrontID uuid.UUID, keyType, identifier string, dest interface{}) error {
	if bs.cache == nil {
		return errors.NewInternalError("Cache not available", nil)
	}

	key := bs.CacheKey(storefrontID, keyType, identifier)
	return bs.cache.Get(key, dest)
}

// SetInCache stores a value in cache with tenant isolation and TTL
func (bs *BaseService) SetInCache(storefrontID uuid.UUID, keyType, identifier string, value interface{}, ttl time.Duration) error {
	if bs.cache == nil {
		return nil // Silently ignore if cache is not available
	}

	key := bs.CacheKey(storefrontID, keyType, identifier)
	return bs.cache.Set(key, value, ttl)
}

// DeleteFromCache removes a value from cache
func (bs *BaseService) DeleteFromCache(storefrontID uuid.UUID, keyType, identifier string) error {
	if bs.cache == nil {
		return nil // Silently ignore if cache is not available
	}

	key := bs.CacheKey(storefrontID, keyType, identifier)
	return bs.cache.Delete(key)
}

// Private helper methods

func (bs *BaseService) getStorefrontFromContext(ctx context.Context) (uuid.UUID, error) {
	if storefrontID, ok := ctx.Value("storefront_id").(uuid.UUID); ok {
		return storefrontID, nil
	}
	return uuid.Nil, fmt.Errorf("storefront ID not found in context")
}

func (bs *BaseService) getTenantTypeFromContext(ctx context.Context) tenant.TenantType {
	if tenantType, ok := ctx.Value("tenant_type").(tenant.TenantType); ok {
		return tenantType
	}
	return tenant.TenantTypeShared
}

func (bs *BaseService) getUserFromContext(ctx context.Context) *uuid.UUID {
	if userID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		return &userID
	}
	return nil
}

func (bs *BaseService) hasAdminAccess(ctx context.Context) bool {
	if adminAccess, ok := ctx.Value("admin_access").(bool); ok {
		return adminAccess
	}
	return false
}

func (bs *BaseService) getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

func (bs *BaseService) getClientIP(ctx context.Context) string {
	if clientIP, ok := ctx.Value("client_ip").(string); ok {
		return clientIP
	}
	return ""
}

func (bs *BaseService) getUserAgent(ctx context.Context) string {
	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		return userAgent
	}
	return ""
}

// Simple email validation helpers
func containsAt(s string) bool {
	for _, char := range s {
		if char == '@' {
			return true
		}
	}
	return false
}

func containsDot(s string) bool {
	for _, char := range s {
		if char == '.' {
			return true
		}
	}
	return false
}
