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

// StorefrontServiceSimple provides core storefront business logic
type StorefrontServiceSimple struct {
	storefrontRepo repository.StorefrontRepository
	tenantResolver tenant.TenantResolver
}

// NewStorefrontServiceSimple creates a new simple storefront service
func NewStorefrontServiceSimple(
	storefrontRepo repository.StorefrontRepository,
	tenantResolver tenant.TenantResolver,
) *StorefrontServiceSimple {
	return &StorefrontServiceSimple{
		storefrontRepo: storefrontRepo,
		tenantResolver: tenantResolver,
	}
}

// CreateStorefront creates a new storefront with validation
func (ss *StorefrontServiceSimple) CreateStorefront(ctx context.Context, req *dto.StorefrontCreateRequest) (*dto.StorefrontResponse, error) {
	// Basic validation
	if req.Name == "" {
		return nil, errors.NewValidationError("name is required", nil)
	}
	if req.Slug == "" {
		return nil, errors.NewValidationError("slug is required", nil)
	}
	if req.OwnerEmail == "" {
		return nil, errors.NewValidationError("owner email is required", nil)
	}
	if req.OwnerName == "" {
		return nil, errors.NewValidationError("owner name is required", nil)
	}

	// Check if slug already exists
	existingStorefront, err := ss.storefrontRepo.GetBySlug(ctx, req.Slug)
	if err == nil && existingStorefront != nil {
		return nil, errors.NewValidationError("slug already exists", nil)
	}

	// Check domain uniqueness if provided
	if req.Domain != nil && *req.Domain != "" {
		existingStorefront, err := ss.storefrontRepo.GetByDomain(ctx, *req.Domain)
		if err == nil && existingStorefront != nil {
			return nil, errors.NewValidationError("domain already exists", nil)
		}
	}

	// Create storefront entity
	storefront := &entity.Storefront{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Domain:      req.Domain,
		Subdomain:   req.Subdomain,
		Status:      entity.StorefrontStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set default settings
	storefront.SetDefaultSettings()

	// Validate storefront
	if err := storefront.Validate(); err != nil {
		return nil, fmt.Errorf("storefront validation failed: %w", err)
	}

	// Save storefront
	if err := ss.storefrontRepo.Create(ctx, storefront); err != nil {
		return nil, fmt.Errorf("failed to create storefront: %w", err)
	}

	// Convert to response
	return ss.entityToResponse(storefront), nil
}

// GetStorefrontByID retrieves storefront by ID
func (ss *StorefrontServiceSimple) GetStorefrontByID(ctx context.Context, storefrontID uuid.UUID) (*dto.StorefrontResponse, error) {
	storefront, err := ss.storefrontRepo.GetByID(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return nil, errors.NewNotFoundError("storefront not found", nil)
	}

	return ss.entityToResponse(storefront), nil
}

// GetStorefrontBySlug retrieves storefront by slug
func (ss *StorefrontServiceSimple) GetStorefrontBySlug(ctx context.Context, slug string) (*dto.StorefrontResponse, error) {
	storefront, err := ss.storefrontRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return nil, errors.NewNotFoundError("storefront not found", nil)
	}

	return ss.entityToResponse(storefront), nil
}

// GetStorefrontByDomain retrieves storefront by domain
func (ss *StorefrontServiceSimple) GetStorefrontByDomain(ctx context.Context, domain string) (*dto.StorefrontResponse, error) {
	storefront, err := ss.storefrontRepo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return nil, errors.NewNotFoundError("storefront not found", nil)
	}

	return ss.entityToResponse(storefront), nil
}

