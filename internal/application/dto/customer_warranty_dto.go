package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CustomerWarrantyRegistrationRequest represents a customer warranty registration request
type CustomerWarrantyRegistrationRequest struct {
	BarcodeValue    string     `json:"barcode_value" binding:"required" example:"WB-2024-001234567"`
	ProductSKU      string     `json:"product_sku" binding:"required" example:"SKU-PHONE-001"`
	SerialNumber    string     `json:"serial_number" binding:"required" example:"SN123456789"`
	PurchaseDate    time.Time  `json:"purchase_date" binding:"required" example:"2024-01-15T00:00:00Z"`
	PurchasePrice   *decimal.Decimal `json:"purchase_price,omitempty" example:"999.99"`
	RetailerName    string     `json:"retailer_name" binding:"required" example:"TechStore Inc"`
	RetailerAddress string     `json:"retailer_address,omitempty" example:"123 Main St, City, State"`
	InvoiceNumber   string     `json:"invoice_number,omitempty" example:"INV-2024-001"`
	CustomerInfo    CustomerRegistrationInfo `json:"customer_info" binding:"required"`
	ProofOfPurchase *ProofOfPurchaseInfo `json:"proof_of_purchase,omitempty"`
}

// CustomerRegistrationInfo represents customer information for warranty registration
type CustomerRegistrationInfo struct {
	FirstName    string `json:"first_name" binding:"required" example:"John"`
	LastName     string `json:"last_name" binding:"required" example:"Doe"`
	Email        string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	PhoneNumber  string `json:"phone_number" binding:"required" example:"+1234567890"`
	Address      CustomerAddressInfo `json:"address" binding:"required"`
	DateOfBirth  *time.Time `json:"date_of_birth,omitempty" example:"1990-01-01T00:00:00Z"`
	Preferences  *CustomerPreferences `json:"preferences,omitempty"`
}

// CustomerAddressInfo represents customer address information
type CustomerAddressInfo struct {
	Street     string `json:"street" binding:"required" example:"123 Main Street"`
	City       string `json:"city" binding:"required" example:"New York"`
	State      string `json:"state" binding:"required" example:"NY"`
	PostalCode string `json:"postal_code" binding:"required" example:"10001"`
	Country    string `json:"country" binding:"required" example:"USA"`
}

// CustomerPreferences represents customer communication preferences
type CustomerPreferences struct {
	EmailNotifications bool   `json:"email_notifications" example:"true"`
	SMSNotifications   bool   `json:"sms_notifications" example:"false"`
	Language           string `json:"language" example:"en"`
	Timezone           string `json:"timezone" example:"America/New_York"`
}

// ProofOfPurchaseInfo represents proof of purchase information
type ProofOfPurchaseInfo struct {
	DocumentType string `json:"document_type" example:"receipt"`
	DocumentURL  string `json:"document_url" example:"https://example.com/receipts/receipt-123.pdf"`
	UploadedAt   time.Time `json:"uploaded_at" example:"2024-01-15T10:30:00Z"`
}

// CustomerWarrantyRegistrationResponse represents the response after warranty registration
type CustomerWarrantyRegistrationResponse struct {
	Success          bool                    `json:"success" example:"true"`
	RegistrationID   uuid.UUID               `json:"registration_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	WarrantyID       uuid.UUID               `json:"warranty_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	BarcodeValue     string                  `json:"barcode_value" example:"WB-2024-001234567"`
	Status           string                  `json:"status" example:"registered"`
	ActivationDate   time.Time               `json:"activation_date" example:"2024-01-15T10:30:00Z"`
	ExpiryDate       time.Time               `json:"expiry_date" example:"2026-01-15T10:30:00Z"`
	WarrantyPeriod   string                  `json:"warranty_period" example:"24 months"`
	Product          CustomerProductInfo     `json:"product"`
	Customer         CustomerInfo            `json:"customer"`
	Coverage         CustomerWarrantyCoverage `json:"coverage"`
	NextSteps        []string                `json:"next_steps"`
	RegistrationTime time.Time               `json:"registration_time" example:"2024-01-15T10:30:00Z"`
}

// CustomerProductInfo represents product information for customer
type CustomerProductInfo struct {
	ID          uuid.UUID        `json:"id" example:"123e4567-e89b-12d3-a456-426614174002"`
	SKU         string           `json:"sku" example:"SKU-PHONE-001"`
	Name        string           `json:"name" example:"Smartphone Pro Max 256GB"`
	Brand       string           `json:"brand" example:"TechBrand"`
	Model       string           `json:"model" example:"Pro Max"`
	Category    string           `json:"category" example:"Electronics"`
	Description string           `json:"description" example:"Latest flagship smartphone"`
	ImageURL    string           `json:"image_url" example:"https://example.com/images/phone.jpg"`
	Price       *decimal.Decimal `json:"price,omitempty" example:"999.99"`
}

