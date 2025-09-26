package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// Customer Address Management DTOs

// CreateAddressRequest represents an address creation request
type CreateAddressRequest struct {
	CustomerID           uuid.UUID          `json:"customer_id" binding:"required"`
	Type                 entity.AddressType `json:"type" binding:"required"`
	Label                *string            `json:"label,omitempty"`
	FirstName            *string            `json:"first_name,omitempty"`
	LastName             *string            `json:"last_name,omitempty"`
	Company              *string            `json:"company,omitempty"`
	Phone                *string            `json:"phone,omitempty"`
	AddressLine1         string             `json:"address_line1" binding:"required,min=1,max=500"`
	AddressLine2         *string            `json:"address_line2,omitempty"`
	City                 string             `json:"city" binding:"required,min=1,max=255"`
	StateProvince        *string            `json:"state_province,omitempty"`
	PostalCode           string             `json:"postal_code" binding:"required,min=1,max=20"`
	Country              string             `json:"country" binding:"required,min=1,max=100"`
	IsDefault            bool               `json:"is_default"`
	DeliveryInstructions *string            `json:"delivery_instructions,omitempty"`
	Latitude             *float64           `json:"latitude,omitempty"`
	Longitude            *float64           `json:"longitude,omitempty"`
}

// UpdateAddressRequest represents an address update request
type UpdateAddressRequest struct {
	Type                 *entity.AddressType `json:"type,omitempty"`
	Label                *string             `json:"label,omitempty"`
	FirstName            *string             `json:"first_name,omitempty"`
	LastName             *string             `json:"last_name,omitempty"`
	Company              *string             `json:"company,omitempty"`
	Phone                *string             `json:"phone,omitempty"`
	AddressLine1         *string             `json:"address_line1,omitempty" binding:"omitempty,min=1,max=500"`
	AddressLine2         *string             `json:"address_line2,omitempty"`
	City                 *string             `json:"city,omitempty" binding:"omitempty,min=1,max=255"`
	StateProvince        *string             `json:"state_province,omitempty"`
	PostalCode           *string             `json:"postal_code,omitempty" binding:"omitempty,min=1,max=20"`
	Country              *string             `json:"country,omitempty" binding:"omitempty,min=1,max=100"`
	IsDefault            *bool               `json:"is_default,omitempty"`
	DeliveryInstructions *string             `json:"delivery_instructions,omitempty"`
	Latitude             *float64            `json:"latitude,omitempty"`
	Longitude            *float64            `json:"longitude,omitempty"`
}

// CustomerAddressResponse represents an address response
type CustomerAddressResponse struct {
	ID                   uuid.UUID          `json:"id"`
	CustomerID           uuid.UUID          `json:"customer_id"`
	Type                 entity.AddressType `json:"type"`
	Label                *string            `json:"label,omitempty"`
	FirstName            *string            `json:"first_name,omitempty"`
	LastName             *string            `json:"last_name,omitempty"`
	Company              *string            `json:"company,omitempty"`
	Phone                *string            `json:"phone,omitempty"`
	AddressLine1         string             `json:"address_line1"`
	AddressLine2         *string            `json:"address_line2,omitempty"`
	City                 string             `json:"city"`
	StateProvince        *string            `json:"state_province,omitempty"`
	PostalCode           string             `json:"postal_code"`
	Country              string             `json:"country"`
	IsDefault            bool               `json:"is_default"`
	IsActive             bool               `json:"is_active"`
	DeliveryInstructions *string            `json:"delivery_instructions,omitempty"`
	Latitude             *float64           `json:"latitude,omitempty"`
	Longitude            *float64           `json:"longitude,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

// Address Validation DTOs

// AddressValidationRequest represents an address validation request
type AddressValidationRequest struct {
	AddressLine1 string  `json:"address_line1" binding:"required"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         string  `json:"city" binding:"required"`
	State        string  `json:"state" binding:"required"`
	PostalCode   string  `json:"postal_code" binding:"required"`
	Country      string  `json:"country" binding:"required,min=2,max=2"`
}

// AddressValidationResponse represents an address validation response
type AddressValidationResponse struct {
	Valid             bool                       `json:"valid"`
	Standardized      *StandardizedAddress       `json:"standardized,omitempty"`
	Suggestions       []*AddressSuggestion       `json:"suggestions,omitempty"`
	ValidationResults map[string]ValidationError `json:"validation_results"`
	Confidence        float64                    `json:"confidence"`
	ValidatedAt       time.Time                  `json:"validated_at"`
}

