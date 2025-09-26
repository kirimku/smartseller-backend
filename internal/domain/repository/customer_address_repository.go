package repository

import (
	"context"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/google/uuid"
)

// CustomerAddressRepository defines the interface for customer address data operations
type CustomerAddressRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, address *entity.CustomerAddress) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CustomerAddress, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error)
	GetDefaultByCustomerID(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error)
	GetBillingAddressByCustomerID(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error)
	GetShippingAddressByCustomerID(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error)
	Update(ctx context.Context, address *entity.CustomerAddress) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Address management operations
	SetAsDefault(ctx context.Context, customerID, addressID uuid.UUID) error
	UnsetDefault(ctx context.Context, customerID uuid.UUID) error
	SetAsDefaultBilling(ctx context.Context, customerID, addressID uuid.UUID) error
	SetAsDefaultShipping(ctx context.Context, customerID, addressID uuid.UUID) error
	
	// Business queries
	GetAddressesByType(ctx context.Context, customerID uuid.UUID, addressType entity.AddressType) ([]*entity.CustomerAddress, error)
	GetActiveAddresses(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error)
	SearchAddresses(ctx context.Context, customerID uuid.UUID, query string) ([]*entity.CustomerAddress, error)
	
	// Multi-tenant operations (addresses are accessed via customer, so tenant isolation is implicit)
	GetAddressesByStorefront(ctx context.Context, storefrontID uuid.UUID, req *GetAddressesRequest) (*AddressListResponse, error)
	GetAddressStats(ctx context.Context, storefrontID uuid.UUID) (*AddressStats, error)
	
	// Geolocation operations
	GetAddressesByCoordinates(ctx context.Context, customerID uuid.UUID, lat, lng float64, radiusKm float64) ([]*entity.CustomerAddress, error)
	GetAddressesByCity(ctx context.Context, storefrontID uuid.UUID, city string) ([]*entity.CustomerAddress, error)
	GetAddressesByCountry(ctx context.Context, storefrontID uuid.UUID, country string) ([]*entity.CustomerAddress, error)
	UpdateCoordinates(ctx context.Context, addressID uuid.UUID, lat, lng float64) error
	
	// Bulk operations
	GetAddressesByIDs(ctx context.Context, addressIDs []uuid.UUID) ([]*entity.CustomerAddress, error)
	BulkUpdateStatus(ctx context.Context, addressIDs []uuid.UUID, isActive bool) error
	DeleteByCustomerID(ctx context.Context, customerID uuid.UUID) error
	
	// Validation operations
	ValidateAddress(ctx context.Context, address *entity.CustomerAddress) error
	CheckDuplicateAddress(ctx context.Context, customerID uuid.UUID, address *entity.CustomerAddress) (bool, error)
	
	// Data cleanup
	CleanupInactiveAddresses(ctx context.Context, olderThanDays int) (int, error)
}

// CustomerSessionRepository defines the interface for customer session management
type CustomerSessionRepository interface {
	// Session lifecycle
	Create(ctx context.Context, session *CustomerSession) error
	GetBySessionToken(ctx context.Context, sessionToken string) (*CustomerSession, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*CustomerSession, error)
	GetByCustomerID(ctx context.Context, storefrontID, customerID uuid.UUID) ([]*CustomerSession, error)
	Update(ctx context.Context, session *CustomerSession) error
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, storefrontID, customerID uuid.UUID) error
	
	// Session management
	RefreshSession(ctx context.Context, refreshToken string, newSessionToken, newRefreshToken string) error
	UpdateLastUsed(ctx context.Context, sessionID uuid.UUID) error
	ExtendSession(ctx context.Context, sessionID uuid.UUID, newExpiresAt time.Time) error
	
	// Security operations
	GetActiveSessions(ctx context.Context, storefrontID, customerID uuid.UUID) ([]*CustomerSession, error)
	GetSessionsByIPAddress(ctx context.Context, ipAddress string, limit int) ([]*CustomerSession, error)
	GetSuspiciousSessions(ctx context.Context, storefrontID uuid.UUID) ([]*CustomerSession, error)
	
	// Cleanup operations
	CleanupExpiredSessions(ctx context.Context) (int, error)
	CleanupRevokedSessions(ctx context.Context, olderThanDays int) (int, error)
	
	// Analytics
	GetSessionStats(ctx context.Context, storefrontID uuid.UUID) (*SessionStats, error)
	GetActiveSessionCount(ctx context.Context, storefrontID uuid.UUID) (int, error)
}

// Supporting types for Customer Address operations
type GetAddressesRequest struct {
	StorefrontID  uuid.UUID                `json:"storefront_id"`
	CustomerID    *uuid.UUID               `json:"customer_id,omitempty"`
	AddressType   *entity.AddressType      `json:"address_type,omitempty"`
	Country       string                   `json:"country,omitempty"`
	City          string                   `json:"city,omitempty"`
	IsActive      *bool                    `json:"is_active,omitempty"`
	IsDefault     *bool                    `json:"is_default,omitempty"`
	Page          int                      `json:"page"`
	PageSize      int                      `json:"page_size"`
	OrderBy       string                   `json:"order_by"`
	SortDesc      bool                     `json:"sort_desc"`
}

type AddressListResponse struct {
	Addresses  []*entity.CustomerAddress `json:"addresses"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

type AddressStats struct {
	TotalAddresses     int                          `json:"total_addresses"`
	AddressesByType    map[entity.AddressType]int   `json:"addresses_by_type"`
	AddressesByCountry map[string]int               `json:"addresses_by_country"`
	AddressesByCity    map[string]int               `json:"addresses_by_city"`
	DefaultAddresses   int                          `json:"default_addresses"`
	ActiveAddresses    int                          `json:"active_addresses"`
}

// Customer Session entity
type CustomerSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	CustomerID   uuid.UUID  `json:"customer_id" db:"customer_id"`
	StorefrontID uuid.UUID  `json:"storefront_id" db:"storefront_id"`
	
	// Session tokens
	SessionToken string `json:"session_token" db:"session_token"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
	
	// Session metadata
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	
	// Timing
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt  time.Time  `json:"last_used_at" db:"last_used_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

type SessionStats struct {
	ActiveSessions    int     `json:"active_sessions"`
	TotalSessions     int     `json:"total_sessions"`
	SessionsToday     int     `json:"sessions_today"`
	SessionsThisWeek  int     `json:"sessions_this_week"`
	SessionsThisMonth int     `json:"sessions_this_month"`
	AvgSessionLength  float64 `json:"avg_session_length_minutes"`
	UniqueUsers       int     `json:"unique_users"`
	DeviceBreakdown   map[string]int `json:"device_breakdown"`
	LocationBreakdown map[string]int `json:"location_breakdown"`
}