// UpdateStorefront updates storefront information
func (ss *StorefrontServiceSimple) UpdateStorefront(ctx context.Context, storefrontID uuid.UUID, req *dto.StorefrontUpdateRequest) (*dto.StorefrontResponse, error) {
	// Get existing storefront
	storefront, err := ss.storefrontRepo.GetByID(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return nil, errors.NewNotFoundError("storefront not found", nil)
	}

	// Update fields if provided
	if req.Name != nil {
		storefront.Name = *req.Name
	}
	if req.Description != nil {
		storefront.Description = req.Description
	}
	if req.Domain != nil {
		// Check domain uniqueness if changing
		if storefront.Domain == nil || *storefront.Domain != *req.Domain {
			if *req.Domain != "" {
				existingStorefront, err := ss.storefrontRepo.GetByDomain(ctx, *req.Domain)
				if err == nil && existingStorefront != nil && existingStorefront.ID != storefrontID {
					return nil, errors.NewValidationError("domain already exists", nil)
				}
			}
		}
		storefront.Domain = req.Domain
	}
	if req.Subdomain != nil {
		storefront.Subdomain = req.Subdomain
	}

	storefront.UpdatedAt = time.Now()

	// Validate updated storefront
	if err := storefront.Validate(); err != nil {
		return nil, fmt.Errorf("storefront validation failed: %w", err)
	}

	// Save updated storefront
	if err := ss.storefrontRepo.Update(ctx, storefront); err != nil {
		return nil, fmt.Errorf("failed to update storefront: %w", err)
	}

	return ss.entityToResponse(storefront), nil
}

// SearchStorefronts searches storefronts with basic pagination
func (ss *StorefrontServiceSimple) SearchStorefronts(ctx context.Context, req *dto.StorefrontSearchRequest) (*dto.PaginatedStorefrontResponse, error) {
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
	searchParams := &repository.StorefrontSearchParams{
		Query:     req.Query,
		Page:      req.Page,
		PageSize:  req.PageSize,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	if req.Status != nil {
		searchParams.Status = *req.Status
	}

	// Perform search
	storefronts, total, err := ss.storefrontRepo.Search(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search storefronts: %w", err)
	}

	// Convert to response DTOs
	storefrontResponses := make([]*dto.StorefrontResponse, len(storefronts))
	for i, storefront := range storefronts {
		storefrontResponses[i] = ss.entityToResponse(storefront)
	}

	return &dto.PaginatedStorefrontResponse{
		Storefronts: storefrontResponses,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}, nil
}

// DeleteStorefront soft deletes a storefront
func (ss *StorefrontServiceSimple) DeleteStorefront(ctx context.Context, storefrontID uuid.UUID) error {
	// Get storefront to ensure it exists
	storefront, err := ss.storefrontRepo.GetByID(ctx, storefrontID)
	if err != nil {
		return fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return errors.NewNotFoundError("storefront not found", nil)
	}

	// Soft delete
	return ss.storefrontRepo.Delete(ctx, storefrontID)
}

// ActivateStorefront activates a storefront
func (ss *StorefrontServiceSimple) ActivateStorefront(ctx context.Context, storefrontID uuid.UUID) error {
	return ss.updateStorefrontStatus(ctx, storefrontID, entity.StorefrontStatusActive)
}

// DeactivateStorefront deactivates a storefront
func (ss *StorefrontServiceSimple) DeactivateStorefront(ctx context.Context, storefrontID uuid.UUID) error {
	return ss.updateStorefrontStatus(ctx, storefrontID, entity.StorefrontStatusInactive)
}

// SuspendStorefront suspends a storefront
func (ss *StorefrontServiceSimple) SuspendStorefront(ctx context.Context, storefrontID uuid.UUID) error {
	return ss.updateStorefrontStatus(ctx, storefrontID, entity.StorefrontStatusSuspended)
}

// ValidateDomain validates domain availability
func (ss *StorefrontServiceSimple) ValidateDomain(ctx context.Context, domain string) (*dto.DomainValidationResponse, error) {
	// Basic domain format validation
	if domain == "" {
		return &dto.DomainValidationResponse{
			Domain:    domain,
			Valid:     false,
			Available: false,
			Message:   "Domain cannot be empty",
			CheckedAt: time.Now(),
		}, nil
	}

	// Check if domain already exists
	existingStorefront, err := ss.storefrontRepo.GetByDomain(ctx, domain)
	if err != nil && !errors.IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to check domain availability: %w", err)
	}

	available := existingStorefront == nil
	message := "Domain is available"
	if !available {
		message = "Domain is already taken"
	}

	return &dto.DomainValidationResponse{
		Domain:    domain,
		Valid:     true, // Basic format validation passed
		Available: available,
		Message:   message,
		CheckedAt: time.Now(),
	}, nil
}