// CustomerInfo represents customer information in responses
type CustomerInfo struct {
	ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174003"`
	FirstName   string    `json:"first_name" example:"John"`
	LastName    string    `json:"last_name" example:"Doe"`
	Email       string    `json:"email" example:"john.doe@example.com"`
	PhoneNumber string    `json:"phone_number" example:"+1234567890"`
}

// CustomerWarrantyCoverage represents warranty coverage information for customer
type CustomerWarrantyCoverage struct {
	CoverageType        string   `json:"coverage_type" example:"comprehensive"`
	CoveredComponents   []string `json:"covered_components" example:"hardware,software,battery"`
	ExcludedComponents  []string `json:"excluded_components" example:"water_damage,physical_abuse"`
	RepairCoverage      bool     `json:"repair_coverage" example:"true"`
	ReplacementCoverage bool     `json:"replacement_coverage" example:"true"`
	LaborCoverage       bool     `json:"labor_coverage" example:"true"`
	PartsCoverage       bool     `json:"parts_coverage" example:"true"`
	Terms               []string `json:"terms"`
	Limitations         []string `json:"limitations,omitempty"`
}

// CustomerWarrantyListRequest represents a request to list customer warranties
type CustomerWarrantyListRequest struct {
	CustomerID   *uuid.UUID `json:"customer_id,omitempty" form:"customer_id"`
	Email        string     `json:"email,omitempty" form:"email"`
	PhoneNumber  string     `json:"phone_number,omitempty" form:"phone_number"`
	Status       string     `json:"status,omitempty" form:"status" example:"active"`
	ProductSKU   string     `json:"product_sku,omitempty" form:"product_sku"`
	Page         int        `json:"page,omitempty" form:"page" example:"1"`
	Limit        int        `json:"limit,omitempty" form:"limit" example:"10"`
	SortBy       string     `json:"sort_by,omitempty" form:"sort_by" example:"created_at"`
	SortOrder    string     `json:"sort_order,omitempty" form:"sort_order" example:"desc"`
}

// CustomerWarrantyListResponse represents the response for listing customer warranties
type CustomerWarrantyListResponse struct {
	Warranties   []CustomerWarrantySummary `json:"warranties"`
	TotalCount   int                       `json:"total_count" example:"25"`
	Page         int                       `json:"page" example:"1"`
	Limit        int                       `json:"limit" example:"10"`
	TotalPages   int                       `json:"total_pages" example:"3"`
	HasNext      bool                      `json:"has_next" example:"true"`
	HasPrevious  bool                      `json:"has_previous" example:"false"`
	RequestTime  time.Time                 `json:"request_time" example:"2024-01-15T10:30:00Z"`
}

// CustomerWarrantySummary represents a summary of customer warranty
type CustomerWarrantySummary struct {
	ID               uuid.UUID           `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	BarcodeValue     string              `json:"barcode_value" example:"WB-2024-001234567"`
	Status           string              `json:"status" example:"active"`
	Product          CustomerProductInfo `json:"product"`
	ActivationDate   time.Time           `json:"activation_date" example:"2024-01-15T10:30:00Z"`
	ExpiryDate       time.Time           `json:"expiry_date" example:"2026-01-15T10:30:00Z"`
	DaysRemaining    int                 `json:"days_remaining" example:"365"`
	WarrantyPeriod   string              `json:"warranty_period" example:"24 months"`
	IsExpired        bool                `json:"is_expired" example:"false"`
	CanClaim         bool                `json:"can_claim" example:"true"`
	ClaimsCount      int                 `json:"claims_count" example:"0"`
	LastClaimDate    *time.Time          `json:"last_claim_date,omitempty"`
}

// CustomerWarrantyDetailResponse represents detailed warranty information for customer
type CustomerWarrantyDetailResponse struct {
	ID               uuid.UUID                `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	BarcodeValue     string                   `json:"barcode_value" example:"WB-2024-001234567"`
	Status           string                   `json:"status" example:"active"`
	Product          CustomerProductInfo      `json:"product"`
	Customer         CustomerInfo             `json:"customer"`
	ActivationDate   time.Time                `json:"activation_date" example:"2024-01-15T10:30:00Z"`
	ExpiryDate       time.Time                `json:"expiry_date" example:"2026-01-15T10:30:00Z"`
	DaysRemaining    int                      `json:"days_remaining" example:"365"`
	WarrantyPeriod   string                   `json:"warranty_period" example:"24 months"`
	IsExpired        bool                     `json:"is_expired" example:"false"`
	CanClaim         bool                     `json:"can_claim" example:"true"`
	Coverage         CustomerWarrantyCoverage `json:"coverage"`
	PurchaseInfo     PurchaseInfo             `json:"purchase_info"`
	ClaimsHistory    []CustomerClaimSummary   `json:"claims_history"`
	Documents        []WarrantyDocument       `json:"documents"`
	SupportInfo      CustomerSupportInfo      `json:"support_info"`
	RetrievalTime    time.Time                `json:"retrieval_time" example:"2024-01-15T10:30:00Z"`
}

