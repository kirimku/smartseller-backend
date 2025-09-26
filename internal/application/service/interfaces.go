package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// CustomerService defines the interface for customer business logic
type CustomerService interface {
	// Customer Registration and Authentication
	RegisterCustomer(ctx context.Context, req *dto.CustomerRegistrationRequest) (*dto.CustomerResponse, error)
	AuthenticateCustomer(ctx context.Context, req *dto.CustomerAuthRequest) (*dto.CustomerAuthResponse, error)
	RefreshToken(ctx context.Context, req *dto.TokenRefreshRequest) (*dto.CustomerAuthResponse, error)
	LogoutCustomer(ctx context.Context, tokenID uuid.UUID) error

	// Password Management
	ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error
	RequestPasswordReset(ctx context.Context, req *dto.PasswordResetRequest) error
	ResetPassword(ctx context.Context, req *dto.PasswordResetConfirmRequest) error

	// Profile Management
	GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*dto.CustomerResponse, error)
	UpdateCustomerProfile(ctx context.Context, customerID uuid.UUID, req *dto.CustomerUpdateRequest) (*dto.CustomerResponse, error)
	DeactivateCustomer(ctx context.Context, customerID uuid.UUID, reason string) error
	ReactivateCustomer(ctx context.Context, customerID uuid.UUID) error

	// Customer Search and Management
	SearchCustomers(ctx context.Context, req *dto.CustomerSearchRequest) (*dto.PaginatedCustomerResponse, error)
	GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*dto.CustomerResponse, error)
	GetCustomerByEmail(ctx context.Context, email string) (*dto.CustomerResponse, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*dto.CustomerResponse, error)

	// Customer Analytics
	GetCustomerStats(ctx context.Context, req *dto.CustomerStatsRequest) (*dto.CustomerStatsResponse, error)
	GetCustomerActivity(ctx context.Context, customerID uuid.UUID, req *dto.ActivityRequest) (*dto.CustomerActivityResponse, error)

	// Bulk Operations
	BulkUpdateCustomers(ctx context.Context, req *dto.BulkCustomerUpdateRequest) (*dto.BulkOperationResponse, error)
	ExportCustomers(ctx context.Context, req *dto.CustomerExportRequest) (*dto.ExportResponse, error)
}

// StorefrontService defines the interface for storefront business logic
type StorefrontService interface {
	// Storefront Creation and Management
	CreateStorefront(ctx context.Context, req *dto.StorefrontCreateRequest) (*dto.StorefrontResponse, error)
	UpdateStorefront(ctx context.Context, storefrontID uuid.UUID, req *dto.StorefrontUpdateRequest) (*dto.StorefrontResponse, error)
	GetStorefront(ctx context.Context, storefrontID uuid.UUID) (*dto.StorefrontResponse, error)
	GetStorefrontBySlug(ctx context.Context, slug string) (*dto.StorefrontResponse, error)
	DeleteStorefront(ctx context.Context, storefrontID uuid.UUID) error

	// Domain and Configuration Management
	ConfigureDomain(ctx context.Context, storefrontID uuid.UUID, req *dto.DomainConfigRequest) error
	ValidateDomain(ctx context.Context, domain string) (*dto.DomainValidationResponse, error)
	UpdateStorefrontSettings(ctx context.Context, storefrontID uuid.UUID, req *dto.StorefrontSettingsRequest) error

	// Status Management
	ActivateStorefront(ctx context.Context, storefrontID uuid.UUID) error
	DeactivateStorefront(ctx context.Context, storefrontID uuid.UUID, reason string) error
	SuspendStorefront(ctx context.Context, storefrontID uuid.UUID, reason string) error

	// Search and Analytics
	SearchStorefronts(ctx context.Context, req *dto.StorefrontSearchRequest) (*dto.PaginatedStorefrontResponse, error)
	GetStorefrontStats(ctx context.Context, storefrontID uuid.UUID, req *dto.StorefrontStatsRequest) (*dto.StorefrontStatsResponse, error)
	GetStorefrontActivity(ctx context.Context, storefrontID uuid.UUID, req *dto.ActivityRequest) (*dto.StorefrontActivityResponse, error)

	// Tenant Management
	MigrateStorefront(ctx context.Context, storefrontID uuid.UUID, targetType string) error
	GetMigrationStatus(ctx context.Context, storefrontID uuid.UUID) (*dto.MigrationStatusResponse, error)
}

