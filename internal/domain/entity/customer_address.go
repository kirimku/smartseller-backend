package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AddressType represents the type of address
type AddressType string

const (
	AddressTypeBilling  AddressType = "billing"
	AddressTypeShipping AddressType = "shipping"
	AddressTypeBoth     AddressType = "both"
)

// IsValid checks if the address type is valid
func (a AddressType) IsValid() bool {
	switch a {
	case AddressTypeBilling, AddressTypeShipping, AddressTypeBoth:
		return true
	default:
		return false
	}
}

// CustomerAddress represents a customer's address
type CustomerAddress struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CustomerID uuid.UUID `json:"customer_id" db:"customer_id"`

	// Address type and identification
	AddressType AddressType `json:"address_type" db:"address_type"`
	Label       *string     `json:"label,omitempty" db:"label"` // e.g., "Home", "Office"

	// Contact person (may differ from customer)
	FirstName *string `json:"first_name,omitempty" db:"first_name"`
	LastName  *string `json:"last_name,omitempty" db:"last_name"`
	Company   *string `json:"company,omitempty" db:"company"`
	Phone     *string `json:"phone,omitempty" db:"phone"`

	// Address details
	AddressLine1  string  `json:"address_line_1" db:"address_line_1"`
	AddressLine2  *string `json:"address_line_2,omitempty" db:"address_line_2"`
	City          string  `json:"city" db:"city"`
	StateProvince *string `json:"state_province,omitempty" db:"state_province"`
	PostalCode    string  `json:"postal_code" db:"postal_code"`
	Country       string  `json:"country" db:"country"`

	// Address metadata
	IsDefault bool `json:"is_default" db:"is_default"`
	IsActive  bool `json:"is_active" db:"is_active"`

	// Coordinates (for delivery optimization)
	Latitude  *float64 `json:"latitude,omitempty" db:"latitude"`
	Longitude *float64 `json:"longitude,omitempty" db:"longitude"`

	// Delivery instructions
	DeliveryInstructions *string `json:"delivery_instructions,omitempty" db:"delivery_instructions"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate performs comprehensive validation of address data
func (a *CustomerAddress) Validate() error {
	if a.CustomerID == uuid.Nil {
		return fmt.Errorf("customer_id is required")
	}

	if !a.AddressType.IsValid() {
		return fmt.Errorf("invalid address type")
	}

	// Required fields validation
	if strings.TrimSpace(a.AddressLine1) == "" {
		return fmt.Errorf("address line 1 is required")
	}

	if strings.TrimSpace(a.City) == "" {
		return fmt.Errorf("city is required")
	}

	if strings.TrimSpace(a.PostalCode) == "" {
		return fmt.Errorf("postal code is required")
	}

	if strings.TrimSpace(a.Country) == "" {
		return fmt.Errorf("country is required")
	}

	// Length validations
	if len(a.AddressLine1) > 500 {
		return fmt.Errorf("address line 1 cannot exceed 500 characters")
	}

	if a.AddressLine2 != nil && len(*a.AddressLine2) > 500 {
		return fmt.Errorf("address line 2 cannot exceed 500 characters")
	}

	if len(a.City) > 255 {
		return fmt.Errorf("city cannot exceed 255 characters")
	}

	if a.StateProvince != nil && len(*a.StateProvince) > 255 {
		return fmt.Errorf("state/province cannot exceed 255 characters")
	}

	if len(a.PostalCode) > 20 {
		return fmt.Errorf("postal code cannot exceed 20 characters")
	}

	if len(a.Country) > 100 {
		return fmt.Errorf("country cannot exceed 100 characters")
	}

	if a.Label != nil && len(*a.Label) > 100 {
		return fmt.Errorf("address label cannot exceed 100 characters")
	}

	if a.FirstName != nil && len(*a.FirstName) > 255 {
		return fmt.Errorf("first name cannot exceed 255 characters")
	}

	if a.LastName != nil && len(*a.LastName) > 255 {
		return fmt.Errorf("last name cannot exceed 255 characters")
	}

	if a.Company != nil && len(*a.Company) > 255 {
		return fmt.Errorf("company cannot exceed 255 characters")
	}

	if a.Phone != nil && *a.Phone != "" && !isValidPhoneForAddress(*a.Phone) {
		return fmt.Errorf("invalid phone format")
	}

	if a.DeliveryInstructions != nil && len(*a.DeliveryInstructions) > 1000 {
		return fmt.Errorf("delivery instructions cannot exceed 1000 characters")
	}

	// Coordinate validation
	if a.Latitude != nil {
		if *a.Latitude < -90 || *a.Latitude > 90 {
			return fmt.Errorf("latitude must be between -90 and 90")
		}
	}

	if a.Longitude != nil {
		if *a.Longitude < -180 || *a.Longitude > 180 {
			return fmt.Errorf("longitude must be between -180 and 180")
		}
	}

	return nil
}

// GetFullAddress returns a formatted full address string
func (a *CustomerAddress) GetFullAddress() string {
	parts := []string{}

	// Add address lines
	parts = append(parts, a.AddressLine1)
	if a.AddressLine2 != nil && *a.AddressLine2 != "" {
		parts = append(parts, *a.AddressLine2)
	}

	// Add city and state/province
	cityPart := a.City
	if a.StateProvince != nil && *a.StateProvince != "" {
		cityPart += ", " + *a.StateProvince
	}
	parts = append(parts, cityPart)

	// Add postal code and country
	parts = append(parts, a.PostalCode+" "+a.Country)

	return strings.Join(parts, ", ")
}

// GetFullName returns the full name of the contact person for this address
func (a *CustomerAddress) GetFullName() string {
	var parts []string

	if a.FirstName != nil && *a.FirstName != "" {
		parts = append(parts, *a.FirstName)
	}
	if a.LastName != nil && *a.LastName != "" {
		parts = append(parts, *a.LastName)
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

// GetDisplayName returns a user-friendly name for the address
func (a *CustomerAddress) GetDisplayName() string {
	if a.Label != nil && *a.Label != "" {
		return *a.Label
	}

	// Fallback to address type
	switch a.AddressType {
	case AddressTypeBilling:
		return "Billing Address"
	case AddressTypeShipping:
		return "Shipping Address"
	case AddressTypeBoth:
		return "Default Address"
	default:
		return "Address"
	}
}

// GetShortAddress returns a condensed version of the address
func (a *CustomerAddress) GetShortAddress() string {
	parts := []string{}

	// Use first few words of address line 1
	addressWords := strings.Fields(a.AddressLine1)
	if len(addressWords) > 3 {
		parts = append(parts, strings.Join(addressWords[:3], " ")+"...")
	} else {
		parts = append(parts, a.AddressLine1)
	}

	// Add city
	parts = append(parts, a.City)

	return strings.Join(parts, ", ")
}

// NormalizeFields trims whitespace and normalizes address fields
func (a *CustomerAddress) NormalizeFields() {
	a.AddressLine1 = strings.TrimSpace(a.AddressLine1)
	if a.AddressLine2 != nil {
		normalized := strings.TrimSpace(*a.AddressLine2)
		if normalized == "" {
			a.AddressLine2 = nil
		} else {
			a.AddressLine2 = &normalized
		}
	}

	a.City = strings.TrimSpace(a.City)
	a.PostalCode = strings.TrimSpace(a.PostalCode)
	a.Country = strings.TrimSpace(a.Country)

	if a.StateProvince != nil {
		normalized := strings.TrimSpace(*a.StateProvince)
		if normalized == "" {
			a.StateProvince = nil
		} else {
			a.StateProvince = &normalized
		}
	}

	if a.Label != nil {
		normalized := strings.TrimSpace(*a.Label)
		if normalized == "" {
			a.Label = nil
		} else {
			a.Label = &normalized
		}
	}

	if a.FirstName != nil {
		normalized := strings.TrimSpace(*a.FirstName)
		if normalized == "" {
			a.FirstName = nil
		} else {
			a.FirstName = &normalized
		}
	}

	if a.LastName != nil {
		normalized := strings.TrimSpace(*a.LastName)
		if normalized == "" {
			a.LastName = nil
		} else {
			a.LastName = &normalized
		}
	}

	if a.Company != nil {
		normalized := strings.TrimSpace(*a.Company)
		if normalized == "" {
			a.Company = nil
		} else {
			a.Company = &normalized
		}
	}

	if a.Phone != nil {
		normalized := strings.TrimSpace(*a.Phone)
		if normalized == "" {
			a.Phone = nil
		} else {
			a.Phone = &normalized
		}
	}

	if a.DeliveryInstructions != nil {
		normalized := strings.TrimSpace(*a.DeliveryInstructions)
		if normalized == "" {
			a.DeliveryInstructions = nil
		} else {
			a.DeliveryInstructions = &normalized
		}
	}
}

// IsBillingAddress checks if this is a billing address
func (a *CustomerAddress) IsBillingAddress() bool {
	return a.AddressType == AddressTypeBilling || a.AddressType == AddressTypeBoth
}

// IsShippingAddress checks if this is a shipping address
func (a *CustomerAddress) IsShippingAddress() bool {
	return a.AddressType == AddressTypeShipping || a.AddressType == AddressTypeBoth
}

// HasCoordinates checks if the address has latitude and longitude set
func (a *CustomerAddress) HasCoordinates() bool {
	return a.Latitude != nil && a.Longitude != nil
}

// IsInCountry checks if the address is in the specified country
func (a *CustomerAddress) IsInCountry(country string) bool {
	return strings.EqualFold(a.Country, country)
}

// IsInCity checks if the address is in the specified city
func (a *CustomerAddress) IsInCity(city string) bool {
	return strings.EqualFold(a.City, city)
}

// Helper function for phone validation specific to addresses
func isValidPhoneForAddress(phone string) bool {
	// More lenient phone validation for addresses - can include extensions
	if strings.TrimSpace(phone) == "" {
		return true // Allow empty phone
	}

	// Basic phone validation - allows various formats
	// This is simpler than customer phone validation to allow landlines, extensions, etc.
	return len(phone) >= 7 && len(phone) <= 20
}
