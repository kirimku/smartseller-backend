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
	"github.com/kirimku/smartseller-backend/pkg/cache"
	"github.com/kirimku/smartseller-backend/pkg/logger"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// CustomerServiceImpl implements the CustomerService interface
type CustomerServiceImpl struct {
	*BaseService
	customerRepo        repository.CustomerRepository
	storefrontRepo      repository.StorefrontRepository
	validationService   ValidationService
	notificationService NotificationService
	eventService        EventService
	cacheService        CacheService
}

// NewCustomerService creates a new customer service
func NewCustomerService(
	tenantResolver tenant.TenantResolver,
	cache cache.Cache,
	customerRepo repository.CustomerRepository,
	storefrontRepo repository.StorefrontRepository,
	validationService ValidationService,
	notificationService NotificationService,
	eventService EventService,
	cacheService CacheService,
) CustomerService {
	return &CustomerServiceImpl{
		BaseService:         NewBaseService(tenantResolver, cache, nil),
		customerRepo:        customerRepo,
		storefrontRepo:      storefrontRepo,
		validationService:   validationService,
		notificationService: notificationService,
		eventService:        eventService,
		cacheService:        cacheService,
	}
}

// RegisterCustomer handles customer registration with business logic validation
func (cs *CustomerServiceImpl) RegisterCustomer(ctx context.Context, req *dto.CustomerRegistrationRequest) (*dto.CustomerResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	cs.LogOperation(serviceCtx, "RegisterCustomer", logger.INFO, "Starting customer registration", map[string]interface{}{
		"email": req.Email,
		"phone": req.Phone,
	})

	// Validate request
	if cs.validationService != nil {
		if validation := cs.validationService.ValidateCustomerRegistration(ctx, req); validation.HasErrors() {
			return nil, cs.HandleServiceError(serviceCtx, "RegisterCustomer", validation.GetError(), map[string]interface{}{
				"validation_errors": validation.Errors,
			})
		}
	}

	// Check email uniqueness
	if cs.validationService != nil {
		if err := cs.validationService.ValidateEmailUniqueness(ctx, req.Email, nil); err != nil {
			return nil, cs.HandleServiceError(serviceCtx, "RegisterCustomer", err, map[string]interface{}{
				"email": req.Email,
			})
		}
	}

	// Check phone uniqueness if provided
	if req.Phone != nil && *req.Phone != "" {
		if cs.validationService != nil {
			if err := cs.validationService.ValidatePhoneUniqueness(ctx, *req.Phone, nil); err != nil {
				return nil, cs.HandleServiceError(serviceCtx, "RegisterCustomer", err, map[string]interface{}{
					"phone": *req.Phone,
				})
			}
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "RegisterCustomer", err, map[string]interface{}{
			"stage": "password_hashing",
		})
	}

	// Create customer entity
	customer := &entity.Customer{
		ID:           uuid.New(),
		StorefrontID: serviceCtx.StorefrontID,
		Email:        req.Email,
		Phone:        req.Phone,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Password:     hashedPassword,
		Status:       entity.CustomerStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Additional optional fields
	if req.DateOfBirth != nil {
		customer.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		customer.Gender = req.Gender
	}

	// Save customer to repository
	savedCustomer, err := cs.customerRepo.Create(ctx, customer)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "RegisterCustomer", err, map[string]interface{}{
			"customer_id": customer.ID,
		})
	}

	// Convert to response
	response := cs.entityToResponse(savedCustomer)

	// Cache customer data
	if cs.cacheService != nil {
		cs.cacheService.CacheCustomer(ctx, savedCustomer, 15*time.Minute)
	}

	// Send welcome notification
	if cs.notificationService != nil {
		go func() {
			if err := cs.notificationService.SendWelcomeEmail(ctx, savedCustomer); err != nil {
				cs.LogOperation(serviceCtx, "RegisterCustomer", logger.WARN, "Failed to send welcome email", map[string]interface{}{
					"customer_id": savedCustomer.ID,
					"error":       err.Error(),
				})
			}
		}()
	}

	// Publish event
	if cs.eventService != nil {
		go func() {
			if err := cs.eventService.PublishCustomerRegistered(ctx, savedCustomer); err != nil {
				cs.LogOperation(serviceCtx, "RegisterCustomer", logger.WARN, "Failed to publish customer registered event", map[string]interface{}{
					"customer_id": savedCustomer.ID,
					"error":       err.Error(),
				})
			}
		}()
	}

	cs.LogOperation(serviceCtx, "RegisterCustomer", logger.INFO, "Customer registration completed successfully", map[string]interface{}{
		"customer_id": savedCustomer.ID,
	})

	return response, nil
}

