package repository

import (
	"context"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/google/uuid"
)

// CustomerRepository defines the interface for customer data operations
// All operations include tenant isolation via storefront_id
type CustomerRepository interface {
	// Core CRUD operations with tenant isolation
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, storefrontID, customerID uuid.UUID) (*entity.Customer, error)
	GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error)
	GetByPhone(ctx context.Context, storefrontID uuid.UUID, phone string) (*entity.Customer, error)
	Update(ctx context.Context, customer *entity.Customer) error
	SoftDelete(ctx context.Context, storefrontID, customerID uuid.UUID) error
	HardDelete(ctx context.Context, storefrontID, customerID uuid.UUID) error
	
	// Business queries with tenant isolation
	GetByStorefront(ctx context.Context, req *GetCustomersRequest) (*CustomerListResponse, error)
	Search(ctx context.Context, storefrontID uuid.UUID, req *SearchCustomersRequest) (*CustomerSearchResult, error)
	GetTopCustomers(ctx context.Context, storefrontID uuid.UUID, limit int) ([]*entity.Customer, error)
	GetCustomerStats(ctx context.Context, storefrontID uuid.UUID) (*CustomerStats, error)
	GetCustomerSegments(ctx context.Context, storefrontID uuid.UUID) ([]*CustomerSegment, error)
	
	// Authentication-specific operations
	GetByEmailVerificationToken(ctx context.Context, token string) (*entity.Customer, error)
	GetByPasswordResetToken(ctx context.Context, token string) (*entity.Customer, error)
	UpdateLastLogin(ctx context.Context, storefrontID, customerID uuid.UUID) error
	UpdateRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string, expiresAt *time.Time) error
	ClearRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID) error
	UpdateEmailVerification(ctx context.Context, storefrontID, customerID uuid.UUID, verified bool) error
	UpdatePhoneVerification(ctx context.Context, storefrontID, customerID uuid.UUID, verified bool) error
	UpdateFailedLoginAttempts(ctx context.Context, storefrontID, customerID uuid.UUID, attempts int) error
	LockAccount(ctx context.Context, storefrontID, customerID uuid.UUID, until *time.Time) error
	UnlockAccount(ctx context.Context, storefrontID, customerID uuid.UUID) error
	
	// Password management
	UpdatePassword(ctx context.Context, storefrontID, customerID uuid.UUID, passwordHash string) error
	SetPasswordResetToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string, expiresAt time.Time) error
	ClearPasswordResetToken(ctx context.Context, storefrontID, customerID uuid.UUID) error
	
	// Email/Phone verification
	SetEmailVerificationToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string) error
	SetPhoneVerificationToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string) error
	
	// Customer metrics and analytics
	UpdateCustomerMetrics(ctx context.Context, storefrontID, customerID uuid.UUID, metrics CustomerMetricsUpdate) error
	GetCustomerActivity(ctx context.Context, storefrontID, customerID uuid.UUID, limit int) ([]*CustomerActivity, error)
	
	// Bulk operations
	GetCustomersByIDs(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID) ([]*entity.Customer, error)
	GetCustomersByStatus(ctx context.Context, storefrontID uuid.UUID, status entity.CustomerStatus) ([]*entity.Customer, error)
	GetCustomersByType(ctx context.Context, storefrontID uuid.UUID, customerType entity.CustomerType) ([]*entity.Customer, error)
	GetCustomersWithTags(ctx context.Context, storefrontID uuid.UUID, tags []string) ([]*entity.Customer, error)
	BulkUpdateStatus(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID, status entity.CustomerStatus) error
	BulkAddTags(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID, tags []string) error
	BulkRemoveTags(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID, tags []string) error
	
	// Data validation and cleanup
	ValidateUniqueEmail(ctx context.Context, storefrontID uuid.UUID, email string, excludeCustomerID *uuid.UUID) error
	ValidateUniquePhone(ctx context.Context, storefrontID uuid.UUID, phone string, excludeCustomerID *uuid.UUID) error
	CleanupExpiredTokens(ctx context.Context) (int, error)
	CleanupExpiredSessions(ctx context.Context) (int, error)
}

