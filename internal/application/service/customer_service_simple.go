package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/middleware"
)

// CustomerServiceSimple provides core customer business logic
type CustomerServiceSimple struct {
	customerRepo   repository.CustomerRepository
	tenantResolver tenant.TenantResolver
}

// NewCustomerServiceSimple creates a new simple customer service
func NewCustomerServiceSimple(
	customerRepo repository.CustomerRepository,
	tenantResolver tenant.TenantResolver,
) *CustomerServiceSimple {
	return &CustomerServiceSimple{
		customerRepo:   customerRepo,
		tenantResolver: tenantResolver,
	}
}

// RegisterCustomer handles customer registration with basic validation
func (cs *CustomerServiceSimple) RegisterCustomer(ctx context.Context, req *dto.CustomerRegistrationRequest) (*dto.CustomerResponse, error) {
	log.Printf("[DEBUG] RegisterCustomer called with email: %s", req.Email)
	
	// Get tenant context to extract storefront ID and seller ID
	tenantContext := middleware.GetTenantContextFromRequest(ctx)
	if tenantContext == nil {
		log.Printf("[ERROR] Failed to get tenant context")
		return nil, fmt.Errorf("failed to get tenant context")
	}
	
	storefrontID := tenantContext.StorefrontID
	sellerID := tenantContext.SellerID
	log.Printf("[DEBUG] Got storefront ID: %s, seller ID: %s", storefrontID, sellerID)

	// Basic validation
	if req.Email == "" {
		return nil, errors.NewValidationError("email is required", nil)
	}
	if req.FirstName == "" {
		return nil, errors.NewValidationError("first name is required", nil)
	}
	if req.LastName == "" {
		return nil, errors.NewValidationError("last name is required", nil)
	}
	if req.Password == "" {
		return nil, errors.NewValidationError("password is required", nil)
	}

	// Check if email already exists
	log.Printf("[DEBUG] Checking if email exists: %s", req.Email)
	existingCustomer, err := cs.customerRepo.GetByEmail(ctx, storefrontID, req.Email)
	if err != nil {
		log.Printf("[DEBUG] Error checking existing customer (this is normal if customer doesn't exist): %v", err)
	}
	if err == nil && existingCustomer != nil {
		log.Printf("[ERROR] Email already exists: %s", req.Email)
		return nil, errors.NewValidationError("email already exists", nil)
	}
	log.Printf("[DEBUG] Email is available: %s", req.Email)

	// Create customer entity
	customer := &entity.Customer{
		ID:           uuid.New(),
		StorefrontID: storefrontID,
		Email:        &req.Email,
		FirstName:    &req.FirstName,
		LastName:     &req.LastName,
		Phone:        req.Phone,
		DateOfBirth:  req.DateOfBirth,
		Gender:       req.Gender,
		Status:       entity.CustomerStatusActive,
		CustomerType: entity.CustomerTypeRegular,
		CreatedBy:    sellerID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Set default preferences
	customer.SetDefaultPreferences()

	// Hash password - for now just store as is (you should hash in production)
	customer.PasswordHash = &req.Password

	// Validate customer
	log.Printf("[DEBUG] Validating customer entity")
	if err := customer.Validate(); err != nil {
		log.Printf("[ERROR] Customer validation failed: %v", err)
		return nil, fmt.Errorf("customer validation failed: %w", err)
	}
	log.Printf("[DEBUG] Customer validation passed")

	// Save customer
	log.Printf("[DEBUG] Saving customer to database")
	if err := cs.customerRepo.Create(ctx, customer); err != nil {
		log.Printf("[ERROR] Failed to create customer: %v", err)
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	log.Printf("[DEBUG] Customer saved successfully")

	// Convert to response
	log.Printf("[DEBUG] Converting customer to response")
	return cs.entityToResponse(customer), nil
}

// GetCustomerByID retrieves customer by ID
func (cs *CustomerServiceSimple) GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*dto.CustomerResponse, error) {
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	customer, err := cs.customerRepo.GetByID(ctx, storefrontID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("customer not found")
	}

	return cs.entityToResponse(customer), nil
}

// GetCustomerByEmail retrieves customer by email
func (cs *CustomerServiceSimple) GetCustomerByEmail(ctx context.Context, email string) (*dto.CustomerResponse, error) {
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	customer, err := cs.customerRepo.GetByEmail(ctx, storefrontID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("customer not found")
	}

	return cs.entityToResponse(customer), nil
}

// AuthenticateCustomer handles customer authentication
func (cs *CustomerServiceSimple) AuthenticateCustomer(ctx context.Context, req *dto.CustomerAuthRequest) (*dto.CustomerAuthResponse, error) {
	// Get storefront ID from context
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	// Basic validation
	if req.Email == "" && req.Phone == "" {
		return nil, errors.NewValidationError("email or phone is required", nil)
	}
	if req.Password == "" {
		return nil, errors.NewValidationError("password is required", nil)
	}

	// Get customer by email or phone
	var customer *entity.Customer
	if req.Email != "" {
		customer, err = cs.customerRepo.GetByEmail(ctx, storefrontID, req.Email)
	} else {
		customer, err = cs.customerRepo.GetByPhone(ctx, storefrontID, req.Phone)
	}

	if err != nil || customer == nil {
		return nil, errors.NewValidationError("invalid credentials", nil)
	}

	// TODO: Implement password verification
	// For now, just return a basic response
	return &dto.CustomerAuthResponse{
		Customer:     cs.entityToResponse(customer),
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}

// UpdateCustomer updates customer information
func (cs *CustomerServiceSimple) UpdateCustomer(ctx context.Context, customerID uuid.UUID, req *dto.CustomerUpdateRequest) (*dto.CustomerResponse, error) {
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	// Get existing customer
	customer, err := cs.customerRepo.GetByID(ctx, storefrontID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("customer not found")
	}

	// Update fields if provided
	if req.Email != nil {
		customer.Email = req.Email
	}
	if req.Phone != nil {
		customer.Phone = req.Phone
	}
	if req.FirstName != nil {
		customer.FirstName = req.FirstName
	}
	if req.LastName != nil {
		customer.LastName = req.LastName
	}
	if req.DateOfBirth != nil {
		customer.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		customer.Gender = req.Gender
	}

	customer.UpdatedAt = time.Now()

	// Validate updated customer
	if err := customer.Validate(); err != nil {
		return nil, fmt.Errorf("customer validation failed: %w", err)
	}

	// Save updated customer
	if err := cs.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return cs.entityToResponse(customer), nil
}

// SearchCustomers searches customers with basic pagination
func (cs *CustomerServiceSimple) SearchCustomers(ctx context.Context, req *dto.CustomerSearchRequest) (*dto.PaginatedCustomerResponse, error) {
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Convert to repository search params
	searchParams := &repository.SearchCustomersRequest{
		Query:    req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.Status != nil {
		searchParams.Status = req.Status
	}
	if req.CustomerType != nil {
		searchParams.CustomerType = req.CustomerType
	}

	// Perform search
	result, err := cs.customerRepo.Search(ctx, storefrontID, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	// Convert to response DTOs
	customerResponses := make([]*dto.CustomerResponse, len(result.Customers))
	for i, customer := range result.Customers {
		customerResponses[i] = cs.entityToResponse(customer)
	}

	total := int64(result.Total)
	return &dto.PaginatedCustomerResponse{
		Customers:  customerResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}, nil
}

// RefreshToken handles token refresh
func (cs *CustomerServiceSimple) RefreshToken(ctx context.Context, req *dto.TokenRefreshRequest) (*dto.CustomerAuthResponse, error) {
	// TODO: Implement token refresh logic
	return nil, fmt.Errorf("token refresh not implemented")
}

// LogoutCustomer handles customer logout
func (cs *CustomerServiceSimple) LogoutCustomer(ctx context.Context, tokenID uuid.UUID) error {
	// TODO: Implement logout logic
	return fmt.Errorf("logout not implemented")
}

// ChangePassword handles password change
func (cs *CustomerServiceSimple) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error {
	// TODO: Implement password change logic
	return fmt.Errorf("change password not implemented")
}

// RequestPasswordReset handles password reset request
func (cs *CustomerServiceSimple) RequestPasswordReset(ctx context.Context, req *dto.PasswordResetRequest) error {
	// TODO: Implement password reset request logic
	return fmt.Errorf("password reset request not implemented")
}

// ResetPassword handles password reset confirmation
func (cs *CustomerServiceSimple) ResetPassword(ctx context.Context, req *dto.PasswordResetConfirmRequest) error {
	// TODO: Implement password reset confirmation logic
	return fmt.Errorf("password reset confirmation not implemented")
}

// GetCustomerProfile retrieves customer profile
func (cs *CustomerServiceSimple) GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*dto.CustomerResponse, error) {
	return cs.GetCustomerByID(ctx, customerID)
}

// UpdateCustomerProfile updates customer profile
func (cs *CustomerServiceSimple) UpdateCustomerProfile(ctx context.Context, customerID uuid.UUID, req *dto.CustomerUpdateRequest) (*dto.CustomerResponse, error) {
	return cs.UpdateCustomer(ctx, customerID, req)
}

// DeactivateCustomer deactivates a customer
func (cs *CustomerServiceSimple) DeactivateCustomer(ctx context.Context, customerID uuid.UUID, reason string) error {
	// TODO: Implement customer deactivation logic
	return fmt.Errorf("customer deactivation not implemented")
}

// ReactivateCustomer reactivates a customer
func (cs *CustomerServiceSimple) ReactivateCustomer(ctx context.Context, customerID uuid.UUID) error {
	// TODO: Implement customer reactivation logic
	return fmt.Errorf("customer reactivation not implemented")
}

// GetCustomerByPhone retrieves customer by phone
func (cs *CustomerServiceSimple) GetCustomerByPhone(ctx context.Context, phone string) (*dto.CustomerResponse, error) {
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	customer, err := cs.customerRepo.GetByPhone(ctx, storefrontID, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("customer not found")
	}

	return cs.entityToResponse(customer), nil
}

// GetCustomerActivity retrieves customer activity
func (cs *CustomerServiceSimple) GetCustomerActivity(ctx context.Context, customerID uuid.UUID, req *dto.ActivityRequest) (*dto.CustomerActivityResponse, error) {
	// TODO: Implement customer activity logic
	return nil, fmt.Errorf("customer activity not implemented")
}

// BulkUpdateCustomers handles bulk customer updates
func (cs *CustomerServiceSimple) BulkUpdateCustomers(ctx context.Context, req *dto.BulkCustomerUpdateRequest) (*dto.BulkOperationResponse, error) {
	// TODO: Implement bulk customer updates logic
	return nil, fmt.Errorf("bulk customer updates not implemented")
}

// ExportCustomers handles customer export
func (cs *CustomerServiceSimple) ExportCustomers(ctx context.Context, req *dto.CustomerExportRequest) (*dto.ExportResponse, error) {
	// TODO: Implement customer export logic
	return nil, fmt.Errorf("customer export not implemented")
}

// DeleteCustomer soft deletes a customer
func (cs *CustomerServiceSimple) DeleteCustomer(ctx context.Context, customerID uuid.UUID) error {
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get storefront context: %w", err)
	}

	// Get customer to ensure it exists
	customer, err := cs.customerRepo.GetByID(ctx, storefrontID, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if customer == nil {
		return errors.NewNotFoundError("customer not found")
	}

	// Soft delete
	return cs.customerRepo.SoftDelete(ctx, storefrontID, customerID)
}

// GetCustomerStats returns basic customer statistics
func (cs *CustomerServiceSimple) GetCustomerStats(ctx context.Context, req *dto.CustomerStatsRequest) (*dto.CustomerStatsResponse, error) {
	// Get storefront ID from tenant context
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	// Get customer statistics
	stats, err := cs.customerRepo.GetCustomerStats(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer stats: %w", err)
	}

	// Determine period from request
	period := "all_time"
	if req != nil && req.Period != "" {
		period = req.Period
	}

	// Convert repository.CustomerStats to dto.CustomerStatsResponse
	return &dto.CustomerStatsResponse{
		TotalCustomers:    int64(stats.TotalCustomers),
		ActiveCustomers:   int64(stats.ActiveCustomers),
		InactiveCustomers: int64(stats.TotalCustomers - stats.ActiveCustomers),
		NewCustomers:      int64(stats.NewThisMonth),
		Period:            period,
		Timestamp:         time.Now(),
	}, nil
}

// Helper methods

func (cs *CustomerServiceSimple) getStorefrontFromContext(ctx context.Context) (uuid.UUID, error) {
	// Use the middleware's helper function to get tenant context
	tenantContext := middleware.GetTenantContextFromRequest(ctx)
	if tenantContext != nil {
		log.Printf("[DEBUG] Found tenant context using middleware helper: %s", tenantContext.StorefrontID)
		return tenantContext.StorefrontID, nil
	}
	
	log.Printf("[DEBUG] No tenant context found using middleware helper")
	return uuid.Nil, fmt.Errorf("storefront ID not found in context")
}

func (cs *CustomerServiceSimple) entityToResponse(customer *entity.Customer) *dto.CustomerResponse {
	response := &dto.CustomerResponse{
		ID:           customer.ID,
		StorefrontID: customer.StorefrontID,
		Email:        customer.Email,
		Phone:        customer.Phone,
		FirstName:    customer.FirstName,
		LastName:     customer.LastName,
		FullName:     customer.FullName,
		DateOfBirth:  customer.DateOfBirth,
		Gender:       customer.Gender,
		Status:       string(customer.Status),
		CustomerType: customer.CustomerType,
		Preferences:  customer.Preferences,
		LastLoginAt:  customer.LastLoginAt,
		CreatedAt:    customer.CreatedAt,
		UpdatedAt:    customer.UpdatedAt,
	}

	return response
}