// GetStorefrontStats returns basic storefront statistics
func (ss *StorefrontServiceSimple) GetStorefrontStats(ctx context.Context, storefrontID uuid.UUID) (*dto.StorefrontStatsResponse, error) {
	// Get storefront to ensure it exists
	storefront, err := ss.storefrontRepo.GetByID(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return nil, errors.NewNotFoundError("storefront not found", nil)
	}

	stats, err := ss.storefrontRepo.GetStats(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront stats: %w", err)
	}

	return &dto.StorefrontStatsResponse{
		StorefrontID:   storefrontID,
		TotalCustomers: stats.TotalCustomers,
		TotalOrders:    stats.TotalOrders,
		TotalRevenue:   stats.TotalRevenue,
		ActiveProducts: stats.ActiveProducts,
		ConversionRate: stats.ConversionRate,
		Period:         "all_time",
		Timestamp:      time.Now(),
	}, nil
}

// ConfigureDomain configures domain settings for a storefront
func (ss *StorefrontServiceSimple) ConfigureDomain(ctx context.Context, storefrontID uuid.UUID, req *dto.DomainConfigRequest) error {
	// Get existing storefront
	storefront, err := ss.storefrontRepo.GetByID(ctx, storefrontID)
	if err != nil {
		return fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return errors.NewNotFoundError("storefront not found", nil)
	}

	// Check domain uniqueness
	if req.Domain != "" {
		existingStorefront, err := ss.storefrontRepo.GetByDomain(ctx, req.Domain)
		if err == nil && existingStorefront != nil && existingStorefront.ID != storefrontID {
			return errors.NewValidationError("domain already exists", nil)
		}
	}

	// Update domain settings
	storefront.Domain = &req.Domain
	if req.Subdomain != nil {
		storefront.Subdomain = req.Subdomain
	}
	storefront.UpdatedAt = time.Now()

	// Save updated storefront
	if err := ss.storefrontRepo.Update(ctx, storefront); err != nil {
		return fmt.Errorf("failed to update storefront: %w", err)
	}

	return nil
}

// Helper methods

func (ss *StorefrontServiceSimple) updateStorefrontStatus(ctx context.Context, storefrontID uuid.UUID, status entity.StorefrontStatus) error {
	// Get existing storefront
	storefront, err := ss.storefrontRepo.GetByID(ctx, storefrontID)
	if err != nil {
		return fmt.Errorf("failed to get storefront: %w", err)
	}

	if storefront == nil {
		return errors.NewNotFoundError("storefront not found", nil)
	}

	// Update status
	storefront.Status = status
	storefront.UpdatedAt = time.Now()

	// Save updated storefront
	if err := ss.storefrontRepo.Update(ctx, storefront); err != nil {
		return fmt.Errorf("failed to update storefront status: %w", err)
	}

	return nil
}

func (ss *StorefrontServiceSimple) entityToResponse(storefront *entity.Storefront) *dto.StorefrontResponse {
	return &dto.StorefrontResponse{
		ID:          storefront.ID,
		Name:        storefront.Name,
		Slug:        storefront.Slug,
		Description: storefront.Description,
		Domain:      storefront.Domain,
		Subdomain:   storefront.Subdomain,
		Status:      string(storefront.Status),
		Settings:    storefront.Settings,
		CreatedAt:   storefront.CreatedAt,
		UpdatedAt:   storefront.UpdatedAt,
	}
}