// AuthenticateCustomer handles customer authentication
func (cs *CustomerServiceImpl) AuthenticateCustomer(ctx context.Context, req *dto.CustomerAuthRequest) (*dto.CustomerAuthResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	cs.LogOperation(serviceCtx, "AuthenticateCustomer", logger.INFO, "Starting customer authentication", map[string]interface{}{
		"identifier": req.Email, // Using email as identifier in logs
	})

	// Get customer by email or phone
	var customer *entity.Customer
	if req.Email != "" {
		customer, err = cs.customerRepo.GetByEmail(ctx, req.Email)
	} else if req.Phone != "" {
		customer, err = cs.customerRepo.GetByPhone(ctx, req.Phone)
	} else {
		return nil, errors.NewValidationError("Email or phone is required", nil)
	}

	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "AuthenticateCustomer", err, map[string]interface{}{
			"email": req.Email,
			"phone": req.Phone,
		})
	}

	if customer == nil {
		return nil, errors.NewAuthenticationError("Invalid credentials", nil)
	}

	// Check customer status
	if customer.Status != entity.CustomerStatusActive {
		return nil, errors.NewAuthenticationError("Account is not active", map[string]interface{}{
			"status": customer.Status,
		})
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, customer.Password) {
		return nil, errors.NewAuthenticationError("Invalid credentials", nil)
	}

	// Generate tokens
	accessToken, refreshToken, err := cs.generateTokens(customer)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "AuthenticateCustomer", err, map[string]interface{}{
			"customer_id": customer.ID,
		})
	}

	// Update last login
	customer.LastLoginAt = &[]time.Time{time.Now()}[0]
	cs.customerRepo.Update(ctx, customer.ID, customer)

	// Create auth token record
	authToken := &entity.CustomerAuthToken{
		ID:           uuid.New(),
		CustomerID:   customer.ID,
		StorefrontID: customer.StorefrontID,
		TokenType:    entity.TokenTypeRefresh,
		Token:        refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:    time.Now(),
	}

	// Save refresh token
	if err := cs.customerRepo.CreateAuthToken(ctx, authToken); err != nil {
		cs.LogOperation(serviceCtx, "AuthenticateCustomer", logger.WARN, "Failed to save refresh token", map[string]interface{}{
			"customer_id": customer.ID,
			"error":       err.Error(),
		})
	}

	response := &dto.CustomerAuthResponse{
		Customer:     cs.entityToResponse(customer),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
	}

	cs.LogOperation(serviceCtx, "AuthenticateCustomer", logger.INFO, "Customer authentication successful", map[string]interface{}{
		"customer_id": customer.ID,
	})

	return response, nil
}

// GetCustomerProfile retrieves customer profile with caching
func (cs *CustomerServiceImpl) GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*dto.CustomerResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	cs.LogOperation(serviceCtx, "GetCustomerProfile", logger.DEBUG, "Retrieving customer profile", map[string]interface{}{
		"customer_id": customerID,
	})

	// Check cache first
	if cs.cacheService != nil {
		if cachedCustomer, err := cs.cacheService.GetCachedCustomer(ctx, customerID); err == nil && cachedCustomer != nil {
			return cs.entityToResponse(cachedCustomer), nil
		}
	}

	// Get from repository
	customer, err := cs.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "GetCustomerProfile", err, map[string]interface{}{
			"customer_id": customerID,
		})
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("Customer not found", map[string]interface{}{
			"customer_id": customerID,
		})
	}

	// Check tenant access
	if err := cs.CheckTenantAccess(serviceCtx, customer.StorefrontID); err != nil {
		return nil, err
	}

	// Cache the result
	if cs.cacheService != nil {
		go cs.cacheService.CacheCustomer(ctx, customer, 15*time.Minute)
	}

	return cs.entityToResponse(customer), nil
}

