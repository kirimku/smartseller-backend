package entity

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BarcodeStatus represents the lifecycle status of a warranty barcode
type BarcodeStatus string

const (
	BarcodeStatusGenerated   BarcodeStatus = "generated"   // Created but not distributed
	BarcodeStatusDistributed BarcodeStatus = "distributed" // Sent to retailers/packaging
	BarcodeStatusActivated   BarcodeStatus = "activated"   // Customer registered warranty
	BarcodeStatusUsed        BarcodeStatus = "used"        // Warranty claim submitted
	BarcodeStatusExpired     BarcodeStatus = "expired"     // Warranty period ended
)

// Valid validates the barcode status
func (bs BarcodeStatus) Valid() bool {
	switch bs {
	case BarcodeStatusGenerated, BarcodeStatusDistributed, BarcodeStatusActivated, BarcodeStatusUsed, BarcodeStatusExpired:
		return true
	default:
		return false
	}
}

// String returns the string representation of BarcodeStatus
func (bs BarcodeStatus) String() string {
	return string(bs)
}

// Value implements the driver.Valuer interface for database storage
func (bs BarcodeStatus) Value() (driver.Value, error) {
	return string(bs), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (bs *BarcodeStatus) Scan(value interface{}) error {
	if value == nil {
		*bs = BarcodeStatusGenerated
		return nil
	}
	if str, ok := value.(string); ok {
		*bs = BarcodeStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into BarcodeStatus", value)
}

// DistributionInfo represents barcode distribution details
type DistributionInfo struct {
	DistributedAt  *time.Time `json:"distributed_at,omitempty"`
	DistributedTo  *string    `json:"distributed_to,omitempty"`
	BatchNumber    *string    `json:"batch_number,omitempty"`
	BatchID        *uuid.UUID `json:"batch_id,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	RecipientType  *string    `json:"recipient_type,omitempty"` // retailer, distributor, direct
	TrackingNumber *string    `json:"tracking_number,omitempty"`
	ShippingMethod *string    `json:"shipping_method,omitempty"`
}

// Value implements driver.Valuer interface for database storage
func (di DistributionInfo) Value() (driver.Value, error) {
	return json.Marshal(di)
}

// Scan implements sql.Scanner interface for database retrieval
func (di *DistributionInfo) Scan(value interface{}) error {
	if value == nil {
		*di = DistributionInfo{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into DistributionInfo", value)
	}

	return json.Unmarshal(b, di)
}

// ActivationInfo represents barcode activation details
type ActivationInfo struct {
	ActivatedAt      *time.Time             `json:"activated_at,omitempty"`
	CustomerID       *uuid.UUID             `json:"customer_id,omitempty"`
	PurchaseDate     *time.Time             `json:"purchase_date,omitempty"`
	PurchaseLocation *string                `json:"purchase_location,omitempty"`
	PurchaseInvoice  *string                `json:"purchase_invoice,omitempty"`
	PurchasePrice    *decimal.Decimal       `json:"purchase_price,omitempty"`
	RetailerInfo     map[string]interface{} `json:"retailer_info,omitempty"`
	ActivationMethod *string                `json:"activation_method,omitempty"` // qr_scan, manual_entry, auto
	ActivationSource *string                `json:"activation_source,omitempty"` // mobile_app, website, retail_system
}

// Value implements driver.Valuer interface for database storage
func (ai ActivationInfo) Value() (driver.Value, error) {
	return json.Marshal(ai)
}

// Scan implements sql.Scanner interface for database retrieval
func (ai *ActivationInfo) Scan(value interface{}) error {
	if value == nil {
		*ai = ActivationInfo{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into ActivationInfo", value)
	}

	return json.Unmarshal(b, ai)
}

// WarrantyBarcode represents a warranty barcode/QR code in the system
type WarrantyBarcode struct {
	// Primary identification
	ID            uuid.UUID `json:"id" db:"id"`
	BarcodeNumber string    `json:"barcode_number" db:"barcode_number"`
	QRCodeData    string    `json:"qr_code_data" db:"qr_code_data"`

	// Product and tenant associations
	ProductID    uuid.UUID `json:"product_id" db:"product_id"`
	StorefrontID uuid.UUID `json:"storefront_id" db:"storefront_id"`

	// Generation metadata for security tracking
	GeneratedAt       time.Time `json:"generated_at" db:"generated_at"`
	GenerationMethod  string    `json:"generation_method" db:"generation_method"`
	EntropyBits       int       `json:"entropy_bits" db:"entropy_bits"`
	GenerationAttempt int       `json:"generation_attempt" db:"generation_attempt"`
	CollisionChecked  bool      `json:"collision_checked" db:"collision_checked"`

	// Distribution tracking
	BatchID           *uuid.UUID `json:"batch_id,omitempty" db:"batch_id"`
	BatchNumber       *string    `json:"batch_number,omitempty" db:"batch_number"`
	DistributedAt     *time.Time `json:"distributed_at,omitempty" db:"distributed_at"`
	DistributedTo     *string    `json:"distributed_to,omitempty" db:"distributed_to"`
	DistributionNotes *string    `json:"distribution_notes,omitempty" db:"distribution_notes"`

	// Activation tracking
	ActivatedAt      *time.Time `json:"activated_at,omitempty" db:"activated_at"`
	CustomerID       *uuid.UUID `json:"customer_id,omitempty" db:"customer_id"`
	PurchaseDate     *time.Time `json:"purchase_date,omitempty" db:"purchase_date"`
	PurchaseLocation *string    `json:"purchase_location,omitempty" db:"purchase_location"`
	PurchaseInvoice  *string    `json:"purchase_invoice,omitempty" db:"purchase_invoice"`

	// Status management
	Status BarcodeStatus `json:"status" db:"status"`

	// Warranty period
	WarrantyPeriodMonths int        `json:"warranty_period_months" db:"warranty_period_months"`
	ExpiryDate           *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`

	// Audit fields
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields (not stored in database)
	IsExpired     bool   `json:"is_expired" db:"-"`
	IsActive      bool   `json:"is_active" db:"-"`
	DaysRemaining *int   `json:"days_remaining,omitempty" db:"-"`
	QRCodeURL     string `json:"qr_code_url" db:"-"`
}

// Character set for secure barcode generation (excludes confusing characters)
const BarcodeCharacterSet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const BarcodeRandomLength = 12

// NewWarrantyBarcode creates a new warranty barcode with default values
func NewWarrantyBarcode(productID, storefrontID, createdBy uuid.UUID, warrantyPeriodMonths int) *WarrantyBarcode {
	return &WarrantyBarcode{
		ID:                   uuid.New(),
		ProductID:            productID,
		StorefrontID:         storefrontID,
		GeneratedAt:          time.Now(),
		GenerationMethod:     "CSPRNG",
		EntropyBits:          60, // 12 chars * 5 bits per char
		GenerationAttempt:    1,
		CollisionChecked:     false, // Will be set to true after uniqueness check
		Status:               BarcodeStatusGenerated,
		WarrantyPeriodMonths: warrantyPeriodMonths,
		CreatedBy:            createdBy,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// GenerateBarcodeNumber generates a cryptographically secure barcode number
// Format: REX[YY][RANDOM_12] where YY is current year and RANDOM_12 is secure random
func (wb *WarrantyBarcode) GenerateBarcodeNumber() error {
	// Get current year (last 2 digits)
	currentYear := time.Now().Year() % 100
	yearPrefix := fmt.Sprintf("REX%02d", currentYear)

	// Generate cryptographically secure random string
	randomPart, err := generateSecureRandomString(BarcodeRandomLength)
	if err != nil {
		return fmt.Errorf("failed to generate secure random string: %w", err)
	}

	wb.BarcodeNumber = yearPrefix + randomPart

	// Generate QR code data URL for warranty claims
	wb.QRCodeData = fmt.Sprintf("https://warranty.smartseller.com/claim/%s", wb.BarcodeNumber)

	return nil
}

// generateSecureRandomString generates a cryptographically secure random string
func generateSecureRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = BarcodeCharacterSet[randomBytes[i]%byte(len(BarcodeCharacterSet))]
	}

	return string(result), nil
}

// ValidateBarcodeFormat validates the barcode number format
func (wb *WarrantyBarcode) ValidateBarcodeFormat() error {
	if wb.BarcodeNumber == "" {
		return fmt.Errorf("barcode number is required")
	}

	// Validate format: REX[YY][RANDOM_12]
	pattern := regexp.MustCompile(`^REX\d{2}[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{12}$`)
	if !pattern.MatchString(wb.BarcodeNumber) {
		return fmt.Errorf("invalid barcode format: expected REX[YY][RANDOM_12], got %s", wb.BarcodeNumber)
	}

	return nil
}

// Validate performs comprehensive validation of the warranty barcode
func (wb *WarrantyBarcode) Validate() error {
	// Required fields
	if wb.ProductID == uuid.Nil {
		return fmt.Errorf("product_id is required")
	}
	if wb.StorefrontID == uuid.Nil {
		return fmt.Errorf("storefront_id is required")
	}
	if wb.CreatedBy == uuid.Nil {
		return fmt.Errorf("created_by is required")
	}

	// Validate barcode format
	if err := wb.ValidateBarcodeFormat(); err != nil {
		return fmt.Errorf("barcode format validation failed: %w", err)
	}

	// Validate QR code data
	if wb.QRCodeData == "" {
		return fmt.Errorf("qr_code_data is required")
	}

	// Validate status
	if !wb.Status.Valid() {
		return fmt.Errorf("invalid barcode status: %s", wb.Status)
	}

	// Validate warranty period
	if wb.WarrantyPeriodMonths <= 0 {
		return fmt.Errorf("warranty period must be positive")
	}

	// Validate generation metadata
	if wb.GenerationAttempt <= 0 {
		return fmt.Errorf("generation attempt must be positive")
	}
	if wb.EntropyBits <= 0 {
		return fmt.Errorf("entropy bits must be positive")
	}

	return nil
}

// Activate activates the barcode with customer and purchase information
func (wb *WarrantyBarcode) Activate(customerID uuid.UUID, purchaseDate time.Time, purchaseLocation, purchaseInvoice string) error {
	if wb.Status != BarcodeStatusGenerated && wb.Status != BarcodeStatusDistributed {
		return fmt.Errorf("cannot activate barcode with status: %s", wb.Status)
	}

	now := time.Now()
	wb.ActivatedAt = &now
	wb.CustomerID = &customerID
	wb.PurchaseDate = &purchaseDate
	if purchaseLocation != "" {
		wb.PurchaseLocation = &purchaseLocation
	}
	if purchaseInvoice != "" {
		wb.PurchaseInvoice = &purchaseInvoice
	}
	wb.Status = BarcodeStatusActivated
	wb.UpdatedAt = now

	// Calculate expiry date
	wb.calculateExpiryDate()

	return nil
}

// MarkAsDistributed marks the barcode as distributed
func (wb *WarrantyBarcode) MarkAsDistributed(distributedTo string, batchID *uuid.UUID, notes string) error {
	if wb.Status != BarcodeStatusGenerated {
		return fmt.Errorf("can only distribute generated barcodes, current status: %s", wb.Status)
	}

	now := time.Now()
	wb.DistributedAt = &now
	wb.DistributedTo = &distributedTo
	wb.BatchID = batchID
	if notes != "" {
		wb.DistributionNotes = &notes
	}
	wb.Status = BarcodeStatusDistributed
	wb.UpdatedAt = now

	return nil
}

// MarkAsUsed marks the barcode as used for warranty claim
func (wb *WarrantyBarcode) MarkAsUsed() error {
	if wb.Status != BarcodeStatusActivated {
		return fmt.Errorf("can only use activated barcodes, current status: %s", wb.Status)
	}

	wb.Status = BarcodeStatusUsed
	wb.UpdatedAt = time.Now()

	return nil
}

// CheckExpiry checks if the warranty has expired and updates status if needed
func (wb *WarrantyBarcode) CheckExpiry() {
	if wb.ExpiryDate != nil && time.Now().After(*wb.ExpiryDate) {
		if wb.Status == BarcodeStatusActivated || wb.Status == BarcodeStatusGenerated || wb.Status == BarcodeStatusDistributed {
			wb.Status = BarcodeStatusExpired
			wb.UpdatedAt = time.Now()
		}
	}
}

// calculateExpiryDate calculates the warranty expiry date
func (wb *WarrantyBarcode) calculateExpiryDate() {
	if wb.PurchaseDate != nil && wb.WarrantyPeriodMonths > 0 {
		expiryDate := wb.PurchaseDate.AddDate(0, wb.WarrantyPeriodMonths, 0)
		wb.ExpiryDate = &expiryDate
	}
}

// ComputeFields calculates computed fields
func (wb *WarrantyBarcode) ComputeFields() {
	now := time.Now()

	// Check if expired
	wb.IsExpired = wb.ExpiryDate != nil && now.After(*wb.ExpiryDate)

	// Check if active (activated and not expired)
	wb.IsActive = wb.Status == BarcodeStatusActivated && !wb.IsExpired

	// Calculate days remaining
	if wb.ExpiryDate != nil && !wb.IsExpired {
		days := int(wb.ExpiryDate.Sub(now).Hours() / 24)
		wb.DaysRemaining = &days
	}

	// Generate QR code URL for frontend
	wb.QRCodeURL = fmt.Sprintf("/api/v1/warranty/qr/%s", wb.BarcodeNumber)
}

// IsValidForClaim checks if the barcode can be used for warranty claims
func (wb *WarrantyBarcode) IsValidForClaim() bool {
	wb.CheckExpiry()
	return wb.Status == BarcodeStatusActivated && !wb.IsExpired
}

// GetWarrantyInfo returns warranty information for display
type WarrantyInfo struct {
	BarcodeNumber        string     `json:"barcode_number"`
	Status               string     `json:"status"`
	PurchaseDate         *time.Time `json:"purchase_date,omitempty"`
	ExpiryDate           *time.Time `json:"expiry_date,omitempty"`
	WarrantyPeriodMonths int        `json:"warranty_period_months"`
	IsExpired            bool       `json:"is_expired"`
	IsActive             bool       `json:"is_active"`
	DaysRemaining        *int       `json:"days_remaining,omitempty"`
	CanClaim             bool       `json:"can_claim"`
}

// GetWarrantyInfo returns warranty information
func (wb *WarrantyBarcode) GetWarrantyInfo() *WarrantyInfo {
	wb.ComputeFields()

	return &WarrantyInfo{
		BarcodeNumber:        wb.BarcodeNumber,
		Status:               wb.Status.String(),
		PurchaseDate:         wb.PurchaseDate,
		ExpiryDate:           wb.ExpiryDate,
		WarrantyPeriodMonths: wb.WarrantyPeriodMonths,
		IsExpired:            wb.IsExpired,
		IsActive:             wb.IsActive,
		DaysRemaining:        wb.DaysRemaining,
		CanClaim:             wb.IsValidForClaim(),
	}
}

// CanTransitionTo checks if the barcode can transition to the specified status
func (wb *WarrantyBarcode) CanTransitionTo(newStatus BarcodeStatus) bool {
	switch wb.Status {
	case BarcodeStatusGenerated:
		return newStatus == BarcodeStatusDistributed || newStatus == BarcodeStatusActivated || newStatus == BarcodeStatusExpired
	case BarcodeStatusDistributed:
		return newStatus == BarcodeStatusActivated || newStatus == BarcodeStatusExpired
	case BarcodeStatusActivated:
		return newStatus == BarcodeStatusUsed || newStatus == BarcodeStatusExpired
	case BarcodeStatusUsed:
		return newStatus == BarcodeStatusExpired
	case BarcodeStatusExpired:
		return false // Terminal status
	default:
		return false
	}
}

// UpdateStatus updates the barcode status with validation
func (wb *WarrantyBarcode) UpdateStatus(newStatus BarcodeStatus) error {
	if !newStatus.Valid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	if !wb.CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", wb.Status, newStatus)
	}

	wb.Status = newStatus
	wb.UpdatedAt = time.Now()
	return nil
}

// String returns a string representation of the warranty barcode
func (wb *WarrantyBarcode) String() string {
	return fmt.Sprintf("WarrantyBarcode{ID: %s, Number: %s, Status: %s, Product: %s}",
		wb.ID.String(), wb.BarcodeNumber, wb.Status, wb.ProductID.String())
}
