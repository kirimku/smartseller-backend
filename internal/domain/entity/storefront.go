package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StorefrontStatus represents the status of a storefront
type StorefrontStatus string

const (
	StorefrontStatusActive    StorefrontStatus = "active"
	StorefrontStatusInactive  StorefrontStatus = "inactive"
	StorefrontStatusSuspended StorefrontStatus = "suspended"
)

// IsValid checks if the storefront status is valid
func (s StorefrontStatus) IsValid() bool {
	switch s {
	case StorefrontStatusActive, StorefrontStatusInactive, StorefrontStatusSuspended:
		return true
	default:
		return false
	}
}

// StorefrontSettings represents flexible configuration for a storefront
type StorefrontSettings struct {
	EnableGuestCheckout      bool     `json:"enable_guest_checkout"`
	RequirePhoneVerification bool     `json:"require_phone_verification"`
	RequireEmailVerification bool     `json:"require_email_verification"`
	AllowedPaymentMethods    []string `json:"allowed_payment_methods"`
	MinOrderAmount           *float64 `json:"min_order_amount"`
	MaxOrderAmount           *float64 `json:"max_order_amount"`
	Currency                 string   `json:"currency"`
	Timezone                 string   `json:"timezone"`
	Language                 string   `json:"language"`
	ShippingZones            []string `json:"shipping_zones"`
	TaxSettings              struct {
		EnableTax    bool    `json:"enable_tax"`
		TaxRate      float64 `json:"tax_rate"`
		TaxInclusive bool    `json:"tax_inclusive"`
	} `json:"tax_settings"`
	SocialMedia struct {
		Facebook  string `json:"facebook"`
		Instagram string `json:"instagram"`
		Twitter   string `json:"twitter"`
		WhatsApp  string `json:"whatsapp"`
	} `json:"social_media"`
}