// Request/Response types for Customer operations
type GetCustomersRequest struct {
	StorefrontID uuid.UUID                `json:"storefront_id"`
	Page         int                       `json:"page"`
	PageSize     int                       `json:"page_size"`
	Search       string                    `json:"search,omitempty"`
	Status       *entity.CustomerStatus    `json:"status,omitempty"`
	CustomerType *entity.CustomerType      `json:"customer_type,omitempty"`
	Tags         []string                  `json:"tags,omitempty"`
	OrderBy      string                    `json:"order_by"`
	SortDesc     bool                      `json:"sort_desc"`
	DateFrom     *time.Time                `json:"date_from,omitempty"`
	DateTo       *time.Time                `json:"date_to,omitempty"`
	HasEmail     *bool                     `json:"has_email,omitempty"`
	HasPhone     *bool                     `json:"has_phone,omitempty"`
	IsVerified   *bool                     `json:"is_verified,omitempty"`
}

type CustomerListResponse struct {
	Customers  []*entity.Customer `json:"customers"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

type SearchCustomersRequest struct {
	Query         string                   `json:"query"`
	SearchFields  []string                 `json:"search_fields,omitempty"` // email, phone, name, etc.
	Status        *entity.CustomerStatus   `json:"status,omitempty"`
	CustomerType  *entity.CustomerType     `json:"customer_type,omitempty"`
	Page          int                      `json:"page"`
	PageSize      int                      `json:"page_size"`
}

type CustomerSearchResult struct {
	Customers []*entity.Customer `json:"customers"`
	Total     int                `json:"total"`
	Query     string             `json:"query"`
}

type CustomerStats struct {
	TotalCustomers       int     `json:"total_customers"`
	ActiveCustomers      int     `json:"active_customers"`
	VerifiedCustomers    int     `json:"verified_customers"`
	NewThisMonth         int     `json:"new_this_month"`
	NewThisWeek          int     `json:"new_this_week"`
	NewToday             int     `json:"new_today"`
	TotalRevenue         float64 `json:"total_revenue"`
	AvgOrderValue        float64 `json:"avg_order_value"`
	AvgCustomerLifetime  float64 `json:"avg_customer_lifetime"`
	ChurnRate            float64 `json:"churn_rate"`
	RetentionRate        float64 `json:"retention_rate"`
	TopSpenders          []struct {
		CustomerID uuid.UUID `json:"customer_id"`
		Name       string    `json:"name"`
		TotalSpent float64   `json:"total_spent"`
	} `json:"top_spenders"`
	CustomersByType      map[entity.CustomerType]int `json:"customers_by_type"`
	CustomersByStatus    map[entity.CustomerStatus]int `json:"customers_by_status"`
	GrowthTrend         []CustomerGrowthPoint `json:"growth_trend"`
}

type CustomerGrowthPoint struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

type CustomerSegment struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Count        int     `json:"count"`
	Criteria     map[string]interface{} `json:"criteria"`
	AvgOrderValue float64 `json:"avg_order_value"`
	TotalRevenue float64 `json:"total_revenue"`
}

type CustomerMetricsUpdate struct {
	TotalOrders       *int       `json:"total_orders,omitempty"`
	TotalSpent        *float64   `json:"total_spent,omitempty"`
	AverageOrderValue *float64   `json:"average_order_value,omitempty"`
	LastOrderDate     *time.Time `json:"last_order_date,omitempty"`
}

type CustomerActivity struct {
	ID           uuid.UUID   `json:"id"`
	CustomerID   uuid.UUID   `json:"customer_id"`
	StorefrontID uuid.UUID   `json:"storefront_id"`
	ActivityType string      `json:"activity_type"` // login, order, profile_update, etc.
	Description  string      `json:"description"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IPAddress    *string     `json:"ip_address,omitempty"`
	UserAgent    *string     `json:"user_agent,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}