// UpdateCustomerProfile updates customer profile with validation
func (cs *CustomerServiceImpl) UpdateCustomerProfile(ctx context.Context, customerID uuid.UUID, req *dto.CustomerUpdateRequest) (*dto.CustomerResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	cs.LogOperation(serviceCtx, "UpdateCustomerProfile", logger.INFO, "Starting customer profile update", map[string]interface{}{
		"customer_id": customerID,
	})

	// Validate request
	if cs.validationService != nil {
		if validation := cs.validationService.ValidateCustomerUpdate(ctx, req); validation.HasErrors() {
			return nil, cs.HandleServiceError(serviceCtx, "UpdateCustomerProfile", validation.GetError(), map[string]interface{}{
				"validation_errors": validation.Errors,
			})
		}
	}

	// Get existing customer
	customer, err := cs.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "UpdateCustomerProfile", err, map[string]interface{}{
			"customer_id": customerID,
		})
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("Customer not found", map[string]interface{}{
			"customer_id": customerID,
		})
	}

	// Check tenant access
	if err := cs.CheckTenantAccess(serviceCtx, customer.StorefrontID); err != nil {
		return nil, err
	}

	// Check email uniqueness if changing email
	if req.Email != nil && *req.Email != customer.Email {
		if cs.validationService != nil {
			if err := cs.validationService.ValidateEmailUniqueness(ctx, *req.Email, &customerID); err != nil {
				return nil, cs.HandleServiceError(serviceCtx, "UpdateCustomerProfile", err, map[string]interface{}{
					"new_email": *req.Email,
				})
			}
		}
		customer.Email = *req.Email
	}

	// Check phone uniqueness if changing phone
	if req.Phone != nil && (customer.Phone == nil || *req.Phone != *customer.Phone) {
		if *req.Phone != "" {
			if cs.validationService != nil {
				if err := cs.validationService.ValidatePhoneUniqueness(ctx, *req.Phone, &customerID); err != nil {
					return nil, cs.HandleServiceError(serviceCtx, "UpdateCustomerProfile", err, map[string]interface{}{
						"new_phone": *req.Phone,
					})
				}
			}
		}
		customer.Phone = req.Phone
	}

	// Update other fields
	if req.FirstName != nil {
		customer.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		customer.LastName = *req.LastName
	}
	if req.DateOfBirth != nil {
		customer.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		customer.Gender = req.Gender
	}

	customer.UpdatedAt = time.Now()

	// Save updated customer
	updatedCustomer, err := cs.customerRepo.Update(ctx, customerID, customer)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "UpdateCustomerProfile", err, map[string]interface{}{
			"customer_id": customerID,
		})
	}

	// Invalidate cache
	if cs.cacheService != nil {
		go cs.cacheService.InvalidateCustomerCache(ctx, customerID)
	}

	// Publish event
	if cs.eventService != nil {
		go func() {
			if err := cs.eventService.PublishCustomerUpdated(ctx, updatedCustomer); err != nil {
				cs.LogOperation(serviceCtx, "UpdateCustomerProfile", logger.WARN, "Failed to publish customer updated event", map[string]interface{}{
					"customer_id": customerID,
					"error":       err.Error(),
				})
			}
		}()
	}

	cs.LogOperation(serviceCtx, "UpdateCustomerProfile", logger.INFO, "Customer profile updated successfully", map[string]interface{}{
		"customer_id": customerID,
	})

	return cs.entityToResponse(updatedCustomer), nil
}

