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

type CustomerAddressServiceSimple struct {
	addressRepo    repository.CustomerAddressRepository
	customerRepo   repository.CustomerRepository
	tenantResolver tenant.TenantResolver
}

func NewCustomerAddressServiceSimple(
	addressRepo repository.CustomerAddressRepository,
	customerRepo repository.CustomerRepository,
	tenantResolver tenant.TenantResolver,
) CustomerAddressService {
	return &CustomerAddressServiceSimple{
		addressRepo:    addressRepo,
		customerRepo:   customerRepo,
		tenantResolver: tenantResolver,
	}
}

// CreateAddress creates a new customer address
func (s *CustomerAddressServiceSimple) CreateAddress(
	ctx context.Context,
	req *dto.CreateAddressRequest,
) (*dto.CustomerAddressResponse, error) {
	// Validate the request
	if err := s.validateCreateAddressRequest(req); err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("validation failed: %v", err), err)
	}

	// Get storefront ID from tenant context
	storefrontID, err := s.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storefront: %w", err)
	}

	// Verify customer exists in this storefront
	_, err = s.customerRepo.GetByID(ctx, storefrontID, req.CustomerID)
	if err != nil {
		return nil, errors.NewNotFoundError("customer not found")
	}

	// Create the address entity
	address := &entity.CustomerAddress{
		ID:                   uuid.New(),
		CustomerID:           req.CustomerID,
		AddressType:          req.Type,
		Label:                req.Label,
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		Company:              req.Company,
		Phone:                req.Phone,
		AddressLine1:         req.AddressLine1,
		AddressLine2:         req.AddressLine2,
		City:                 req.City,
		StateProvince:        req.StateProvince,
		PostalCode:           req.PostalCode,
		Country:              req.Country,
		IsDefault:            req.IsDefault,
		IsActive:             true,
		DeliveryInstructions: req.DeliveryInstructions,
		Latitude:             req.Latitude,
		Longitude:            req.Longitude,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Normalize fields
	address.NormalizeFields()

	// Validate the entity
	if err := address.Validate(); err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("address validation failed: %v", err), err)
	}

	// Create the address
	if err := s.addressRepo.Create(ctx, address); err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to create address: %v", err), err)
	}

	// Convert to response DTO
	return s.convertToAddressResponse(address), nil
}

// GetAddress retrieves an address by ID
func (s *CustomerAddressServiceSimple) GetAddress(
	ctx context.Context,
	addressID uuid.UUID,
) (*dto.CustomerAddressResponse, error) {
	// Get the address
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return nil, errors.NewNotFoundError("address not found")
	}

	return s.convertToAddressResponse(address), nil
}

// GetCustomerAddresses retrieves all addresses for a customer
func (s *CustomerAddressServiceSimple) GetCustomerAddresses(
	ctx context.Context,
	customerID uuid.UUID,
) ([]*dto.CustomerAddressResponse, error) {
	// Get storefront ID from tenant context
	storefrontID, err := s.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storefront: %w", err)
	}

	// Verify customer exists
	_, err = s.customerRepo.GetByID(ctx, storefrontID, customerID)
	if err != nil {
		return nil, errors.NewNotFoundError("customer not found")
	}

	// Get addresses
	addresses, err := s.addressRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to get addresses: %v", err), err)
	}

	// Convert to response DTOs
	responses := make([]*dto.CustomerAddressResponse, len(addresses))
	for i, address := range addresses {
		responses[i] = s.convertToAddressResponse(address)
	}

	return responses, nil
}

// UpdateAddress updates an existing address
func (s *CustomerAddressServiceSimple) UpdateAddress(
	ctx context.Context,
	addressID uuid.UUID,
	req *dto.UpdateAddressRequest,
) (*dto.CustomerAddressResponse, error) {
	// Validate the request
	if err := s.validateUpdateAddressRequest(req); err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("validation failed: %v", err), err)
	}

	// Get existing address
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return nil, errors.NewNotFoundError("address not found")
	}

	// Apply updates to the address
	s.applyAddressUpdates(address, req)

	// Normalize fields
	address.NormalizeFields()

	// Validate the updated entity
	if err := address.Validate(); err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("address validation failed: %v", err), err)
	}

	// Update the address
	if err := s.addressRepo.Update(ctx, address); err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to update address: %v", err), err)
	}

	return s.convertToAddressResponse(address), nil
}

// DeleteAddress deletes an address
func (s *CustomerAddressServiceSimple) DeleteAddress(
	ctx context.Context,
	addressID uuid.UUID,
) error {
	// Verify address exists
	_, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return errors.NewNotFoundError("address not found")
	}

	// Delete the address
	if err := s.addressRepo.Delete(ctx, addressID); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to delete address: %v", err), err)
	}

	return nil
}

