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

// CustomerStatus represents the status of a customer
type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusBlocked  CustomerStatus = "blocked"
)

// IsValid checks if the customer status is valid
func (s CustomerStatus) IsValid() bool {
	switch s {
	case CustomerStatusActive, CustomerStatusInactive, CustomerStatusBlocked:
		return true
	default:
		return false
	}
}

// Gender represents customer gender options
type Gender string

const (
	GenderMale           Gender = "male"
	GenderFemale         Gender = "female"
	GenderOther          Gender = "other"
	GenderPreferNotToSay Gender = "prefer_not_to_say"
)

// IsValid checks if the gender is valid
func (g Gender) IsValid() bool {
	switch g {
	case GenderMale, GenderFemale, GenderOther, GenderPreferNotToSay:
		return true
	default:
		return false
	}
}

// CustomerType represents different customer types for segmentation
type CustomerType string

const (
	CustomerTypeRegular   CustomerType = "regular"
	CustomerTypeVIP       CustomerType = "vip"
	CustomerTypeWholesale CustomerType = "wholesale"
)

// IsValid checks if the customer type is valid
func (ct CustomerType) IsValid() bool {
	switch ct {
	case CustomerTypeRegular, CustomerTypeVIP, CustomerTypeWholesale:
		return true
	default:
		return false
	}
}

// CustomerPreferences represents flexible customer preferences stored as JSONB
type CustomerPreferences struct {
	Language                string                 `json:"language"`
	Currency                string                 `json:"currency"`
	Timezone                string                 `json:"timezone"`
	NotificationEmail       bool                   `json:"notification_email"`
	NotificationSMS         bool                   `json:"notification_sms"`
	NotificationPush        bool                   `json:"notification_push"`
	MarketingEmails         bool                   `json:"marketing_emails"`
	NewsletterSubscribed    bool                   `json:"newsletter_subscribed"`
	PreferredPaymentMethod  string                 `json:"preferred_payment_method"`
	PreferredShippingMethod string                 `json:"preferred_shipping_method"`
	CustomFields            map[string]interface{} `json:"custom_fields"`
}

// Value implements driver.Valuer interface for database storage
func (p CustomerPreferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner interface for database retrieval
func (p *CustomerPreferences) Scan(value interface{}) error {
	if value == nil {
		*p = CustomerPreferences{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into CustomerPreferences", value)
	}

	return json.Unmarshal(b, p)
}

// StringSlice represents a slice of strings that can be stored in the database
type StringSlice []string

// Value implements driver.Valuer interface for database storage
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface for database retrieval
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}

	return json.Unmarshal(b, s)
}