// SearchCustomers performs customer search with pagination
func (cs *CustomerServiceImpl) SearchCustomers(ctx context.Context, req *dto.CustomerSearchRequest) (*dto.PaginatedCustomerResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	cs.LogOperation(serviceCtx, "SearchCustomers", logger.DEBUG, "Starting customer search", map[string]interface{}{
		"query":     req.Query,
		"page":      req.Page,
		"page_size": req.PageSize,
	})

	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Limit max page size
	}

	// Convert DTO search request to repository search params
	searchParams := &repository.CustomerSearchParams{
		Query:     req.Query,
		Status:    req.Status,
		Page:      req.Page,
		PageSize:  req.PageSize,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	if req.CreatedAfter != nil {
		searchParams.CreatedAfter = req.CreatedAfter
	}
	if req.CreatedBefore != nil {
		searchParams.CreatedBefore = req.CreatedBefore
	}

	// Perform search
	customers, total, err := cs.customerRepo.Search(ctx, searchParams)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "SearchCustomers", err, map[string]interface{}{
			"search_params": searchParams,
		})
	}

	// Convert to response DTOs
	customerResponses := make([]*dto.CustomerResponse, len(customers))
	for i, customer := range customers {
		customerResponses[i] = cs.entityToResponse(customer)
	}

	response := &dto.PaginatedCustomerResponse{
		Customers:  customerResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	cs.LogOperation(serviceCtx, "SearchCustomers", logger.DEBUG, "Customer search completed", map[string]interface{}{
		"results_count": len(customers),
		"total_count":   total,
	})

	return response, nil
}

// RefreshToken handles token refresh logic
func (cs *CustomerServiceImpl) RefreshToken(ctx context.Context, req *dto.TokenRefreshRequest) (*dto.CustomerAuthResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate refresh token
	authToken, err := cs.customerRepo.GetAuthTokenByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.NewAuthenticationError("Invalid refresh token", nil)
	}

	if authToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.NewAuthenticationError("Refresh token expired", nil)
	}

	// Get customer
	customer, err := cs.customerRepo.GetByID(ctx, authToken.CustomerID)
	if err != nil || customer == nil {
		return nil, errors.NewAuthenticationError("Customer not found", nil)
	}

	// Generate new tokens
	accessToken, refreshToken, err := cs.generateTokens(customer)
	if err != nil {
		return nil, err
	}

	// Update refresh token
	authToken.Token = refreshToken
	authToken.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	cs.customerRepo.UpdateAuthToken(ctx, authToken)

	return &dto.CustomerAuthResponse{
		Customer:     cs.entityToResponse(customer),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}

// ChangePassword handles password change with validation
func (cs *CustomerServiceImpl) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return err
	}

	// Get customer
	customer, err := cs.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil || customer == nil {
		return errors.NewNotFoundError("Customer not found", nil)
	}

	// Check tenant access
	if err := cs.CheckTenantAccess(serviceCtx, customer.StorefrontID); err != nil {
		return err
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, customer.Password) {
		return errors.NewValidationError("Current password is incorrect", nil)
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	customer.Password = hashedPassword
	customer.UpdatedAt = time.Now()

	_, err = cs.customerRepo.Update(ctx, customer.ID, customer)
	if err != nil {
		return err
	}

	// Send notification
	if cs.notificationService != nil {
		go cs.notificationService.SendPasswordChangedEmail(ctx, customer)
	}

	return nil
}

// RequestPasswordReset initiates password reset process
func (cs *CustomerServiceImpl) RequestPasswordReset(ctx context.Context, req *dto.PasswordResetRequest) error {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return err
	}

	// Get customer by email
	customer, err := cs.customerRepo.GetByEmail(ctx, req.Email)
	if err != nil || customer == nil {
		// Don't reveal if email exists or not
		return nil
	}

	// Generate reset token
	resetToken := uuid.New().String()

	// Create password reset token
	authToken := &entity.CustomerAuthToken{
		ID:           uuid.New(),
		CustomerID:   customer.ID,
		StorefrontID: customer.StorefrontID,
		TokenType:    entity.TokenTypePasswordReset,
		Token:        resetToken,
		ExpiresAt:    time.Now().Add(1 * time.Hour), // 1 hour expiry
		CreatedAt:    time.Now(),
	}

	if err := cs.customerRepo.CreateAuthToken(ctx, authToken); err != nil {
		return err
	}

	// Send reset email
	if cs.notificationService != nil {
		go cs.notificationService.SendPasswordResetEmail(ctx, customer, resetToken)
	}

	return nil
}