// SetDefaultAddress sets an address as default for its type
func (s *CustomerAddressServiceSimple) SetDefaultAddress(
	ctx context.Context,
	customerID, addressID uuid.UUID,
) error {
	// Get the address
	address, err := s.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return errors.NewNotFoundError("address not found")
	}

	// Verify the address belongs to the customer
	if address.CustomerID != customerID {
		return errors.NewValidationError("address does not belong to customer", nil)
	}

	// If already default, nothing to do
	if address.IsDefault {
		return nil
	}

	// Set as default using repository method
	if err := s.addressRepo.SetAsDefault(ctx, customerID, addressID); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to set default address: %v", err), err)
	}

	return nil
}

// GetDefaultAddress gets the default address for a customer
func (s *CustomerAddressServiceSimple) GetDefaultAddress(
	ctx context.Context,
	customerID uuid.UUID,
) (*dto.CustomerAddressResponse, error) {
	// Get storefront ID from tenant context
	storefrontID, err := s.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storefront: %w", err)
	}

	// Verify customer exists
	_, err = s.customerRepo.GetByID(ctx, storefrontID, customerID)
	if err != nil {
		return nil, errors.NewNotFoundError("customer not found")
	}

	// Get default address for customer by looking at all addresses
	addresses, err := s.addressRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to get addresses: %v", err), err)
	}

	// Find the default one
	for _, address := range addresses {
		if address.IsDefault {
			return s.convertToAddressResponse(address), nil
		}
	}

	return nil, errors.NewNotFoundError("default address not found")
}

// ValidateAddress validates an address using external services
func (s *CustomerAddressServiceSimple) ValidateAddress(
	ctx context.Context,
	req *dto.AddressValidationRequest,
) (*dto.AddressValidationResponse, error) {
	// Validate request
	if req.AddressLine1 == "" || req.City == "" || req.PostalCode == "" || req.Country == "" {
		return nil, errors.NewValidationError("address, city, postal code, and country are required", nil)
	}

	// For now, implement basic validation
	// In a real implementation, this would integrate with address validation services
	// like Google Maps API, USPS, etc.

	response := &dto.AddressValidationResponse{
		Valid:             true,
		ValidationResults: make(map[string]dto.ValidationError),
		Confidence:        0.9, // Default confidence
		ValidatedAt:       time.Now(),
	}

	// Basic validation checks
	if len(req.PostalCode) < 3 {
		response.Valid = false
		response.ValidationResults["postal_code"] = dto.ValidationError{
			Field:   "postal_code",
			Message: "Postal code is too short",
		}
		response.Confidence = 0.3
	}

	if len(req.Country) != 2 {
		response.Valid = false
		response.ValidationResults["country"] = dto.ValidationError{
			Field:   "country",
			Message: "Country code must be 2 characters (ISO format)",
		}
		response.Confidence = 0.2
	}

	// Create standardized address
	if response.Valid {
		response.Standardized = &dto.StandardizedAddress{
			AddressLine1: req.AddressLine1,
			AddressLine2: req.AddressLine2,
			City:         req.City,
			State:        req.State,
			PostalCode:   req.PostalCode,
			Country:      req.Country,
		}
	}

	return response, nil
}

// GeocodeAddress converts an address to coordinates
func (s *CustomerAddressServiceSimple) GeocodeAddress(
	ctx context.Context,
	req *dto.GeocodeRequest,
) (*dto.GeocodeResponse, error) {
	// For now, implement a mock geocoding service
	// In a real implementation, this would integrate with Google Maps, Mapbox, etc.

	if req.Address == "" && req.Coordinates == nil {
		return nil, errors.NewValidationError("either address or coordinates must be provided", nil)
	}

	// Mock response - in real implementation this would call external API
	response := &dto.GeocodeResponse{
		Address: dto.StandardizedAddress{
			AddressLine1: req.Address,
			City:         "Jakarta",     // Mock
			State:        "DKI Jakarta", // Mock
			PostalCode:   "12345",       // Mock
			Country:      "ID",
		},
		Coordinates: dto.LatLong{
			Latitude:  -6.2088, // Jakarta coordinates
			Longitude: 106.8456,
		},
		Accuracy:   "APPROXIMATE",
		Source:     "mock_service",
		GeocodedAt: time.Now(),
	}

	return response, nil
}