// Customer represents a customer in the multi-tenant system
type Customer struct {
	// Primary identification
	ID           uuid.UUID `json:"id" db:"id"`
	StorefrontID uuid.UUID `json:"storefront_id" db:"storefront_id"`

	// Contact information
	Email *string `json:"email,omitempty" db:"email"`
	Phone *string `json:"phone,omitempty" db:"phone"`

	// Personal information
	FirstName   *string    `json:"first_name,omitempty" db:"first_name"`
	LastName    *string    `json:"last_name,omitempty" db:"last_name"`
	FullName    *string    `json:"full_name,omitempty" db:"full_name"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	Gender      *Gender    `json:"gender,omitempty" db:"gender"`

	// Authentication
	PasswordHash           *string    `json:"-" db:"password_hash"` // Never expose password hash in JSON
	EmailVerifiedAt        *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	EmailVerificationToken *string    `json:"-" db:"email_verification_token"` // Sensitive data
	PhoneVerifiedAt        *time.Time `json:"phone_verified_at,omitempty" db:"phone_verified_at"`
	PhoneVerificationToken *string    `json:"-" db:"phone_verification_token"` // Sensitive data
	PasswordResetToken     *string    `json:"-" db:"password_reset_token"`     // Sensitive data
	PasswordResetExpiresAt *time.Time `json:"-" db:"password_reset_expires_at"`
	RefreshToken           *string    `json:"-" db:"refresh_token"` // Sensitive data
	RefreshTokenExpiresAt  *time.Time `json:"-" db:"refresh_token_expires_at"`
	LastLoginAt            *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	FailedLoginAttempts    int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil            *time.Time `json:"locked_until,omitempty" db:"locked_until"`

	// Customer status and segmentation
	Status       CustomerStatus      `json:"status" db:"status"`
	CustomerType CustomerType        `json:"customer_type" db:"customer_type"`
	Tags         StringSlice         `json:"tags,omitempty" db:"tags"`
	Preferences  CustomerPreferences `json:"preferences" db:"preferences"`

	// Marketing preferences
	AcceptsMarketing   bool       `json:"accepts_marketing" db:"accepts_marketing"`
	MarketingOptInDate *time.Time `json:"marketing_opt_in_date,omitempty" db:"marketing_opt_in_date"`

	// Customer metrics
	TotalOrders       int        `json:"total_orders" db:"total_orders"`
	TotalSpent        float64    `json:"total_spent" db:"total_spent"`
	AverageOrderValue float64    `json:"average_order_value" db:"average_order_value"`
	LastOrderDate     *time.Time `json:"last_order_date,omitempty" db:"last_order_date"`

	// Notes
	Notes         *string `json:"notes,omitempty" db:"notes"`
	InternalNotes *string `json:"-" db:"internal_notes"` // Only visible to sellers

	// Legacy field for backward compatibility with existing system
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Validate performs comprehensive validation of customer data
func (c *Customer) Validate() error {
	if c.StorefrontID == uuid.Nil {
		return fmt.Errorf("storefront_id is required")
	}

	// At least email or phone must be provided
	if (c.Email == nil || *c.Email == "") && (c.Phone == nil || *c.Phone == "") {
		return fmt.Errorf("either email or phone is required")
	}

	// Validate email format if provided
	if c.Email != nil && *c.Email != "" {
		if !isValidEmailForCustomer(*c.Email) {
			return fmt.Errorf("invalid email format")
		}
	}

	// Validate phone format if provided
	if c.Phone != nil && *c.Phone != "" {
		if !isValidPhone(*c.Phone) {
			return fmt.Errorf("invalid phone format")
		}
	}

	// Validate status
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid customer status")
	}

	// Validate customer type
	if !c.CustomerType.IsValid() {
		return fmt.Errorf("invalid customer type")
	}

	// Validate gender if provided
	if c.Gender != nil && !c.Gender.IsValid() {
		return fmt.Errorf("invalid gender")
	}

	// Validate names length
	if c.FirstName != nil && len(*c.FirstName) > 255 {
		return fmt.Errorf("first name cannot exceed 255 characters")
	}
	if c.LastName != nil && len(*c.LastName) > 255 {
		return fmt.Errorf("last name cannot exceed 255 characters")
	}

	// Validate failed login attempts
	if c.FailedLoginAttempts < 0 {
		return fmt.Errorf("failed login attempts cannot be negative")
	}

	// Validate monetary values
	if c.TotalSpent < 0 {
		return fmt.Errorf("total spent cannot be negative")
	}
	if c.AverageOrderValue < 0 {
		return fmt.Errorf("average order value cannot be negative")
	}
	if c.TotalOrders < 0 {
		return fmt.Errorf("total orders cannot be negative")
	}

	return nil
}

// IsEmailVerified checks if the customer's email is verified
func (c *Customer) IsEmailVerified() bool {
	return c.EmailVerifiedAt != nil
}

// IsPhoneVerified checks if the customer's phone is verified
func (c *Customer) IsPhoneVerified() bool {
	return c.PhoneVerifiedAt != nil
}

// IsLocked checks if the customer account is currently locked
func (c *Customer) IsLocked() bool {
	return c.LockedUntil != nil && c.LockedUntil.After(time.Now())
}

// IsActive checks if the customer is active and not deleted
func (c *Customer) IsActive() bool {
	return c.Status == CustomerStatusActive && c.DeletedAt == nil
}

// GetFullName returns the full name, either from the full_name field or composed from first/last name
func (c *Customer) GetFullName() string {
	if c.FullName != nil && *c.FullName != "" {
		return *c.FullName
	}

	var parts []string
	if c.FirstName != nil && *c.FirstName != "" {
		parts = append(parts, *c.FirstName)
	}
	if c.LastName != nil && *c.LastName != "" {
		parts = append(parts, *c.LastName)
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

// GetDisplayName returns the best available name for display purposes
func (c *Customer) GetDisplayName() string {
	fullName := c.GetFullName()
	if fullName != "" {
		return fullName
	}

	if c.Email != nil && *c.Email != "" {
		return *c.Email
	}

	if c.Phone != nil && *c.Phone != "" {
		return *c.Phone
	}

	return "Customer #" + c.ID.String()[:8]
}

// GetContactMethod returns the primary contact method
func (c *Customer) GetContactMethod() string {
	if c.Email != nil && *c.Email != "" {
		return *c.Email
	}

	if c.Phone != nil && *c.Phone != "" {
		return *c.Phone
	}

	return ""
}

// NormalizeEmail converts email to lowercase and trims whitespace
func (c *Customer) NormalizeEmail() {
	if c.Email != nil {
		normalized := strings.ToLower(strings.TrimSpace(*c.Email))
		c.Email = &normalized
	}
}

// NormalizePhone removes non-digit characters and formats the phone number
func (c *Customer) NormalizePhone() {
	if c.Phone != nil && *c.Phone != "" {
		// Basic phone normalization - remove spaces, dashes, parentheses
		phone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(*c.Phone, "")
		if !strings.HasPrefix(phone, "+") && len(phone) > 10 {
			// Assume Indonesian number if no country code
			if strings.HasPrefix(phone, "0") {
				phone = "+62" + phone[1:]
			} else if strings.HasPrefix(phone, "62") {
				phone = "+" + phone
			}
		}
		c.Phone = &phone
	}
}

// UpdateMetrics updates customer metrics based on order data
func (c *Customer) UpdateMetrics(orderCount int, totalSpent float64, lastOrderDate *time.Time) {
	c.TotalOrders = orderCount
	c.TotalSpent = totalSpent

	if orderCount > 0 {
		c.AverageOrderValue = totalSpent / float64(orderCount)
	} else {
		c.AverageOrderValue = 0
	}

	c.LastOrderDate = lastOrderDate
}

// IncrementFailedLogin increments failed login attempts
func (c *Customer) IncrementFailedLogin() {
	c.FailedLoginAttempts++
}

// ResetFailedLogin resets failed login attempts to zero
func (c *Customer) ResetFailedLogin() {
	c.FailedLoginAttempts = 0
	c.LockedUntil = nil
}

// LockAccount locks the account until the specified time
func (c *Customer) LockAccount(until time.Time) {
	c.LockedUntil = &until
}

// SetEmailVerified marks the email as verified
func (c *Customer) SetEmailVerified() {
	now := time.Now()
	c.EmailVerifiedAt = &now
	c.EmailVerificationToken = nil
}

// SetPhoneVerified marks the phone as verified
func (c *Customer) SetPhoneVerified() {
	now := time.Now()
	c.PhoneVerifiedAt = &now
	c.PhoneVerificationToken = nil
}

// UpdateLastLogin updates the last login timestamp
func (c *Customer) UpdateLastLogin() {
	now := time.Now()
	c.LastLoginAt = &now
}

// SetDefaultPreferences sets default preferences for the customer
func (c *Customer) SetDefaultPreferences() {
	c.Preferences = CustomerPreferences{
		Language:             "id",
		Currency:             "IDR",
		Timezone:             "Asia/Jakarta",
		NotificationEmail:    true,
		NotificationSMS:      true,
		NotificationPush:     true,
		MarketingEmails:      false,
		NewsletterSubscribed: false,
		CustomFields:         make(map[string]interface{}),
	}
}

// Helper validation functions for customer-specific needs
func isValidPhone(phone string) bool {
	// International phone number regex - allows + prefix and 7-15 digits
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)
	return phoneRegex.MatchString(phone)
}

// Inline email validation for customer entity
func isValidEmailForCustomer(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