// ResetPassword completes password reset process
func (cs *CustomerServiceImpl) ResetPassword(ctx context.Context, req *dto.PasswordResetConfirmRequest) error {
	// Validate reset token
	authToken, err := cs.customerRepo.GetAuthTokenByToken(ctx, req.Token)
	if err != nil {
		return errors.NewValidationError("Invalid reset token", nil)
	}

	if authToken.TokenType != entity.TokenTypePasswordReset {
		return errors.NewValidationError("Invalid token type", nil)
	}

	if authToken.ExpiresAt.Before(time.Now()) {
		return errors.NewValidationError("Reset token expired", nil)
	}

	// Get customer
	customer, err := cs.customerRepo.GetByID(ctx, authToken.CustomerID)
	if err != nil || customer == nil {
		return errors.NewNotFoundError("Customer not found", nil)
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	customer.Password = hashedPassword
	customer.UpdatedAt = time.Now()

	if _, err := cs.customerRepo.Update(ctx, customer.ID, customer); err != nil {
		return err
	}

	// Delete reset token
	cs.customerRepo.DeleteAuthToken(ctx, authToken.ID)

	// Send confirmation email
	if cs.notificationService != nil {
		go cs.notificationService.SendPasswordChangedEmail(ctx, customer)
	}

	return nil
}

// LogoutCustomer handles customer logout
func (cs *CustomerServiceImpl) LogoutCustomer(ctx context.Context, tokenID uuid.UUID) error {
	return cs.customerRepo.DeleteAuthToken(ctx, tokenID)
}

// DeactivateCustomer deactivates a customer account
func (cs *CustomerServiceImpl) DeactivateCustomer(ctx context.Context, customerID uuid.UUID, reason string) error {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return err
	}

	customer, err := cs.customerRepo.GetByID(ctx, customerID)
	if err != nil || customer == nil {
		return errors.NewNotFoundError("Customer not found", nil)
	}

	// Check tenant access
	if err := cs.CheckTenantAccess(serviceCtx, customer.StorefrontID); err != nil {
		return err
	}

	customer.Status = entity.CustomerStatusInactive
	customer.UpdatedAt = time.Now()

	if _, err := cs.customerRepo.Update(ctx, customerID, customer); err != nil {
		return err
	}

	// Invalidate cache
	if cs.cacheService != nil {
		go cs.cacheService.InvalidateCustomerCache(ctx, customerID)
	}

	// Publish event
	if cs.eventService != nil {
		go cs.eventService.PublishCustomerDeactivated(ctx, customerID, reason)
	}

	return nil
}

// ReactivateCustomer reactivates a customer account
func (cs *CustomerServiceImpl) ReactivateCustomer(ctx context.Context, customerID uuid.UUID) error {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return err
	}

	customer, err := cs.customerRepo.GetByID(ctx, customerID)
	if err != nil || customer == nil {
		return errors.NewNotFoundError("Customer not found", nil)
	}

	// Check tenant access
	if err := cs.CheckTenantAccess(serviceCtx, customer.StorefrontID); err != nil {
		return err
	}

	customer.Status = entity.CustomerStatusActive
	customer.UpdatedAt = time.Now()

	if _, err := cs.customerRepo.Update(ctx, customerID, customer); err != nil {
		return err
	}

	// Invalidate cache
	if cs.cacheService != nil {
		go cs.cacheService.InvalidateCustomerCache(ctx, customerID)
	}

	return nil
}

// GetCustomerByID retrieves customer by ID
func (cs *CustomerServiceImpl) GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*dto.CustomerResponse, error) {
	return cs.GetCustomerProfile(ctx, customerID)
}

// GetCustomerByEmail retrieves customer by email
func (cs *CustomerServiceImpl) GetCustomerByEmail(ctx context.Context, email string) (*dto.CustomerResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	customer, err := cs.customerRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "GetCustomerByEmail", err, map[string]interface{}{
			"email": email,
		})
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("Customer not found", map[string]interface{}{
			"email": email,
		})
	}

	// Check tenant access
	if err := cs.CheckTenantAccess(serviceCtx, customer.StorefrontID); err != nil {
		return nil, err
	}

	return cs.entityToResponse(customer), nil
}