// Value implements driver.Valuer interface for database storage
func (s StorefrontSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface for database retrieval
func (s *StorefrontSettings) Scan(value interface{}) error {
	if value == nil {
		*s = StorefrontSettings{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StorefrontSettings", value)
	}

	return json.Unmarshal(b, s)
}

// Storefront represents a seller's online storefront
type Storefront struct {
	ID              uuid.UUID          `json:"id" db:"id"`
	SellerID        uuid.UUID          `json:"seller_id" db:"seller_id"`
	Name            string             `json:"name" db:"name"`
	Slug            string             `json:"slug" db:"slug"`
	Description     *string            `json:"description,omitempty" db:"description"`
	Domain          *string            `json:"domain,omitempty" db:"domain"`
	Subdomain       *string            `json:"subdomain,omitempty" db:"subdomain"`
	Status          StorefrontStatus   `json:"status" db:"status"`
	Settings        StorefrontSettings `json:"settings" db:"settings"`
	LogoURL         *string            `json:"logo_url,omitempty" db:"logo_url"`
	FaviconURL      *string            `json:"favicon_url,omitempty" db:"favicon_url"`
	PrimaryColor    *string            `json:"primary_color,omitempty" db:"primary_color"`
	SecondaryColor  *string            `json:"secondary_color,omitempty" db:"secondary_color"`
	BusinessName    *string            `json:"business_name,omitempty" db:"business_name"`
	BusinessEmail   *string            `json:"business_email,omitempty" db:"business_email"`
	BusinessPhone   *string            `json:"business_phone,omitempty" db:"business_phone"`
	BusinessAddress *string            `json:"business_address,omitempty" db:"business_address"`
	TaxID           *string            `json:"tax_id,omitempty" db:"tax_id"`
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time         `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Validate performs comprehensive validation of storefront data
func (s *Storefront) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("storefront name is required")
	}
	if len(s.Name) > 255 {
		return fmt.Errorf("storefront name cannot exceed 255 characters")
	}

	if s.Slug == "" {
		return fmt.Errorf("storefront slug is required")
	}
	if !isValidSlug(s.Slug) {
		return fmt.Errorf("storefront slug must be lowercase alphanumeric with hyphens, 3-50 characters")
	}

	if s.SellerID == uuid.Nil {
		return fmt.Errorf("seller_id is required")
	}

	if !s.Status.IsValid() {
		return fmt.Errorf("invalid storefront status")
	}

	// Validate domain if provided
	if s.Domain != nil && *s.Domain != "" {
		if !isValidDomain(*s.Domain) {
			return fmt.Errorf("invalid domain format")
		}
	}

	// Validate subdomain if provided
	if s.Subdomain != nil && *s.Subdomain != "" {
		if !isValidSubdomain(*s.Subdomain) {
			return fmt.Errorf("invalid subdomain format")
		}
	}

	// Validate email if provided
	if s.BusinessEmail != nil && *s.BusinessEmail != "" {
		if !isValidEmail(*s.BusinessEmail) {
			return fmt.Errorf("invalid business email format")
		}
	}

	// Validate color codes if provided
	if s.PrimaryColor != nil && *s.PrimaryColor != "" {
		if !isValidHexColor(*s.PrimaryColor) {
			return fmt.Errorf("invalid primary color format (must be hex)")
		}
	}
	if s.SecondaryColor != nil && *s.SecondaryColor != "" {
		if !isValidHexColor(*s.SecondaryColor) {
			return fmt.Errorf("invalid secondary color format (must be hex)")
		}
	}

	return nil
}

// IsActive checks if the storefront is currently active
func (s *Storefront) IsActive() bool {
	return s.Status == StorefrontStatusActive && s.DeletedAt == nil
}

// GetURL returns the primary URL for the storefront
func (s *Storefront) GetURL() string {
	if s.Domain != nil && *s.Domain != "" {
		return "https://" + *s.Domain
	}
	if s.Subdomain != nil && *s.Subdomain != "" {
		return "https://" + *s.Subdomain + ".smartseller.com" // Replace with your actual domain
	}
	return "https://smartseller.com/store/" + s.Slug
}

// GetDisplayName returns the business name if available, otherwise the storefront name
func (s *Storefront) GetDisplayName() string {
	if s.BusinessName != nil && *s.BusinessName != "" {
		return *s.BusinessName
	}
	return s.Name
}

// GetContactEmail returns the business email if available, falls back to a generic email
func (s *Storefront) GetContactEmail() string {
	if s.BusinessEmail != nil && *s.BusinessEmail != "" {
		return *s.BusinessEmail
	}
	return "support@" + s.Slug + ".smartseller.com"
}

// NormalizeSlug converts and validates the slug to ensure it meets requirements
func (s *Storefront) NormalizeSlug() {
	s.Slug = strings.ToLower(strings.TrimSpace(s.Slug))
	// Replace spaces and underscores with hyphens
	s.Slug = regexp.MustCompile(`[\s_]+`).ReplaceAllString(s.Slug, "-")
	// Remove multiple consecutive hyphens
	s.Slug = regexp.MustCompile(`-+`).ReplaceAllString(s.Slug, "-")
	// Remove leading/trailing hyphens
	s.Slug = strings.Trim(s.Slug, "-")
}

// SetDefaultSettings initializes the storefront with sensible default settings
func (s *Storefront) SetDefaultSettings() {
	s.Settings = StorefrontSettings{
		EnableGuestCheckout:      true,
		RequirePhoneVerification: false,
		RequireEmailVerification: true,
		AllowedPaymentMethods:    []string{"credit_card", "bank_transfer"},
		Currency:                 "IDR",
		Timezone:                 "Asia/Jakarta",
		Language:                 "id",
		TaxSettings: struct {
			EnableTax    bool    `json:"enable_tax"`
			TaxRate      float64 `json:"tax_rate"`
			TaxInclusive bool    `json:"tax_inclusive"`
		}{
			EnableTax:    false,
			TaxRate:      0.0,
			TaxInclusive: true,
		},
	}
}

// Helper validation functions
func isValidSlug(slug string) bool {
	if len(slug) < 1 || len(slug) > 50 {
		return false
	}
	// Allow single characters or proper slug format
	if len(slug) == 1 {
		return regexp.MustCompile(`^[a-z0-9]$`).MatchString(slug)
	}
	return regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,48}[a-z0-9]$`).MatchString(slug)
}

func isValidDomain(domain string) bool {
	// Basic domain validation - you might want to use a more robust library
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9](\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9])*$`)
	return domainRegex.MatchString(domain) && len(domain) <= 253
}

func isValidSubdomain(subdomain string) bool {
	// Subdomain should be similar to slug but can be longer
	if len(subdomain) < 1 || len(subdomain) > 63 {
		return false
	}
	return regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`).MatchString(subdomain) ||
		regexp.MustCompile(`^[a-z0-9]$`).MatchString(subdomain)
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidHexColor(color string) bool {
	hexColorRegex := regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	return hexColorRegex.MatchString(color)
}