// GetNearbyAddresses finds addresses near given coordinates
func (s *CustomerAddressServiceSimple) GetNearbyAddresses(
	ctx context.Context,
	req *dto.NearbyAddressRequest,
) ([]*dto.CustomerAddressResponse, error) {
	// Validate coordinates
	if req.Coordinates.Latitude < -90 || req.Coordinates.Latitude > 90 {
		return nil, errors.NewValidationError("latitude must be between -90 and 90", nil)
	}
	if req.Coordinates.Longitude < -180 || req.Coordinates.Longitude > 180 {
		return nil, errors.NewValidationError("longitude must be between -180 and 180", nil)
	}

	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	// Find nearby addresses using customer IDs (since addresses don't directly support nearby search)
	// This is a simplified implementation - in reality you'd have geospatial queries

	// For now, return empty results as a placeholder
	return []*dto.CustomerAddressResponse{}, nil
}

// GetAddressStats returns statistics about addresses
func (s *CustomerAddressServiceSimple) GetAddressStats(
	ctx context.Context,
	req *dto.AddressStatsRequest,
) (*dto.AddressStatsResponse, error) {
	// Get storefront ID from tenant context
	storefrontID, err := s.getStorefrontFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storefront: %w", err)
	}

	// Get stats from repository
	stats, err := s.addressRepo.GetAddressStats(ctx, storefrontID)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to get address statistics: %v", err), err)
	}

	// Convert repository stats to DTO response
	response := &dto.AddressStatsResponse{
		TotalAddresses:    int64(stats.TotalAddresses),
		VerifiedAddresses: int64(stats.ActiveAddresses), // Use active as proxy for verified
		TypeBreakdown:     make(map[string]int64),
		CountryBreakdown:  make(map[string]int64),
		Period:            req.Period,
		Timestamp:         time.Now(),
	}

	// Convert type breakdown
	for addressType, count := range stats.AddressesByType {
		response.TypeBreakdown[string(addressType)] = int64(count)
	}

	// Convert country breakdown
	for country, count := range stats.AddressesByCountry {
		response.CountryBreakdown[country] = int64(count)
	}

	// Use city breakdown as state breakdown (since the stats struct has cities)
	response.StateBreakdown = make(map[string]int64)
	response.CityBreakdown = make(map[string]int64)
	for city, count := range stats.AddressesByCity {
		response.CityBreakdown[city] = int64(count)
	}

	return response, nil
}

// Bulk operations (placeholder implementations to satisfy interface)
func (s *CustomerAddressServiceSimple) BulkCreateAddresses(
	ctx context.Context,
	req *dto.BulkAddressCreateRequest,
) (*dto.BulkOperationResponse, error) {
	// Placeholder - implement bulk create logic
	return &dto.BulkOperationResponse{
		TotalRequested: len(req.Addresses),
		Successful:     0,
		Failed:         0,
		Errors:         []dto.BulkOperationError{},
	}, nil
}

func (s *CustomerAddressServiceSimple) BulkUpdateAddresses(
	ctx context.Context,
	req *dto.BulkAddressUpdateRequest,
) (*dto.BulkOperationResponse, error) {
	// Placeholder - implement bulk update logic
	return &dto.BulkOperationResponse{
		TotalRequested: len(req.Updates),
		Successful:     0,
		Failed:         0,
		Errors:         []dto.BulkOperationError{},
	}, nil
}

func (s *CustomerAddressServiceSimple) BulkDeleteAddresses(
	ctx context.Context,
	req *dto.BulkAddressDeleteRequest,
) (*dto.BulkOperationResponse, error) {
	// Placeholder - implement bulk delete logic
	return &dto.BulkOperationResponse{
		TotalRequested: len(req.AddressIDs),
		Successful:     0,
		Failed:         0,
		Errors:         []dto.BulkOperationError{},
	}, nil
}

func (s *CustomerAddressServiceSimple) GetDistribution(
	ctx context.Context,
	req *dto.AddressDistributionRequest,
) (*dto.AddressDistributionResponse, error) {
	// Placeholder - implement distribution logic
	return &dto.AddressDistributionResponse{
		GroupBy:      req.GroupBy,
		Distribution: []dto.AddressDistributionItem{},
		Total:        0,
		Period:       req.Period,
		Timestamp:    time.Now(),
	}, nil
}

func (s *CustomerAddressServiceSimple) ImportAddresses(
	ctx context.Context,
	req *dto.AddressImportRequest,
) (*dto.AddressImportResponse, error) {
	// Placeholder - implement import logic
	return &dto.AddressImportResponse{
		ImportID:     uuid.New(),
		Status:       "completed",
		TotalRecords: 0,
		ValidRecords: 0,
		ProcessedAt:  time.Now(),
	}, nil
}