// GetCustomerByPhone retrieves customer by phone
func (cs *CustomerServiceImpl) GetCustomerByPhone(ctx context.Context, phone string) (*dto.CustomerResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	customer, err := cs.customerRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, cs.HandleServiceError(serviceCtx, "GetCustomerByPhone", err, map[string]interface{}{
			"phone": phone,
		})
	}

	if customer == nil {
		return nil, errors.NewNotFoundError("Customer not found", map[string]interface{}{
			"phone": phone,
		})
	}

	// Check tenant access
	if err := cs.CheckTenantAccess(serviceCtx, customer.StorefrontID); err != nil {
		return nil, err
	}

	return cs.entityToResponse(customer), nil
}

// GetCustomerStats retrieves customer statistics
func (cs *CustomerServiceImpl) GetCustomerStats(ctx context.Context, req *dto.CustomerStatsRequest) (*dto.CustomerStatsResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	// This would typically call a repository method for stats
	// For now, return a placeholder response
	return &dto.CustomerStatsResponse{
		TotalCustomers:    100,
		ActiveCustomers:   85,
		InactiveCustomers: 15,
		NewCustomers:      20,
		Period:            "30d",
		Timestamp:         time.Now(),
	}, nil
}

// GetCustomerActivity retrieves customer activity
func (cs *CustomerServiceImpl) GetCustomerActivity(ctx context.Context, customerID uuid.UUID, req *dto.ActivityRequest) (*dto.CustomerActivityResponse, error) {
	serviceCtx, err := cs.NewServiceContext(ctx)
	if err != nil {
		return nil, err
	}

	// This would typically call a repository method for activity
	// For now, return a placeholder response
	return &dto.CustomerActivityResponse{
		CustomerID: customerID,
		Activities: []*dto.Activity{},
		TotalCount: 0,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// BulkUpdateCustomers performs bulk customer updates
func (cs *CustomerServiceImpl) BulkUpdateCustomers(ctx context.Context, req *dto.BulkCustomerUpdateRequest) (*dto.BulkOperationResponse, error) {
	// This would implement bulk update logic
	return &dto.BulkOperationResponse{
		TotalRequested: len(req.CustomerIDs),
		Successful:     0,
		Failed:         0,
		Errors:         []dto.BulkOperationError{},
	}, nil
}

// ExportCustomers exports customers to various formats
func (cs *CustomerServiceImpl) ExportCustomers(ctx context.Context, req *dto.CustomerExportRequest) (*dto.ExportResponse, error) {
	// This would implement export logic
	return &dto.ExportResponse{
		ExportID:  uuid.New(),
		Status:    "pending",
		CreatedAt: time.Now(),
	}, nil
}

// Helper methods

func (cs *CustomerServiceImpl) entityToResponse(customer *entity.Customer) *dto.CustomerResponse {
	response := &dto.CustomerResponse{
		ID:           customer.ID,
		StorefrontID: customer.StorefrontID,
		Email:        customer.Email,
		Phone:        customer.Phone,
		FirstName:    customer.FirstName,
		LastName:     customer.LastName,
		Status:       string(customer.Status),
		CreatedAt:    customer.CreatedAt,
		UpdatedAt:    customer.UpdatedAt,
	}

	if customer.DateOfBirth != nil {
		response.DateOfBirth = customer.DateOfBirth
	}
	if customer.Gender != nil {
		response.Gender = customer.Gender
	}
	if customer.LastLoginAt != nil {
		response.LastLoginAt = customer.LastLoginAt
	}

	return response
}

func (cs *CustomerServiceImpl) generateTokens(customer *entity.Customer) (string, string, error) {
	// This would typically use a JWT library to generate tokens
	// For now, return placeholder tokens
	accessToken := fmt.Sprintf("access_token_%s_%d", customer.ID.String(), time.Now().Unix())
	refreshToken := fmt.Sprintf("refresh_token_%s_%d", customer.ID.String(), time.Now().Unix())

	return accessToken, refreshToken, nil
}