// StandardizedAddress represents a standardized address
type StandardizedAddress struct {
	AddressLine1 string   `json:"address_line1"`
	AddressLine2 *string  `json:"address_line2,omitempty"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	PostalCode   string   `json:"postal_code"`
	Country      string   `json:"country"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

// AddressSuggestion represents an address suggestion
type AddressSuggestion struct {
	Address    StandardizedAddress `json:"address"`
	Confidence float64             `json:"confidence"`
	Source     string              `json:"source"`
}

// GeocodeRequest represents a geocoding request
type GeocodeRequest struct {
	Address     string   `json:"address" binding:"required"`
	Coordinates *LatLong `json:"coordinates,omitempty"`
}

// GeocodeResponse represents a geocoding response
type GeocodeResponse struct {
	Address     StandardizedAddress `json:"address"`
	Coordinates LatLong             `json:"coordinates"`
	Accuracy    string              `json:"accuracy"`
	Source      string              `json:"source"`
	GeocodedAt  time.Time           `json:"geocoded_at"`
}

// LatLong represents latitude and longitude coordinates
type LatLong struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Proximity and Search DTOs

// NearbyAddressRequest represents a nearby address search request
type NearbyAddressRequest struct {
	Coordinates LatLong             `json:"coordinates" binding:"required"`
	Radius      float64             `json:"radius" binding:"required,min=0.1,max=100"` // in kilometers
	Limit       int                 `json:"limit,omitempty" binding:"max=100"`
	Type        *entity.AddressType `json:"type,omitempty"`
}

// Bulk Operations DTOs

// BulkAddressCreateRequest represents a bulk address creation request
type BulkAddressCreateRequest struct {
	Addresses []CreateAddressRequest `json:"addresses" binding:"required,min=1,max=100"`
}

// BulkAddressUpdateRequest represents a bulk address update request
type BulkAddressUpdateRequest struct {
	Updates []AddressUpdateItem `json:"updates" binding:"required,min=1,max=100"`
}

// AddressUpdateItem represents an individual address update in bulk operation
type AddressUpdateItem struct {
	AddressID uuid.UUID            `json:"address_id" binding:"required"`
	Updates   UpdateAddressRequest `json:"updates" binding:"required"`
}

// BulkAddressDeleteRequest represents a bulk address deletion request
type BulkAddressDeleteRequest struct {
	AddressIDs []uuid.UUID `json:"address_ids" binding:"required,min=1,max=100"`
}

// OperationResult represents a single operation result
type OperationResult struct {
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Analytics and Statistics DTOs

// AddressStatsRequest represents an address statistics request
type AddressStatsRequest struct {
	Period    string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// AddressStatsResponse represents address statistics
type AddressStatsResponse struct {
	TotalAddresses    int64            `json:"total_addresses"`
	VerifiedAddresses int64            `json:"verified_addresses"`
	TypeBreakdown     map[string]int64 `json:"type_breakdown"`
	CountryBreakdown  map[string]int64 `json:"country_breakdown"`
	StateBreakdown    map[string]int64 `json:"state_breakdown"`
	CityBreakdown     map[string]int64 `json:"city_breakdown"`
	Period            string           `json:"period"`
	Timestamp         time.Time        `json:"timestamp"`
}

// AddressDistributionRequest represents an address distribution request
type AddressDistributionRequest struct {
	GroupBy   string     `json:"group_by" binding:"required,oneof=country state city type"`
	Period    string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty" binding:"max=100"`
}

// AddressDistributionResponse represents address distribution response
type AddressDistributionResponse struct {
	GroupBy      string                    `json:"group_by"`
	Distribution []AddressDistributionItem `json:"distribution"`
	Total        int64                     `json:"total"`
	Period       string                    `json:"period"`
	Timestamp    time.Time                 `json:"timestamp"`
}

// AddressDistributionItem represents a single distribution item
type AddressDistributionItem struct {
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// Shipping Integration DTOs (for future use)

// ShippingEstimateRequest represents a shipping estimate request
type ShippingEstimateRequest struct {
	FromAddressID uuid.UUID          `json:"from_address_id" binding:"required"`
	ToAddressID   uuid.UUID          `json:"to_address_id" binding:"required"`
	Weight        float64            `json:"weight" binding:"required,min=0.1"`
	Dimensions    *PackageDimensions `json:"dimensions,omitempty"`
	ServiceType   *string            `json:"service_type,omitempty"`
}

// PackageDimensions represents package dimensions
type PackageDimensions struct {
	Length float64 `json:"length" binding:"required,min=1"`
	Width  float64 `json:"width" binding:"required,min=1"`
	Height float64 `json:"height" binding:"required,min=1"`
	Unit   string  `json:"unit" binding:"required,oneof=cm in"`
}

// ShippingEstimateResponse represents a shipping estimate response
type ShippingEstimateResponse struct {
	FromAddress     CustomerAddressResponse `json:"from_address"`
	ToAddress       CustomerAddressResponse `json:"to_address"`
	ShippingOptions []ShippingOption        `json:"shipping_options"`
	EstimatedAt     time.Time               `json:"estimated_at"`
}

// ShippingOption represents a shipping option
type ShippingOption struct {
	Carrier           string     `json:"carrier"`
	ServiceType       string     `json:"service_type"`
	ServiceName       string     `json:"service_name"`
	Cost              float64    `json:"cost"`
	Currency          string     `json:"currency"`
	EstimatedDays     int        `json:"estimated_days"`
	DeliveryDate      *time.Time `json:"delivery_date,omitempty"`
	TrackingAvailable bool       `json:"tracking_available"`
}

// Address Import/Export DTOs

// AddressImportRequest represents an address import request
type AddressImportRequest struct {
	Format   string                 `json:"format" binding:"required,oneof=csv xlsx json"`
	Data     interface{}            `json:"data" binding:"required"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Validate bool                   `json:"validate"`
	DryRun   bool                   `json:"dry_run"`
}

// AddressImportResponse represents an address import response
type AddressImportResponse struct {
	ImportID        uuid.UUID            `json:"import_id"`
	Status          string               `json:"status"`
	TotalRecords    int                  `json:"total_records"`
	ValidRecords    int                  `json:"valid_records"`
	InvalidRecords  int                  `json:"invalid_records"`
	ImportedRecords int                  `json:"imported_records"`
	Errors          []AddressImportError `json:"errors,omitempty"`
	ProcessedAt     time.Time            `json:"processed_at"`
}

// AddressImportError represents an error during address import
type AddressImportError struct {
	Row     int         `json:"row"`
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
}

// AddressExportRequest represents an address export request
type AddressExportRequest struct {
	Format      string                 `json:"format" binding:"required,oneof=csv xlsx json"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Fields      []string               `json:"fields,omitempty"`
	CustomerIDs []uuid.UUID            `json:"customer_ids,omitempty"`
}
