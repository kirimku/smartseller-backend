package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
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
	// Get storefront ID from context
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

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
	existingCustomer, err := cs.customerRepo.GetByEmail(ctx, storefrontID, req.Email)
	if err == nil && existingCustomer != nil {
		return nil, errors.NewValidationError("email already exists", nil)
	}

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
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Set default preferences
	customer.SetDefaultPreferences()

	// Hash password - for now just store as is (you should hash in production)
	customer.PasswordHash = &req.Password

	// Validate customer
	if err := customer.Validate(); err != nil {
		return nil, fmt.Errorf("customer validation failed: %w", err)
	}

	// Save customer
	if err := cs.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Convert to response
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
		return nil, errors.NewNotFoundError("customer not found", nil)
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
		return nil, errors.NewNotFoundError("customer not found", nil)
	}

	return cs.entityToResponse(customer), nil
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
		return nil, errors.NewNotFoundError("customer not found", nil)
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
	searchParams := &repository.CustomerSearchParams{
		Query:     req.Query,
		Page:      req.Page,
		PageSize:  req.PageSize,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	if req.Status != nil {
		searchParams.Status = *req.Status
	}
	if req.CreatedAfter != nil {
		searchParams.CreatedAfter = req.CreatedAfter
	}
	if req.CreatedBefore != nil {
		searchParams.CreatedBefore = req.CreatedBefore
	}

	// Perform search
	customers, total, err := cs.customerRepo.Search(ctx, storefrontID, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	// Convert to response DTOs
	customerResponses := make([]*dto.CustomerResponse, len(customers))
	for i, customer := range customers {
		customerResponses[i] = cs.entityToResponse(customer)
	}

	return &dto.PaginatedCustomerResponse{
		Customers:  customerResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}, nil
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
		return errors.NewNotFoundError("customer not found", nil)
	}

	// Soft delete
	return cs.customerRepo.Delete(ctx, storefrontID, customerID)
}

// GetCustomerStats returns basic customer statistics
func (cs *CustomerServiceSimple) GetCustomerStats(ctx context.Context) (*dto.CustomerStatsResponse, error) {
	storefrontID, err := cs.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront context: %w", err)
	}

	stats, err := cs.customerRepo.GetStats(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer stats: %w", err)
	}

	return &dto.CustomerStatsResponse{
		TotalCustomers:    stats.TotalCustomers,
		ActiveCustomers:   stats.ActiveCustomers,
		InactiveCustomers: stats.InactiveCustomers,
		NewCustomers:      stats.NewCustomers,
		Period:            "all_time",
		Timestamp:         time.Now(),
	}, nil
}

// Helper methods

func (cs *CustomerServiceSimple) getStorefrontFromContext(ctx context.Context) (uuid.UUID, error) {
	if storefrontID, ok := ctx.Value("storefront_id").(uuid.UUID); ok {
		return storefrontID, nil
	}
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