// PurchaseInfo represents purchase information
type PurchaseInfo struct {
	PurchaseDate    time.Time        `json:"purchase_date" example:"2024-01-15T00:00:00Z"`
	PurchasePrice   *decimal.Decimal `json:"purchase_price,omitempty" example:"999.99"`
	RetailerName    string           `json:"retailer_name" example:"TechStore Inc"`
	RetailerAddress string           `json:"retailer_address,omitempty" example:"123 Main St"`
	InvoiceNumber   string           `json:"invoice_number,omitempty" example:"INV-2024-001"`
	SerialNumber    string           `json:"serial_number" example:"SN123456789"`
}

// CustomerClaimSummary represents a summary of customer claim
type CustomerClaimSummary struct {
	ID          uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174004"`
	ClaimNumber string     `json:"claim_number" example:"CLM-2024-001"`
	Status      string     `json:"status" example:"approved"`
	IssueType   string     `json:"issue_type" example:"hardware_failure"`
	SubmittedAt time.Time  `json:"submitted_at" example:"2024-02-15T10:30:00Z"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" example:"2024-02-20T15:45:00Z"`
	Resolution  string     `json:"resolution,omitempty" example:"Replaced defective component"`
}

// WarrantyDocument represents warranty-related documents
type WarrantyDocument struct {
	ID           uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174005"`
	DocumentType string    `json:"document_type" example:"proof_of_purchase"`
	DocumentName string    `json:"document_name" example:"receipt.pdf"`
	DocumentURL  string    `json:"document_url" example:"https://example.com/docs/receipt.pdf"`
	UploadedAt   time.Time `json:"uploaded_at" example:"2024-01-15T10:30:00Z"`
	FileSize     int64     `json:"file_size" example:"1024000"`
	MimeType     string    `json:"mime_type" example:"application/pdf"`
}

// CustomerSupportInfo represents customer support information
type CustomerSupportInfo struct {
	SupportEmail   string   `json:"support_email" example:"support@techbrand.com"`
	SupportPhone   string   `json:"support_phone" example:"+1-800-SUPPORT"`
	SupportHours   string   `json:"support_hours" example:"Mon-Fri 9AM-6PM EST"`
	OnlinePortal   string   `json:"online_portal" example:"https://support.techbrand.com"`
	ChatSupport    bool     `json:"chat_support" example:"true"`
	ServiceCenters []string `json:"service_centers"`
}

// CustomerWarrantyUpdateRequest represents a request to update warranty information
type CustomerWarrantyUpdateRequest struct {
	CustomerInfo    *CustomerRegistrationInfo `json:"customer_info,omitempty"`
	Preferences     *CustomerPreferences      `json:"preferences,omitempty"`
	NotificationSettings *NotificationSettings `json:"notification_settings,omitempty"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	ExpiryReminders      bool `json:"expiry_reminders" example:"true"`
	ClaimUpdates         bool `json:"claim_updates" example:"true"`
	MaintenanceReminders bool `json:"maintenance_reminders" example:"false"`
	ProductRecalls       bool `json:"product_recalls" example:"true"`
}

// CustomerWarrantyUpdateResponse represents the response after updating warranty
type CustomerWarrantyUpdateResponse struct {
	Success     bool      `json:"success" example:"true"`
	WarrantyID  uuid.UUID `json:"warranty_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Message     string    `json:"message" example:"Warranty information updated successfully"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	UpdatedFields []string `json:"updated_fields" example:"customer_info,preferences"`
}

// CustomerWarrantyErrorResponse represents error responses for customer warranty operations
type CustomerWarrantyErrorResponse struct {
	Error     string    `json:"error" example:"warranty_not_found"`
	Message   string    `json:"message" example:"The specified warranty could not be found"`
	Code      string    `json:"code" example:"CWR_404"`
	Details   string    `json:"details,omitempty" example:"Warranty ID: 123e4567-e89b-12d3-a456-426614174000"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	RequestID string    `json:"request_id,omitempty" example:"req_123456789"`
}