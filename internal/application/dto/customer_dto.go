package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// Customer Registration and Authentication DTOs

// CustomerRegistrationRequest represents a customer registration request
type CustomerRegistrationRequest struct {
	Email       string         `json:"email" binding:"required,email"`
	Phone       *string        `json:"phone,omitempty"`
	FirstName   string         `json:"first_name" binding:"required,min=1,max=255"`
	LastName    string         `json:"last_name" binding:"required,min=1,max=255"`
	Password    string         `json:"password" binding:"required,min=8,max=100"`
	DateOfBirth *time.Time     `json:"date_of_birth,omitempty"`
	Gender      *entity.Gender `json:"gender,omitempty"`
}

// CustomerAuthRequest represents a customer authentication request
type CustomerAuthRequest struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password" binding:"required"`
}

// CustomerAuthResponse represents a customer authentication response
type CustomerAuthResponse struct {
	Customer     *CustomerResponse `json:"customer"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	TokenType    string            `json:"token_type"`
	ExpiresIn    int64             `json:"expires_in"`
}

// TokenRefreshRequest represents a token refresh request
type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CustomerID      uuid.UUID `json:"customer_id" binding:"required"`
	CurrentPassword string    `json:"current_password" binding:"required"`
	NewPassword     string    `json:"new_password" binding:"required,min=8,max=100"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest represents a password reset confirmation request
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=100"`
}

// Customer Profile and Management DTOs

// CustomerResponse represents a customer response
type CustomerResponse struct {
	ID           uuid.UUID                  `json:"id"`
	StorefrontID uuid.UUID                  `json:"storefront_id"`
	Email        *string                    `json:"email,omitempty"`
	Phone        *string                    `json:"phone,omitempty"`
	FirstName    *string                    `json:"first_name,omitempty"`
	LastName     *string                    `json:"last_name,omitempty"`
	FullName     *string                    `json:"full_name,omitempty"`
	DateOfBirth  *time.Time                 `json:"date_of_birth,omitempty"`
	Gender       *entity.Gender             `json:"gender,omitempty"`
	Status       string                     `json:"status"`
	CustomerType entity.CustomerType        `json:"customer_type"`
	Preferences  entity.CustomerPreferences `json:"preferences"`
	LastLoginAt  *time.Time                 `json:"last_login_at,omitempty"`
	CreatedAt    time.Time                  `json:"created_at"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

// CustomerUpdateRequest represents a customer profile update request
type CustomerUpdateRequest struct {
	Email       *string        `json:"email,omitempty" binding:"omitempty,email"`
	Phone       *string        `json:"phone,omitempty"`
	FirstName   *string        `json:"first_name,omitempty" binding:"omitempty,min=1,max=255"`
	LastName    *string        `json:"last_name,omitempty" binding:"omitempty,min=1,max=255"`
	DateOfBirth *time.Time     `json:"date_of_birth,omitempty"`
	Gender      *entity.Gender `json:"gender,omitempty"`
}

// Customer Search and Pagination DTOs

// CustomerSearchRequest represents a customer search request
type CustomerSearchRequest struct {
	Query         string                 `json:"query,omitempty"`
	Status        *entity.CustomerStatus `json:"status,omitempty"`
	CustomerType  *entity.CustomerType   `json:"customer_type,omitempty"`
	CreatedAfter  *time.Time             `json:"created_after,omitempty"`
	CreatedBefore *time.Time             `json:"created_before,omitempty"`
	Page          int                    `json:"page" binding:"min=1"`
	PageSize      int                    `json:"page_size" binding:"min=1,max=100"`
	SortBy        string                 `json:"sort_by,omitempty"`
	SortOrder     string                 `json:"sort_order,omitempty" binding:"omitempty,oneof=asc desc"`
}

// PaginatedCustomerResponse represents a paginated customer response
type PaginatedCustomerResponse struct {
	Customers  []*CustomerResponse `json:"customers"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int64               `json:"total_pages"`
}

// Customer Statistics and Analytics DTOs

// CustomerStatsRequest represents a customer statistics request
type CustomerStatsRequest struct {
	Period    string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// CustomerStatsResponse represents customer statistics
type CustomerStatsResponse struct {
	TotalCustomers    int64     `json:"total_customers"`
	ActiveCustomers   int64     `json:"active_customers"`
	InactiveCustomers int64     `json:"inactive_customers"`
	NewCustomers      int64     `json:"new_customers"`
	Period            string    `json:"period"`
	Timestamp         time.Time `json:"timestamp"`
}

// ActivityRequest represents an activity request
type ActivityRequest struct {
	Page     int        `json:"page" binding:"min=1"`
	PageSize int        `json:"page_size" binding:"min=1,max=100"`
	Since    *time.Time `json:"since,omitempty"`
}

// Activity represents an activity entry
type Activity struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// CustomerActivityResponse represents customer activity response
type CustomerActivityResponse struct {
	CustomerID uuid.UUID   `json:"customer_id"`
	Activities []*Activity `json:"activities"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

// Bulk Operations DTOs

// BulkCustomerUpdateRequest represents a bulk customer update request
type BulkCustomerUpdateRequest struct {
	CustomerIDs []uuid.UUID            `json:"customer_ids" binding:"required,min=1"`
	Updates     map[string]interface{} `json:"updates" binding:"required"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
}

// BulkOperationError represents an error in bulk operations
type BulkOperationError struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
	Code    string    `json:"code"`
}

// BulkOperationResponse represents a bulk operation response
type BulkOperationResponse struct {
	TotalRequested int                  `json:"total_requested"`
	Successful     int                  `json:"successful"`
	Failed         int                  `json:"failed"`
	Errors         []BulkOperationError `json:"errors,omitempty"`
}

// Export DTOs

// CustomerExportRequest represents a customer export request
type CustomerExportRequest struct {
	Format  string                 `json:"format" binding:"required,oneof=csv xlsx json"`
	Filters map[string]interface{} `json:"filters,omitempty"`
	Fields  []string               `json:"fields,omitempty"`
}

// Cache DTOs

// CacheStatsResponse represents cache statistics
type CacheStatsResponse struct {
	HitRate       float64          `json:"hit_rate"`
	TotalHits     int64            `json:"total_hits"`
	TotalMisses   int64            `json:"total_misses"`
	TotalKeys     int64            `json:"total_keys"`
	MemoryUsage   int64            `json:"memory_usage"`
	KeysByPattern map[string]int64 `json:"keys_by_pattern"`
	Timestamp     time.Time        `json:"timestamp"`
}

// CacheWarmupRequest represents a cache warmup request
type CacheWarmupRequest struct {
	Patterns    []string `json:"patterns" binding:"required"`
	Concurrency int      `json:"concurrency,omitempty" binding:"min=1,max=10"`
}

// Notification and Event DTOs

// SystemAlert represents a system alert
type SystemAlert struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID           uuid.UUID `json:"id"`
	Metric       string    `json:"metric"`
	Value        float64   `json:"value"`
	Threshold    float64   `json:"threshold"`
	StorefrontID uuid.UUID `json:"storefront_id,omitempty"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}

// BulkNotificationRequest represents a bulk notification request
type BulkNotificationRequest struct {
	Type         string                 `json:"type" binding:"required"`
	Recipients   []string               `json:"recipients" binding:"required,min=1"`
	Subject      string                 `json:"subject" binding:"required"`
	Content      string                 `json:"content" binding:"required"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ScheduledFor *time.Time             `json:"scheduled_for,omitempty"`
}

// Event DTOs

// Event represents a generic event
type Event struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// SystemEvent represents a system-level event
type SystemEvent struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Severity  string                 `json:"severity,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