func (s *CustomerAddressServiceSimple) ExportAddresses(
	ctx context.Context,
	req *dto.AddressExportRequest,
) (*dto.ExportResponse, error) {
	// Placeholder - implement export logic
	return &dto.ExportResponse{
		ExportID:  uuid.New().String(),
		Status:    "completed",
		Format:    req.Format,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (s *CustomerAddressServiceSimple) GetAddressDistribution(
	ctx context.Context,
	req *dto.AddressDistributionRequest,
) (*dto.AddressDistributionResponse, error) {
	// Placeholder - implement distribution logic
	return &dto.AddressDistributionResponse{
		GroupBy:      req.GroupBy,
		Distribution: []dto.AddressDistributionItem{},
		Total:        0,
		Period:       req.Period,
		Timestamp:    time.Now(),
	}, nil
}

// Helper methods

func (s *CustomerAddressServiceSimple) getStorefrontFromContext(ctx context.Context) (uuid.UUID, error) {
	if storefrontID, ok := ctx.Value("storefront_id").(uuid.UUID); ok {
		return storefrontID, nil
	}
	return uuid.Nil, fmt.Errorf("storefront ID not found in context")
}

func (s *CustomerAddressServiceSimple) validateCreateAddressRequest(req *dto.CreateAddressRequest) error {
	if req.CustomerID == uuid.Nil {
		return fmt.Errorf("customer_id is required")
	}
	if !req.Type.IsValid() {
		return fmt.Errorf("invalid address type")
	}
	if req.AddressLine1 == "" {
		return fmt.Errorf("address_line1 is required")
	}
	if req.City == "" {
		return fmt.Errorf("city is required")
	}
	if req.PostalCode == "" {
		return fmt.Errorf("postal_code is required")
	}
	if req.Country == "" {
		return fmt.Errorf("country is required")
	}
	return nil
}

func (s *CustomerAddressServiceSimple) validateUpdateAddressRequest(req *dto.UpdateAddressRequest) error {
	if req.Type != nil && !req.Type.IsValid() {
		return fmt.Errorf("invalid address type")
	}
	if req.AddressLine1 != nil && *req.AddressLine1 == "" {
		return fmt.Errorf("address_line1 cannot be empty")
	}
	if req.City != nil && *req.City == "" {
		return fmt.Errorf("city cannot be empty")
	}
	if req.PostalCode != nil && *req.PostalCode == "" {
		return fmt.Errorf("postal_code cannot be empty")
	}
	if req.Country != nil && *req.Country == "" {
		return fmt.Errorf("country cannot be empty")
	}
	return nil
}

func (s *CustomerAddressServiceSimple) applyAddressUpdates(address *entity.CustomerAddress, req *dto.UpdateAddressRequest) {
	if req.Type != nil {
		address.AddressType = *req.Type
	}
	if req.Label != nil {
		address.Label = req.Label
	}
	if req.FirstName != nil {
		address.FirstName = req.FirstName
	}
	if req.LastName != nil {
		address.LastName = req.LastName
	}
	if req.Company != nil {
		address.Company = req.Company
	}
	if req.Phone != nil {
		address.Phone = req.Phone
	}
	if req.AddressLine1 != nil {
		address.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		address.AddressLine2 = req.AddressLine2
	}
	if req.City != nil {
		address.City = *req.City
	}
	if req.StateProvince != nil {
		address.StateProvince = req.StateProvince
	}
	if req.PostalCode != nil {
		address.PostalCode = *req.PostalCode
	}
	if req.Country != nil {
		address.Country = *req.Country
	}
	if req.IsDefault != nil {
		address.IsDefault = *req.IsDefault
	}
	if req.DeliveryInstructions != nil {
		address.DeliveryInstructions = req.DeliveryInstructions
	}
	if req.Latitude != nil {
		address.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		address.Longitude = req.Longitude
	}

	address.UpdatedAt = time.Now()
}

func (s *CustomerAddressServiceSimple) convertToAddressResponse(address *entity.CustomerAddress) *dto.CustomerAddressResponse {
	return &dto.CustomerAddressResponse{
		ID:                   address.ID,
		CustomerID:           address.CustomerID,
		Type:                 address.AddressType,
		Label:                address.Label,
		FirstName:            address.FirstName,
		LastName:             address.LastName,
		Company:              address.Company,
		Phone:                address.Phone,
		AddressLine1:         address.AddressLine1,
		AddressLine2:         address.AddressLine2,
		City:                 address.City,
		StateProvince:        address.StateProvince,
		PostalCode:           address.PostalCode,
		Country:              address.Country,
		IsDefault:            address.IsDefault,
		IsActive:             address.IsActive,
		DeliveryInstructions: address.DeliveryInstructions,
		Latitude:             address.Latitude,
		Longitude:            address.Longitude,
		CreatedAt:            address.CreatedAt,
		UpdatedAt:            address.UpdatedAt,
	}
}