// CustomerAddressService defines the interface for customer address management
type CustomerAddressService interface {
	// Address Management
	CreateAddress(ctx context.Context, req *dto.CreateAddressRequest) (*dto.CustomerAddressResponse, error)
	UpdateAddress(ctx context.Context, addressID uuid.UUID, req *dto.UpdateAddressRequest) (*dto.CustomerAddressResponse, error)
	GetAddress(ctx context.Context, addressID uuid.UUID) (*dto.CustomerAddressResponse, error)
	DeleteAddress(ctx context.Context, addressID uuid.UUID) error

	// Customer Address Operations
	GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*dto.CustomerAddressResponse, error)
	SetDefaultAddress(ctx context.Context, customerID, addressID uuid.UUID) error
	GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*dto.CustomerAddressResponse, error)

	// Address Validation and Enhancement
	ValidateAddress(ctx context.Context, req *dto.AddressValidationRequest) (*dto.AddressValidationResponse, error)
	GeocodeAddress(ctx context.Context, req *dto.GeocodeRequest) (*dto.GeocodeResponse, error)
	GetNearbyAddresses(ctx context.Context, req *dto.NearbyAddressRequest) ([]*dto.CustomerAddressResponse, error)

	// Bulk Operations
	BulkCreateAddresses(ctx context.Context, req *dto.BulkAddressCreateRequest) (*dto.BulkOperationResponse, error)
	BulkUpdateAddresses(ctx context.Context, req *dto.BulkAddressUpdateRequest) (*dto.BulkOperationResponse, error)
	BulkDeleteAddresses(ctx context.Context, req *dto.BulkAddressDeleteRequest) (*dto.BulkOperationResponse, error)

	// Analytics and Reporting
	GetAddressStats(ctx context.Context, req *dto.AddressStatsRequest) (*dto.AddressStatsResponse, error)
	GetAddressDistribution(ctx context.Context, req *dto.AddressDistributionRequest) (*dto.AddressDistributionResponse, error)
}

// ValidationService defines the interface for data validation
type ValidationService interface {
	// Customer Validation
	ValidateCustomerRegistration(ctx context.Context, req *dto.CustomerRegistrationRequest) *ValidationResult
	ValidateCustomerUpdate(ctx context.Context, req *dto.CustomerUpdateRequest) *ValidationResult
	ValidateEmailUniqueness(ctx context.Context, email string, excludeCustomerID *uuid.UUID) error
	ValidatePhoneUniqueness(ctx context.Context, phone string, excludeCustomerID *uuid.UUID) error

	// Storefront Validation
	ValidateStorefrontCreation(ctx context.Context, req *dto.StorefrontCreateRequest) *ValidationResult
	ValidateStorefrontUpdate(ctx context.Context, req *dto.StorefrontUpdateRequest) *ValidationResult
	ValidateSlugUniqueness(ctx context.Context, slug string, excludeStorefrontID *uuid.UUID) error
	ValidateDomainUniqueness(ctx context.Context, domain string, excludeStorefrontID *uuid.UUID) error

	// Address Validation
	ValidateAddressCreation(ctx context.Context, req *dto.CreateAddressRequest) *ValidationResult
	ValidateAddressUpdate(ctx context.Context, req *dto.UpdateAddressRequest) *ValidationResult
	ValidateAddressFormat(ctx context.Context, address *entity.CustomerAddress) *ValidationResult

	// Business Rule Validation
	ValidateBusinessRules(ctx context.Context, entity string, operation string, data interface{}) error
	ValidateTenantConstraints(ctx context.Context, storefrontID uuid.UUID, operation string, data interface{}) error
}

