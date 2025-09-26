package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// ValidationServiceSimple provides simple validation logic
type ValidationServiceSimple struct {
	customerRepo   repository.CustomerRepository
	storefrontRepo repository.StorefrontRepository
	addressRepo    repository.CustomerAddressRepository
}

// NewValidationServiceSimple creates a new simple validation service
func NewValidationServiceSimple(
	customerRepo repository.CustomerRepository,
	storefrontRepo repository.StorefrontRepository,
	addressRepo repository.CustomerAddressRepository,
) ValidationService {
	return &ValidationServiceSimple{
		customerRepo:   customerRepo,
		storefrontRepo: storefrontRepo,
		addressRepo:    addressRepo,
	}
}

// Customer Validation

func (v *ValidationServiceSimple) ValidateCustomerRegistration(
	ctx context.Context,
	req *dto.CustomerRegistrationRequest,
) *ValidationResult {
	result := NewValidationResult()

	// Basic required field validations
	if req.Email == "" {
		result.AddError("email", "Email is required", "required", req.Email)
	} else if !v.isValidEmail(req.Email) {
		result.AddError("email", "Invalid email format", "format", req.Email)
	}

	if req.FirstName == "" {
		result.AddError("first_name", "First name is required", "required", req.FirstName)
	}

	if req.LastName == "" {
		result.AddError("last_name", "Last name is required", "required", req.LastName)
	}

	return result
}

func (v *ValidationServiceSimple) ValidateCustomerUpdate(
	ctx context.Context,
	req *dto.CustomerUpdateRequest,
) *ValidationResult {
	result := NewValidationResult()

	// Email format validation if provided
	if req.Email != nil && *req.Email != "" && !v.isValidEmail(*req.Email) {
		result.AddError("email", "Invalid email format", "format", *req.Email)
	}

	return result
}

func (v *ValidationServiceSimple) ValidateEmailUniqueness(
	ctx context.Context,
	email string,
	excludeCustomerID *uuid.UUID,
) error {
	// Basic uniqueness check - placeholder implementation
	return nil
}

func (v *ValidationServiceSimple) ValidatePhoneUniqueness(
	ctx context.Context,
	phone string,
	excludeCustomerID *uuid.UUID,
) error {
	// Basic uniqueness check - placeholder implementation
	return nil
}

// Storefront Validation

func (v *ValidationServiceSimple) ValidateStorefrontCreation(
	ctx context.Context,
	req *dto.StorefrontCreateRequest,
) *ValidationResult {
	result := NewValidationResult()

	// Basic required field validations
	if req.Name == "" {
		result.AddError("name", "Storefront name is required", "required", req.Name)
	}

	if req.Slug == "" {
		result.AddError("slug", "Slug is required", "required", req.Slug)
	} else if !v.isValidSlug(req.Slug) {
		result.AddError("slug", "Invalid slug format", "format", req.Slug)
	}

	return result
}

func (v *ValidationServiceSimple) ValidateStorefrontUpdate(
	ctx context.Context,
	req *dto.StorefrontUpdateRequest,
) *ValidationResult {
	result := NewValidationResult()

	// Basic validation for provided fields
	if req.Name != nil && *req.Name == "" {
		result.AddError("name", "Storefront name cannot be empty", "required", *req.Name)
	}

	return result
}

func (v *ValidationServiceSimple) ValidateSlugUniqueness(
	ctx context.Context,
	slug string,
	excludeStorefrontID *uuid.UUID,
) error {
	// Placeholder implementation
	return nil
}

func (v *ValidationServiceSimple) ValidateDomainUniqueness(
	ctx context.Context,
	domain string,
	excludeStorefrontID *uuid.UUID,
) error {
	// Placeholder implementation
	return nil
}

// Address Validation

func (v *ValidationServiceSimple) ValidateAddressCreation(
	ctx context.Context,
	req *dto.CreateAddressRequest,
) *ValidationResult {
	result := NewValidationResult()

	// Basic required field validations
	if req.CustomerID == uuid.Nil {
		result.AddError("customer_id", "Customer ID is required", "required", req.CustomerID.String())
	}

	if req.AddressLine1 == "" {
		result.AddError("address_line1", "Address line 1 is required", "required", req.AddressLine1)
	}

	if req.City == "" {
		result.AddError("city", "City is required", "required", req.City)
	}

	if req.PostalCode == "" {
		result.AddError("postal_code", "Postal code is required", "required", req.PostalCode)
	}

	if req.Country == "" {
		result.AddError("country", "Country is required", "required", req.Country)
	}

	return result
}

func (v *ValidationServiceSimple) ValidateAddressUpdate(
	ctx context.Context,
	req *dto.UpdateAddressRequest,
) *ValidationResult {
	result := NewValidationResult()

	// Basic validation for non-empty fields
	if req.AddressLine1 != nil && *req.AddressLine1 == "" {
		result.AddError("address_line1", "Address line 1 cannot be empty", "required", *req.AddressLine1)
	}

	if req.City != nil && *req.City == "" {
		result.AddError("city", "City cannot be empty", "required", *req.City)
	}

	return result
}

func (v *ValidationServiceSimple) ValidateAddressFormat(
	ctx context.Context,
	address *entity.CustomerAddress,
) *ValidationResult {
	result := NewValidationResult()

	// Use entity validation
	if err := address.Validate(); err != nil {
		result.AddError("address", err.Error(), "format", address)
	}

	return result
}

// Business Rule Validation

func (v *ValidationServiceSimple) ValidateBusinessRules(
	ctx context.Context,
	entityName string,
	operation string,
	data interface{},
) error {
	// Placeholder for business rule validation
	return nil
}

func (v *ValidationServiceSimple) ValidateTenantConstraints(
	ctx context.Context,
	storefrontID uuid.UUID,
	operation string,
	data interface{},
) error {
	// Placeholder for tenant constraint validation
	return nil
}

// Helper methods

func (v *ValidationServiceSimple) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (v *ValidationServiceSimple) isValidSlug(slug string) bool {
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	return slugRegex.MatchString(slug) && !strings.HasPrefix(slug, "-") && !strings.HasSuffix(slug, "-")
}