// CacheService defines the interface for caching operations
type CacheService interface {
	// Customer Caching
	CacheCustomer(ctx context.Context, customer *entity.Customer, ttl time.Duration) error
	GetCachedCustomer(ctx context.Context, customerID uuid.UUID) (*entity.Customer, error)
	InvalidateCustomerCache(ctx context.Context, customerID uuid.UUID) error

	// Storefront Caching
	CacheStorefront(ctx context.Context, storefront *entity.Storefront, ttl time.Duration) error
	GetCachedStorefront(ctx context.Context, storefrontID uuid.UUID) (*entity.Storefront, error)
	GetCachedStorefrontBySlug(ctx context.Context, slug string) (*entity.Storefront, error)
	InvalidateStorefrontCache(ctx context.Context, storefrontID uuid.UUID) error

	// Address Caching
	CacheCustomerAddresses(ctx context.Context, customerID uuid.UUID, addresses []*entity.CustomerAddress, ttl time.Duration) error
	GetCachedCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error)
	InvalidateAddressCache(ctx context.Context, customerID uuid.UUID) error

	// Statistics Caching
	CacheStats(ctx context.Context, key string, stats interface{}, ttl time.Duration) error
	GetCachedStats(ctx context.Context, key string, dest interface{}) error
	InvalidateStatsCache(ctx context.Context, pattern string) error

	// Cache Management
	ClearCache(ctx context.Context, pattern string) error
	GetCacheStats(ctx context.Context) (*dto.CacheStatsResponse, error)
	WarmupCache(ctx context.Context, req *dto.CacheWarmupRequest) error
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	// Customer Notifications
	SendWelcomeEmail(ctx context.Context, customer *entity.Customer) error
	SendPasswordResetEmail(ctx context.Context, customer *entity.Customer, resetToken string) error
	SendPasswordChangedEmail(ctx context.Context, customer *entity.Customer) error
	SendAccountActivationEmail(ctx context.Context, customer *entity.Customer) error

	// Storefront Notifications
	SendStorefrontCreatedEmail(ctx context.Context, storefront *entity.Storefront, adminEmail string) error
	SendDomainConfiguredEmail(ctx context.Context, storefront *entity.Storefront, domain string) error

	// System Notifications
	SendSystemAlert(ctx context.Context, alert *dto.SystemAlert) error
	SendPerformanceAlert(ctx context.Context, alert *dto.PerformanceAlert) error

	// Bulk Notifications
	SendBulkNotifications(ctx context.Context, req *dto.BulkNotificationRequest) (*dto.BulkOperationResponse, error)
}

// AnalyticsService defines the interface for analytics and reporting
type AnalyticsService interface {
	// Customer Analytics
	GetCustomerMetrics(ctx context.Context, req *dto.CustomerMetricsRequest) (*dto.CustomerMetricsResponse, error)
	GetCustomerSegmentation(ctx context.Context, req *dto.SegmentationRequest) (*dto.SegmentationResponse, error)
	GetCustomerLifetimeValue(ctx context.Context, customerID uuid.UUID) (*dto.LifetimeValueResponse, error)

	// Storefront Analytics
	GetStorefrontMetrics(ctx context.Context, req *dto.StorefrontMetricsRequest) (*dto.StorefrontMetricsResponse, error)
	GetStorefrontPerformance(ctx context.Context, storefrontID uuid.UUID, req *dto.PerformanceRequest) (*dto.PerformanceResponse, error)

	// System Analytics
	GetSystemMetrics(ctx context.Context, req *dto.SystemMetricsRequest) (*dto.SystemMetricsResponse, error)
	GetTenantUsage(ctx context.Context, req *dto.TenantUsageRequest) (*dto.TenantUsageResponse, error)

	// Reporting
	GenerateReport(ctx context.Context, req *dto.ReportRequest) (*dto.ReportResponse, error)
	ScheduleReport(ctx context.Context, req *dto.ScheduledReportRequest) (*dto.ScheduledReportResponse, error)
	GetReportStatus(ctx context.Context, reportID uuid.UUID) (*dto.ReportStatusResponse, error)
}

// EventService defines the interface for event handling and publishing
type EventService interface {
	// Customer Events
	PublishCustomerRegistered(ctx context.Context, customer *entity.Customer) error
	PublishCustomerUpdated(ctx context.Context, customer *entity.Customer) error
	PublishCustomerDeactivated(ctx context.Context, customerID uuid.UUID, reason string) error

	// Storefront Events
	PublishStorefrontCreated(ctx context.Context, storefront *entity.Storefront) error
	PublishStorefrontUpdated(ctx context.Context, storefront *entity.Storefront) error
	PublishStorefrontStatusChanged(ctx context.Context, storefrontID uuid.UUID, oldStatus, newStatus string) error

	// Address Events
	PublishAddressCreated(ctx context.Context, address *entity.CustomerAddress) error
	PublishAddressUpdated(ctx context.Context, address *entity.CustomerAddress) error
	PublishDefaultAddressChanged(ctx context.Context, customerID, addressID uuid.UUID) error

	// System Events
	PublishSystemEvent(ctx context.Context, event *dto.SystemEvent) error

	// Event Subscription
	SubscribeToEvents(ctx context.Context, eventTypes []string, handler func(*dto.Event) error) error
	UnsubscribeFromEvents(ctx context.Context, subscriptionID string) error
}
