# SmartSeller Customer Management Implementation Plan

## 📋 **Document Overview**

**Document**: Customer Management Implementation Plan  
**Focus**: Multi-Tenant Customer Registration & Authentication System  
**Version**: 1.0  
**Created**: September 26, 2025  
**Status**: Implementation Ready  
**Owner**: SmartSeller Development Team  

---

## 🎯 **Implementation Overview**

### **Objective**
Implement a complete multi-tenant customer management system that allows:
- **Customer Registration**: End customers can register on specific storefronts
- **Customer Authentication**: Secure login with JWT tokens scoped to storefronts
- **Profile Management**: Customer profile and address management
- **Seller Dashboard**: Sellers can manage their customers

### **Architecture Strategy**
- **Multi-Tenancy**: Shared Database + Row-Level Isolation (current)
- **Future-Proof**: Abstract repository pattern for future database-per-tenant migration
- **Security**: JWT tokens scoped to storefronts with complete data isolation
- **Scalability**: Foundation ready for hybrid tenant strategies

### **Total Timeline**: 4 weeks (28 days)
### **Team**: 3 backend developers

---

## 📊 **Implementation Phases**

### **Phase 1**: Database Foundation & Entities (Week 1)
### **Phase 2**: Repository Layer & Tenant Resolution (Week 2) 
### **Phase 3**: Use Cases & Business Logic (Week 3)
### **Phase 4**: API Layer & Integration (Week 4)

---

## 🗄️ **Phase 1: Database Foundation & Entities (7 days)**

### **Phase Goals**
- Create all necessary database tables with proper multi-tenant design
- Implement core entity models with validation
- Set up migration infrastructure
- Establish data isolation patterns

### **Day 1-2: Database Schema Design & Migration**

#### **Task 1.1: Create Migration Files**
```bash
Files to create:
- migrations/20251001_001_create_storefronts_table.sql
- migrations/20251001_002_create_storefront_configs_table.sql
- migrations/20251002_003_create_customers_table.sql
- migrations/20251002_004_create_customer_addresses_table.sql
- migrations/20251003_005_create_customer_sessions_table.sql
```

**Storefront Tables:**
```sql
-- migrations/20251001_001_create_storefronts_table.sql
CREATE TABLE storefronts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    domain VARCHAR(255),
    subdomain VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance', 'suspended')),
    currency VARCHAR(3) DEFAULT 'IDR',
    language VARCHAR(5) DEFAULT 'id',
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_storefronts_seller_id ON storefronts(seller_id);
CREATE INDEX idx_storefronts_slug ON storefronts(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_storefronts_status ON storefronts(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_storefronts_domain ON storefronts(domain) WHERE domain IS NOT NULL;

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_storefronts_updated_at BEFORE UPDATE ON storefronts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

```sql
-- migrations/20251001_002_create_storefront_configs_table.sql  
CREATE TABLE storefront_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id) ON DELETE CASCADE,
    logo_url VARCHAR(500),
    favicon_url VARCHAR(500),
    primary_color VARCHAR(7) DEFAULT '#007bff',
    secondary_color VARCHAR(7) DEFAULT '#6c757d',
    accent_color VARCHAR(7),
    font_family VARCHAR(100) DEFAULT 'Arial, sans-serif',
    meta_title VARCHAR(255),
    meta_description TEXT,
    social_media_links JSONB DEFAULT '{}',
    contact_info JSONB DEFAULT '{}',
    business_hours JSONB DEFAULT '{}',
    custom_css TEXT,
    custom_js TEXT,
    google_analytics_id VARCHAR(50),
    facebook_pixel_id VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(storefront_id)
);

CREATE INDEX idx_storefront_configs_storefront_id ON storefront_configs(storefront_id);
CREATE TRIGGER update_storefront_configs_updated_at BEFORE UPDATE ON storefront_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Customer Tables:**
```sql
-- migrations/20251002_003_create_customers_table.sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id), -- TENANT KEY
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    password_salt VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other')),
    profile_picture VARCHAR(500),
    
    -- Verification status
    email_verified_at TIMESTAMPTZ,
    phone_verified_at TIMESTAMPTZ,
    email_verification_token VARCHAR(255),
    phone_verification_token VARCHAR(255),
    
    -- Account status
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'blocked')),
    last_login_at TIMESTAMPTZ,
    
    -- Customer analytics
    total_orders INTEGER DEFAULT 0,
    total_spent DECIMAL(15,2) DEFAULT 0,
    average_order_value DECIMAL(15,2) DEFAULT 0,
    lifetime_value DECIMAL(15,2) DEFAULT 0,
    
    -- Preferences
    preferences JSONB DEFAULT '{}',
    
    -- Tags and notes (for sellers)
    tags TEXT[] DEFAULT '{}',
    notes TEXT,
    
    -- Security
    refresh_token VARCHAR(500),
    refresh_token_expires_at TIMESTAMPTZ,
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMPTZ,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Multi-tenant constraint: email unique per storefront
    UNIQUE(storefront_id, email)
);

-- Indexes for multi-tenant queries
CREATE INDEX idx_customers_storefront_email ON customers(storefront_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_storefront_phone ON customers(storefront_id, phone) WHERE deleted_at IS NULL AND phone IS NOT NULL;
CREATE INDEX idx_customers_storefront_status ON customers(storefront_id, status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_email_verification ON customers(email_verification_token) WHERE email_verification_token IS NOT NULL;
CREATE INDEX idx_customers_password_reset ON customers(password_reset_token) WHERE password_reset_token IS NOT NULL;

-- Full-text search index
CREATE INDEX idx_customers_search ON customers USING gin(to_tsvector('english', first_name || ' ' || last_name || ' ' || email)) WHERE deleted_at IS NULL;

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

```sql
-- migrations/20251002_004_create_customer_addresses_table.sql
CREATE TABLE customer_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    type VARCHAR(20) DEFAULT 'both' CHECK (type IN ('billing', 'shipping', 'both')),
    label VARCHAR(50) NOT NULL,
    
    -- Address details
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    company VARCHAR(255),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    province VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) DEFAULT 'Indonesia',
    phone VARCHAR(20),
    
    -- Settings
    is_default BOOLEAN DEFAULT FALSE,
    
    -- Geolocation (for delivery optimization)
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_customer_addresses_customer_id ON customer_addresses(customer_id);
CREATE INDEX idx_customer_addresses_default ON customer_addresses(customer_id, is_default) WHERE is_default = true;
CREATE INDEX idx_customer_addresses_type ON customer_addresses(customer_id, type);

CREATE TRIGGER update_customer_addresses_updated_at BEFORE UPDATE ON customer_addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Ensure only one default address per customer
CREATE OR REPLACE FUNCTION ensure_single_default_address()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_default = true THEN
        UPDATE customer_addresses 
        SET is_default = false 
        WHERE customer_id = NEW.customer_id AND id != NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_default_address
    BEFORE INSERT OR UPDATE ON customer_addresses
    FOR EACH ROW EXECUTE FUNCTION ensure_single_default_address();
```

```sql
-- migrations/20251003_005_create_customer_sessions_table.sql
CREATE TABLE customer_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    session_token VARCHAR(500) NOT NULL UNIQUE,
    refresh_token VARCHAR(500) NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_customer_sessions_customer_id ON customer_sessions(customer_id);
CREATE INDEX idx_customer_sessions_storefront_id ON customer_sessions(storefront_id);
CREATE INDEX idx_customer_sessions_session_token ON customer_sessions(session_token) WHERE revoked_at IS NULL;
CREATE INDEX idx_customer_sessions_refresh_token ON customer_sessions(refresh_token) WHERE revoked_at IS NULL;
CREATE INDEX idx_customer_sessions_expires_at ON customer_sessions(expires_at);

CREATE TRIGGER update_customer_sessions_updated_at BEFORE UPDATE ON customer_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**Success Criteria Day 1-2:**
- ✅ All migration files created and tested
- ✅ Database schema properly indexes for multi-tenant queries
- ✅ Foreign key relationships established
- ✅ Triggers and constraints working

### **Day 3-4: Entity Models Implementation**

#### **Task 1.2: Core Entity Structs**

```go
// internal/domain/entity/storefront.go
package entity

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "time"
    "github.com/google/uuid"
)

type StorefrontStatus string

const (
    StorefrontStatusActive      StorefrontStatus = "active"
    StorefrontStatusInactive    StorefrontStatus = "inactive"
    StorefrontStatusMaintenance StorefrontStatus = "maintenance"
    StorefrontStatusSuspended   StorefrontStatus = "suspended"
)

func (s StorefrontStatus) IsValid() bool {
    switch s {
    case StorefrontStatusActive, StorefrontStatusInactive, StorefrontStatusMaintenance, StorefrontStatusSuspended:
        return true
    default:
        return false
    }
}

type StorefrontSettings struct {
    EnableGuestCheckout      bool    `json:"enable_guest_checkout"`
    RequireEmailVerification bool    `json:"require_email_verification"`
    AllowReviews            bool    `json:"allow_reviews"`
    EnableWishlist          bool    `json:"enable_wishlist"`
    MinOrderAmount          *float64 `json:"min_order_amount"`
    MaxOrderAmount          *float64 `json:"max_order_amount"`
}

func (s StorefrontSettings) Value() (driver.Value, error) {
    return json.Marshal(s)
}

func (s *StorefrontSettings) Scan(value interface{}) error {
    if value == nil {
        return nil
    }
    b, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("cannot scan %T into StorefrontSettings", value)
    }
    return json.Unmarshal(b, s)
}

type Storefront struct {
    ID          uuid.UUID          `json:"id" db:"id"`
    SellerID    uuid.UUID          `json:"seller_id" db:"seller_id"`
    Name        string             `json:"name" db:"name"`
    Slug        string             `json:"slug" db:"slug"`
    Description *string            `json:"description" db:"description"`
    Domain      *string            `json:"domain" db:"domain"`
    Subdomain   *string            `json:"subdomain" db:"subdomain"`
    Status      StorefrontStatus   `json:"status" db:"status"`
    Currency    string             `json:"currency" db:"currency"`
    Language    string             `json:"language" db:"language"`
    Timezone    string             `json:"timezone" db:"timezone"`
    Settings    StorefrontSettings `json:"settings" db:"settings"`
    CreatedAt   time.Time          `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
    DeletedAt   *time.Time         `json:"deleted_at,omitempty" db:"deleted_at"`
}

func (s *Storefront) Validate() error {
    if s.Name == "" {
        return fmt.Errorf("storefront name is required")
    }
    if s.Slug == "" {
        return fmt.Errorf("storefront slug is required")
    }
    if !s.Status.IsValid() {
        return fmt.Errorf("invalid storefront status: %s", s.Status)
    }
    if s.SellerID == uuid.Nil {
        return fmt.Errorf("seller ID is required")
    }
    return nil
}
```

```go
// internal/domain/entity/customer.go
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

type CustomerStatus string

const (
    CustomerStatusActive    CustomerStatus = "active"
    CustomerStatusInactive  CustomerStatus = "inactive"
    CustomerStatusSuspended CustomerStatus = "suspended"
    CustomerStatusBlocked   CustomerStatus = "blocked"
)

func (s CustomerStatus) IsValid() bool {
    switch s {
    case CustomerStatusActive, CustomerStatusInactive, CustomerStatusSuspended, CustomerStatusBlocked:
        return true
    default:
        return false
    }
}

type Gender string

const (
    GenderMale   Gender = "male"
    GenderFemale Gender = "female"
    GenderOther  Gender = "other"
)

func (g Gender) IsValid() bool {
    switch g {
    case GenderMale, GenderFemale, GenderOther:
        return true
    default:
        return false
    }
}

type CustomerPreferences struct {
    Language           string `json:"language"`
    Currency           string `json:"currency"`
    EmailNotifications bool   `json:"email_notifications"`
    SMSNotifications   bool   `json:"sms_notifications"`
    MarketingEmails    bool   `json:"marketing_emails"`
    OrderUpdates       bool   `json:"order_updates"`
    NewsletterSubscribed bool `json:"newsletter_subscribed"`
}

func (p CustomerPreferences) Value() (driver.Value, error) {
    return json.Marshal(p)
}

func (p *CustomerPreferences) Scan(value interface{}) error {
    if value == nil {
        *p = CustomerPreferences{
            Language:           "en",
            Currency:           "IDR",
            EmailNotifications: true,
            SMSNotifications:   false,
            MarketingEmails:    true,
            OrderUpdates:       true,
            NewsletterSubscribed: false,
        }
        return nil
    }
    b, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("cannot scan %T into CustomerPreferences", value)
    }
    return json.Unmarshal(b, p)
}

type Customer struct {
    // Primary identification
    ID           uuid.UUID `json:"id" db:"id"`
    StorefrontID uuid.UUID `json:"storefront_id" db:"storefront_id"` // TENANT KEY

    // Authentication
    Email        string  `json:"email" db:"email"`
    Phone        *string `json:"phone" db:"phone"`
    PasswordHash string  `json:"-" db:"password_hash"`
    PasswordSalt string  `json:"-" db:"password_salt"`

    // Personal Information
    FirstName   string     `json:"first_name" db:"first_name"`
    LastName    string     `json:"last_name" db:"last_name"`
    DateOfBirth *time.Time `json:"date_of_birth" db:"date_of_birth"`
    Gender      *Gender    `json:"gender" db:"gender"`
    ProfilePicture *string `json:"profile_picture" db:"profile_picture"`

    // Verification Status
    EmailVerifiedAt         *time.Time `json:"email_verified_at" db:"email_verified_at"`
    PhoneVerifiedAt         *time.Time `json:"phone_verified_at" db:"phone_verified_at"`
    EmailVerificationToken  *string    `json:"-" db:"email_verification_token"`
    PhoneVerificationToken  *string    `json:"-" db:"phone_verification_token"`

    // Account Status
    Status      CustomerStatus `json:"status" db:"status"`
    LastLoginAt *time.Time     `json:"last_login_at" db:"last_login_at"`

    // Business Intelligence
    TotalOrders       int     `json:"total_orders" db:"total_orders"`
    TotalSpent        float64 `json:"total_spent" db:"total_spent"`
    AverageOrderValue float64 `json:"average_order_value" db:"average_order_value"`
    LifetimeValue     float64 `json:"lifetime_value" db:"lifetime_value"`

    // Preferences & Metadata
    Preferences CustomerPreferences `json:"preferences" db:"preferences"`
    Tags        []string           `json:"tags" db:"tags"`
    Notes       *string            `json:"notes" db:"notes"`

    // Security & Sessions
    RefreshToken           *string    `json:"-" db:"refresh_token"`
    RefreshTokenExpiresAt  *time.Time `json:"-" db:"refresh_token_expires_at"`
    PasswordResetToken     *string    `json:"-" db:"password_reset_token"`
    PasswordResetExpiresAt *time.Time `json:"-" db:"password_reset_expires_at"`
    FailedLoginAttempts    int        `json:"-" db:"failed_login_attempts"`
    LockedUntil            *time.Time `json:"-" db:"locked_until"`

    // Audit
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

func (c *Customer) Validate() error {
    if c.StorefrontID == uuid.Nil {
        return fmt.Errorf("storefront ID is required")
    }
    if c.Email == "" {
        return fmt.Errorf("email is required")
    }
    if !isValidEmail(c.Email) {
        return fmt.Errorf("invalid email format")
    }
    if c.FirstName == "" {
        return fmt.Errorf("first name is required")
    }
    if c.LastName == "" {
        return fmt.Errorf("last name is required")
    }
    if c.PasswordHash == "" {
        return fmt.Errorf("password hash is required")
    }
    if c.PasswordSalt == "" {
        return fmt.Errorf("password salt is required")
    }
    if !c.Status.IsValid() {
        return fmt.Errorf("invalid customer status: %s", c.Status)
    }
    if c.Gender != nil && !c.Gender.IsValid() {
        return fmt.Errorf("invalid gender: %s", *c.Gender)
    }
    if c.Phone != nil && !isValidPhone(*c.Phone) {
        return fmt.Errorf("invalid phone format")
    }
    return nil
}

func (c *Customer) IsEmailVerified() bool {
    return c.EmailVerifiedAt != nil
}

func (c *Customer) IsPhoneVerified() bool {
    return c.PhoneVerifiedAt != nil
}

func (c *Customer) IsLocked() bool {
    return c.LockedUntil != nil && c.LockedUntil.After(time.Now())
}

func (c *Customer) GetFullName() string {
    return strings.TrimSpace(c.FirstName + " " + c.LastName)
}

func (c *Customer) NormalizeEmail() {
    c.Email = strings.ToLower(strings.TrimSpace(c.Email))
}

// Helper functions
func isValidEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

func isValidPhone(phone string) bool {
    phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
    return phoneRegex.MatchString(phone)
}
```

```go
// internal/domain/entity/customer_address.go
package entity

import (
    "fmt"
    "strings"
    "time"
    "github.com/google/uuid"
)

type AddressType string

const (
    AddressTypeBilling  AddressType = "billing"
    AddressTypeShipping AddressType = "shipping"
    AddressTypeBoth     AddressType = "both"
)

func (a AddressType) IsValid() bool {
    switch a {
    case AddressTypeBilling, AddressTypeShipping, AddressTypeBoth:
        return true
    default:
        return false
    }
}

type CustomerAddress struct {
    ID         uuid.UUID   `json:"id" db:"id"`
    CustomerID uuid.UUID   `json:"customer_id" db:"customer_id"`
    Type       AddressType `json:"type" db:"type"`
    Label      string      `json:"label" db:"label"`

    // Address Information
    FirstName    string  `json:"first_name" db:"first_name"`
    LastName     string  `json:"last_name" db:"last_name"`
    Company      *string `json:"company" db:"company"`
    AddressLine1 string  `json:"address_line1" db:"address_line1"`
    AddressLine2 *string `json:"address_line2" db:"address_line2"`
    City         string  `json:"city" db:"city"`
    Province     string  `json:"province" db:"province"`
    PostalCode   string  `json:"postal_code" db:"postal_code"`
    Country      string  `json:"country" db:"country"`
    Phone        *string `json:"phone" db:"phone"`

    // Settings
    IsDefault bool `json:"is_default" db:"is_default"`

    // Geolocation
    Latitude  *float64 `json:"latitude" db:"latitude"`
    Longitude *float64 `json:"longitude" db:"longitude"`

    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (a *CustomerAddress) Validate() error {
    if a.CustomerID == uuid.Nil {
        return fmt.Errorf("customer ID is required")
    }
    if a.Label == "" {
        return fmt.Errorf("address label is required")
    }
    if a.FirstName == "" {
        return fmt.Errorf("first name is required")
    }
    if a.LastName == "" {
        return fmt.Errorf("last name is required")
    }
    if a.AddressLine1 == "" {
        return fmt.Errorf("address line 1 is required")
    }
    if a.City == "" {
        return fmt.Errorf("city is required")
    }
    if a.Province == "" {
        return fmt.Errorf("province is required")
    }
    if a.PostalCode == "" {
        return fmt.Errorf("postal code is required")
    }
    if a.Country == "" {
        return fmt.Errorf("country is required")
    }
    if !a.Type.IsValid() {
        return fmt.Errorf("invalid address type: %s", a.Type)
    }
    if a.Phone != nil && !isValidPhone(*a.Phone) {
        return fmt.Errorf("invalid phone format")
    }
    return nil
}

func (a *CustomerAddress) GetFullAddress() string {
    parts := []string{}
    
    if a.Company != nil && *a.Company != "" {
        parts = append(parts, *a.Company)
    }
    
    parts = append(parts, a.AddressLine1)
    
    if a.AddressLine2 != nil && *a.AddressLine2 != "" {
        parts = append(parts, *a.AddressLine2)
    }
    
    parts = append(parts, fmt.Sprintf("%s, %s %s", a.City, a.Province, a.PostalCode))
    parts = append(parts, a.Country)
    
    return strings.Join(parts, ", ")
}

func (a *CustomerAddress) GetFullName() string {
    return strings.TrimSpace(a.FirstName + " " + a.LastName)
}
```

**Success Criteria Day 3-4:**
- ✅ All entity structs implemented with proper validation
- ✅ JSON marshaling/unmarshaling working for JSONB fields
- ✅ Business logic methods implemented
- ✅ Unit tests for entity validation

### **Day 5: Domain Interfaces & Error Types**

#### **Task 1.3: Repository Interfaces**

```go
// internal/domain/repository/storefront_repository.go
package repository

import (
    "context"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/google/uuid"
)

type StorefrontRepository interface {
    // Core CRUD operations
    Create(ctx context.Context, storefront *entity.Storefront) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Storefront, error)
    GetBySlug(ctx context.Context, slug string) (*entity.Storefront, error)
    GetBySellerID(ctx context.Context, sellerID uuid.UUID) ([]*entity.Storefront, error)
    Update(ctx context.Context, storefront *entity.Storefront) error
    SoftDelete(ctx context.Context, id uuid.UUID) error

    // Business queries
    GetActiveStorefronts(ctx context.Context) ([]*entity.Storefront, error)
    GetByDomain(ctx context.Context, domain string) (*entity.Storefront, error)
    ExistsBySlug(ctx context.Context, slug string) (bool, error)
    ExistsByDomain(ctx context.Context, domain string) (bool, error)
    
    // For multi-tenancy future-proofing
    GetStorefrontStats(ctx context.Context, storefrontID uuid.UUID) (*StorefrontStats, error)
}

type StorefrontStats struct {
    CustomerCount   int     `json:"customer_count"`
    OrderCount      int     `json:"order_count"`
    TotalRevenue    float64 `json:"total_revenue"`
    AvgQueryTime    int64   `json:"avg_query_time_ms"`
}
```

```go
// internal/domain/repository/customer_repository.go
package repository

import (
    "context"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/google/uuid"
)

type CustomerRepository interface {
    // Core CRUD - All include tenant isolation via storefront_id
    Create(ctx context.Context, customer *entity.Customer) error
    GetByID(ctx context.Context, storefrontID, customerID uuid.UUID) (*entity.Customer, error)
    GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error)
    GetByPhone(ctx context.Context, storefrontID uuid.UUID, phone string) (*entity.Customer, error)
    Update(ctx context.Context, customer *entity.Customer) error
    SoftDelete(ctx context.Context, storefrontID, customerID uuid.UUID) error

    // Business queries with tenant isolation
    GetByStorefront(ctx context.Context, req *GetCustomersRequest) (*CustomerListResponse, error)
    Search(ctx context.Context, storefrontID uuid.UUID, req *SearchCustomersRequest) (*CustomerSearchResult, error)
    GetTopCustomers(ctx context.Context, storefrontID uuid.UUID, limit int) ([]*entity.Customer, error)
    GetCustomerStats(ctx context.Context, storefrontID uuid.UUID) (*CustomerStats, error)

    // Authentication-specific operations
    GetByEmailVerificationToken(ctx context.Context, token string) (*entity.Customer, error)
    GetByPasswordResetToken(ctx context.Context, token string) (*entity.Customer, error)
    UpdateLastLogin(ctx context.Context, storefrontID, customerID uuid.UUID) error
    UpdateRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string, expiresAt *time.Time) error
    ClearRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID) error
    UpdateEmailVerification(ctx context.Context, storefrontID, customerID uuid.UUID, verified bool) error
    UpdateFailedLoginAttempts(ctx context.Context, storefrontID, customerID uuid.UUID, attempts int) error
    LockAccount(ctx context.Context, storefrontID, customerID uuid.UUID, until *time.Time) error
}

// Request/Response types
type GetCustomersRequest struct {
    StorefrontID uuid.UUID
    Page         int
    PageSize     int
    Search       string
    Status       *entity.CustomerStatus
    OrderBy      string
    SortDesc     bool
}

type CustomerListResponse struct {
    Customers  []*entity.Customer `json:"customers"`
    Total      int                `json:"total"`
    Page       int                `json:"page"`
    PageSize   int                `json:"page_size"`
    TotalPages int                `json:"total_pages"`
}

type SearchCustomersRequest struct {
    Query    string
    Page     int
    PageSize int
}

type CustomerSearchResult struct {
    Customers []*entity.Customer `json:"customers"`
    Total     int                `json:"total"`
    Query     string             `json:"query"`
}

type CustomerStats struct {
    TotalCustomers    int     `json:"total_customers"`
    ActiveCustomers   int     `json:"active_customers"`
    VerifiedCustomers int     `json:"verified_customers"`
    NewThisMonth      int     `json:"new_this_month"`
    TotalRevenue      float64 `json:"total_revenue"`
    AvgOrderValue     float64 `json:"avg_order_value"`
}
```

```go
// internal/domain/repository/customer_address_repository.go
package repository

import (
    "context"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/google/uuid"
)

type CustomerAddressRepository interface {
    Create(ctx context.Context, address *entity.CustomerAddress) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.CustomerAddress, error)
    GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error)
    GetDefaultByCustomerID(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error)
    Update(ctx context.Context, address *entity.CustomerAddress) error
    Delete(ctx context.Context, id uuid.UUID) error
    SetAsDefault(ctx context.Context, customerID, addressID uuid.UUID) error
}
```

#### **Task 1.4: Domain Error Types**

```go
// internal/domain/errors/customer_errors.go
package errors

import (
    "fmt"
)

// Customer-related errors
var (
    ErrCustomerNotFound         = NewDomainError("CUSTOMER_NOT_FOUND", "Customer not found")
    ErrEmailAlreadyExists       = NewDomainError("EMAIL_ALREADY_EXISTS", "Email already exists")
    ErrPhoneAlreadyExists       = NewDomainError("PHONE_ALREADY_EXISTS", "Phone number already exists")
    ErrInvalidCredentials       = NewDomainError("INVALID_CREDENTIALS", "Invalid email or password")
    ErrAccountSuspended         = NewDomainError("ACCOUNT_SUSPENDED", "Account is suspended")
    ErrAccountLocked            = NewDomainError("ACCOUNT_LOCKED", "Account is temporarily locked due to too many failed login attempts")
    ErrEmailNotVerified         = NewDomainError("EMAIL_NOT_VERIFIED", "Email address is not verified")
    ErrInvalidEmailToken        = NewDomainError("INVALID_EMAIL_TOKEN", "Invalid or expired email verification token")
    ErrInvalidPasswordToken     = NewDomainError("INVALID_PASSWORD_TOKEN", "Invalid or expired password reset token")
    ErrWeakPassword            = NewDomainError("WEAK_PASSWORD", "Password does not meet security requirements")
    ErrInvalidRefreshToken     = NewDomainError("INVALID_REFRESH_TOKEN", "Invalid or expired refresh token")
)

// Storefront-related errors
var (
    ErrStorefrontNotFound       = NewDomainError("STOREFRONT_NOT_FOUND", "Storefront not found")
    ErrStorefrontInactive       = NewDomainError("STOREFRONT_INACTIVE", "Storefront is not active")
    ErrSlugAlreadyExists        = NewDomainError("SLUG_ALREADY_EXISTS", "Storefront slug already exists")
    ErrDomainAlreadyExists      = NewDomainError("DOMAIN_ALREADY_EXISTS", "Domain already exists")
    ErrUnauthorizedStorefront   = NewDomainError("UNAUTHORIZED_STOREFRONT", "Not authorized to access this storefront")
)

// Address-related errors
var (
    ErrAddressNotFound         = NewDomainError("ADDRESS_NOT_FOUND", "Address not found")
    ErrInvalidAddressType      = NewDomainError("INVALID_ADDRESS_TYPE", "Invalid address type")
    ErrCannotDeleteDefaultAddr = NewDomainError("CANNOT_DELETE_DEFAULT_ADDRESS", "Cannot delete default address")
)

// Multi-tenancy errors
var (
    ErrTenantMismatch          = NewDomainError("TENANT_MISMATCH", "Resource does not belong to this tenant")
    ErrTenantNotFound          = NewDomainError("TENANT_NOT_FOUND", "Tenant not found")
    ErrTenantAccessDenied      = NewDomainError("TENANT_ACCESS_DENIED", "Access denied for this tenant")
)

// Domain error type
type DomainError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewDomainError(code, message string) *DomainError {
    return &DomainError{
        Code:    code,
        Message: message,
    }
}

// Check if error is of specific type
func IsCustomerNotFound(err error) bool {
    if domainErr, ok := err.(*DomainError); ok {
        return domainErr.Code == "CUSTOMER_NOT_FOUND"
    }
    return false
}

func IsStorefrontNotFound(err error) bool {
    if domainErr, ok := err.(*DomainError); ok {
        return domainErr.Code == "STOREFRONT_NOT_FOUND"
    }
    return false
}
```

**Success Criteria Day 5:**
- ✅ All repository interfaces defined with multi-tenant considerations
- ✅ Domain error types implemented
- ✅ Request/Response DTOs created
- ✅ Error handling patterns established

### **Day 6-7: Future-Proof Architecture Foundation**

#### **Task 1.5: Tenant Resolution System**

```go
// internal/infrastructure/tenant/tenant_resolver.go
package tenant

import (
    "context"
    "database/sql"
    "fmt"
    "sync"
    "time"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/google/uuid"
)

// TenantType defines different tenant isolation strategies
type TenantType string

const (
    TenantTypeShared   TenantType = "shared"   // Current: Row-level isolation
    TenantTypeSchema   TenantType = "schema"   // Future: Schema per tenant
    TenantTypeDatabase TenantType = "database" // Future: Database per tenant
)

// TenantContext holds tenant information for the current request
type TenantContext struct {
    StorefrontID   uuid.UUID
    StorefrontSlug string
    SellerID       uuid.UUID
    TenantType     TenantType
}

// TenantResolver handles tenant database resolution
type TenantResolver interface {
    GetTenantType(ctx context.Context, storefrontID uuid.UUID) (TenantType, error)
    GetDatabaseConnection(ctx context.Context, storefrontID uuid.UUID) (*sql.DB, error)
    GetStorefrontBySlug(ctx context.Context, slug string) (*entity.Storefront, error)
    CreateTenantContext(storefront *entity.Storefront) *TenantContext
}

type tenantResolver struct {
    sharedDB     *sql.DB
    tenantDBs    map[uuid.UUID]*sql.DB
    config       *TenantConfig
    cache        TenantCache
    mu           sync.RWMutex
}

type TenantConfig struct {
    DefaultTenantType        TenantType            `yaml:"default_tenant_type"`
    TenantOverrides         map[string]TenantType `yaml:"tenant_overrides"`
    SharedDatabaseURL       string                `yaml:"shared_database_url"`
    TenantDatabasePattern   string                `yaml:"tenant_database_pattern"`
    MaxConnectionsPerTenant int                   `yaml:"max_connections_per_tenant"`
    MigrationThresholds     MigrationThresholds   `yaml:"migration_thresholds"`
}

type MigrationThresholds struct {
    CustomerCount    int           `yaml:"customer_count"`
    OrderCount       int           `yaml:"order_count"`
    AvgQueryTime     time.Duration `yaml:"avg_query_time"`
}

func NewTenantResolver(sharedDB *sql.DB, config *TenantConfig, cache TenantCache) TenantResolver {
    return &tenantResolver{
        sharedDB:  sharedDB,
        tenantDBs: make(map[uuid.UUID]*sql.DB),
        config:    config,
        cache:     cache,
    }
}

func (tr *tenantResolver) GetTenantType(ctx context.Context, storefrontID uuid.UUID) (TenantType, error) {
    // Check explicit configuration overrides first
    if tenantType, exists := tr.config.TenantOverrides[storefrontID.String()]; exists {
        return tenantType, nil
    }
    
    // Check automatic migration thresholds (future implementation)
    stats, err := tr.getStorefrontStats(ctx, storefrontID)
    if err != nil {
        // If we can't get stats, default to shared
        return tr.config.DefaultTenantType, nil
    }
    
    // Apply automatic migration rules
    if tr.shouldMigrateToDatabase(stats) {
        return TenantTypeDatabase, nil
    }
    
    if tr.shouldMigrateToSchema(stats) {
        return TenantTypeSchema, nil
    }
    
    return tr.config.DefaultTenantType, nil
}

func (tr *tenantResolver) GetDatabaseConnection(ctx context.Context, storefrontID uuid.UUID) (*sql.DB, error) {
    tenantType, err := tr.GetTenantType(ctx, storefrontID)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve tenant type: %w", err)
    }
    
    switch tenantType {
    case TenantTypeShared, TenantTypeSchema:
        return tr.sharedDB, nil
    case TenantTypeDatabase:
        return tr.getTenantDatabase(storefrontID)
    default:
        return nil, fmt.Errorf("unsupported tenant type: %s", tenantType)
    }
}

func (tr *tenantResolver) GetStorefrontBySlug(ctx context.Context, slug string) (*entity.Storefront, error) {
    // Try cache first
    if storefront := tr.cache.GetStorefront(slug); storefront != nil {
        return storefront, nil
    }
    
    // Query database
    query := `
        SELECT id, seller_id, name, slug, description, domain, subdomain, 
               status, currency, language, timezone, settings, 
               created_at, updated_at, deleted_at
        FROM storefronts 
        WHERE slug = $1 AND deleted_at IS NULL
    `
    
    row := tr.sharedDB.QueryRowContext(ctx, query, slug)
    
    storefront := &entity.Storefront{}
    err := row.Scan(
        &storefront.ID, &storefront.SellerID, &storefront.Name,
        &storefront.Slug, &storefront.Description, &storefront.Domain,
        &storefront.Subdomain, &storefront.Status, &storefront.Currency,
        &storefront.Language, &storefront.Timezone, &storefront.Settings,
        &storefront.CreatedAt, &storefront.UpdatedAt, &storefront.DeletedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("storefront not found")
        }
        return nil, err
    }
    
    // Cache for 1 hour
    tr.cache.SetStorefront(slug, storefront, time.Hour)
    
    return storefront, nil
}

func (tr *tenantResolver) CreateTenantContext(storefront *entity.Storefront) *TenantContext {
    tenantType, _ := tr.GetTenantType(context.Background(), storefront.ID)
    
    return &TenantContext{
        StorefrontID:   storefront.ID,
        StorefrontSlug: storefront.Slug,
        SellerID:       storefront.SellerID,
        TenantType:     tenantType,
    }
}

// Helper methods
func (tr *tenantResolver) getTenantDatabase(storefrontID uuid.UUID) (*sql.DB, error) {
    tr.mu.RLock()
    if db, exists := tr.tenantDBs[storefrontID]; exists {
        tr.mu.RUnlock()
        return db, nil
    }
    tr.mu.RUnlock()
    
    // Create new connection
    tr.mu.Lock()
    defer tr.mu.Unlock()
    
    // Double-check pattern
    if db, exists := tr.tenantDBs[storefrontID]; exists {
        return db, nil
    }
    
    // Generate database URL for this tenant
    dbURL := fmt.Sprintf(tr.config.TenantDatabasePattern, storefrontID.String())
    
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to tenant database: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(tr.config.MaxConnectionsPerTenant)
    db.SetMaxIdleConns(tr.config.MaxConnectionsPerTenant / 4)
    db.SetConnMaxLifetime(time.Hour)
    
    tr.tenantDBs[storefrontID] = db
    return db, nil
}

func (tr *tenantResolver) shouldMigrateToDatabase(stats *StorefrontStats) bool {
    return stats.CustomerCount > tr.config.MigrationThresholds.CustomerCount ||
           stats.OrderCount > tr.config.MigrationThresholds.OrderCount ||
           time.Duration(stats.AvgQueryTime)*time.Millisecond > tr.config.MigrationThresholds.AvgQueryTime
}

func (tr *tenantResolver) shouldMigrateToSchema(stats *StorefrontStats) bool {
    // Schema migration for medium-size storefronts
    return stats.CustomerCount > tr.config.MigrationThresholds.CustomerCount/2
}

func (tr *tenantResolver) getStorefrontStats(ctx context.Context, storefrontID uuid.UUID) (*StorefrontStats, error) {
    // Implementation would query actual statistics
    // For now, return default stats
    return &StorefrontStats{
        CustomerCount: 0,
        OrderCount:    0,
        AvgQueryTime:  50, // 50ms
    }, nil
}
```

#### **Task 1.6: Caching System**

```go
// internal/infrastructure/tenant/tenant_cache.go
package tenant

import (
    "sync"
    "time"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
)

type TenantCache interface {
    GetStorefront(slug string) *entity.Storefront
    SetStorefront(slug string, storefront *entity.Storefront, ttl time.Duration)
    InvalidateStorefront(slug string)
    GetCustomer(storefrontID, customerID string) *entity.Customer
    SetCustomer(storefrontID, customerID string, customer *entity.Customer, ttl time.Duration)
    InvalidateCustomer(storefrontID, customerID string)
}

type cacheItem struct {
    data      interface{}
    expiresAt time.Time
}

type inMemoryCache struct {
    items map[string]cacheItem
    mu    sync.RWMutex
}

func NewInMemoryTenantCache() TenantCache {
    cache := &inMemoryCache{
        items: make(map[string]cacheItem),
    }
    
    // Start cleanup goroutine
    go cache.cleanup()
    
    return cache
}

func (c *inMemoryCache) GetStorefront(slug string) *entity.Storefront {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    key := "storefront:" + slug
    item, exists := c.items[key]
    if !exists || time.Now().After(item.expiresAt) {
        return nil
    }
    
    return item.data.(*entity.Storefront)
}

func (c *inMemoryCache) SetStorefront(slug string, storefront *entity.Storefront, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    key := "storefront:" + slug
    c.items[key] = cacheItem{
        data:      storefront,
        expiresAt: time.Now().Add(ttl),
    }
}

func (c *inMemoryCache) InvalidateStorefront(slug string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    key := "storefront:" + slug
    delete(c.items, key)
}

func (c *inMemoryCache) GetCustomer(storefrontID, customerID string) *entity.Customer {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    key := "customer:" + storefrontID + ":" + customerID
    item, exists := c.items[key]
    if !exists || time.Now().After(item.expiresAt) {
        return nil
    }
    
    return item.data.(*entity.Customer)
}

func (c *inMemoryCache) SetCustomer(storefrontID, customerID string, customer *entity.Customer, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    key := "customer:" + storefrontID + ":" + customerID
    c.items[key] = cacheItem{
        data:      customer,
        expiresAt: time.Now().Add(ttl),
    }
}

func (c *inMemoryCache) InvalidateCustomer(storefrontID, customerID string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    key := "customer:" + storefrontID + ":" + customerID
    delete(c.items, key)
}

func (c *inMemoryCache) cleanup() {
    ticker := time.NewTicker(time.Minute * 5)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, item := range c.items {
            if now.After(item.expiresAt) {
                delete(c.items, key)
            }
        }
        c.mu.Unlock()
    }
}
```

**Success Criteria Day 6-7:**
- ✅ Tenant resolver system implemented
- ✅ Caching system for storefronts and customers
- ✅ Configuration system for tenant strategies
- ✅ Foundation ready for future migration to database-per-tenant

---

## 🎯 **Phase 1 Success Criteria**

At the end of Phase 1, you should have:

### **✅ Database Foundation**
- All multi-tenant tables created with proper indexes
- Foreign key relationships and constraints working
- Migration system ready for production deployment

### **✅ Entity Models**
- Complete entity structs with validation
- JSON marshaling for JSONB fields
- Business logic methods implemented

### **✅ Domain Interfaces** 
- Repository interfaces with multi-tenant considerations
- Domain error types for proper error handling
- Future-proof architecture foundation

### **✅ Tenant Resolution System**
- Abstract tenant resolver for future migration
- Caching system for performance
- Configuration-driven tenant strategies

### **Ready for Phase 2**
- Repository implementation with tenant isolation
- Query building system for different tenant types
- Performance monitoring foundation

---

## 🏗️ **Phase 2: Repository Layer & Tenant Resolution (7 days)**

### **Phase Goals**
- Implement abstract repository pattern with tenant-awareness
- Create dynamic query building for different tenant types
- Set up performance monitoring and metrics
- Build foundation for seamless tenant strategy migration

### **Day 8-9: Abstract Repository Implementation**

#### **Task 2.1: Base Repository with Tenant Support**

```go
// internal/infrastructure/repository/base_repository.go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
    "github.com/google/uuid"
)

// QueryBuilder handles dynamic query construction based on tenant type
type QueryBuilder interface {
    Select(columns ...string) QueryBuilder
    From(table string) QueryBuilder
    Where(condition string, args ...interface{}) QueryBuilder
    TenantWhere(storefrontID uuid.UUID) QueryBuilder
    Join(joinType, table, condition string) QueryBuilder
    OrderBy(column, direction string) QueryBuilder
    Limit(limit int) QueryBuilder
    Offset(offset int) QueryBuilder
    Build() (query string, args []interface{})
}

type queryBuilder struct {
    tenantCtx    *tenant.TenantContext
    selectCols   []string
    fromTable    string
    whereConds   []whereCondition
    joins        []joinClause
    orderBy      []orderClause
    limitVal     *int
    offsetVal    *int
    argCounter   int
    args         []interface{}
}

type whereCondition struct {
    condition string
    args      []interface{}
}

type joinClause struct {
    joinType  string
    table     string
    condition string
}

type orderClause struct {
    column    string
    direction string
}

func NewQueryBuilder(tenantCtx *tenant.TenantContext) QueryBuilder {
    return &queryBuilder{
        tenantCtx:  tenantCtx,
        selectCols: make([]string, 0),
        whereConds: make([]whereCondition, 0),
        joins:      make([]joinClause, 0),
        orderBy:    make([]orderClause, 0),
        args:       make([]interface{}, 0),
        argCounter: 0,
    }
}

func (qb *queryBuilder) Select(columns ...string) QueryBuilder {
    qb.selectCols = append(qb.selectCols, columns...)
    return qb
}

func (qb *queryBuilder) From(table string) QueryBuilder {
    // Add schema prefix for schema-based tenancy
    if qb.tenantCtx.TenantType == tenant.TenantTypeSchema {
        qb.fromTable = fmt.Sprintf("tenant_%s.%s", qb.tenantCtx.StorefrontID.String(), table)
    } else {
        qb.fromTable = table
    }
    return qb
}

func (qb *queryBuilder) Where(condition string, args ...interface{}) QueryBuilder {
    // Adjust placeholder numbers
    adjustedCondition := qb.adjustPlaceholders(condition, len(args))
    qb.whereConds = append(qb.whereConds, whereCondition{
        condition: adjustedCondition,
        args:      args,
    })
    qb.args = append(qb.args, args...)
    return qb
}

func (qb *queryBuilder) TenantWhere(storefrontID uuid.UUID) QueryBuilder {
    // Only add storefront_id filter for shared database
    if qb.tenantCtx.TenantType == tenant.TenantTypeShared {
        qb.argCounter++
        condition := fmt.Sprintf("storefront_id = $%d", qb.argCounter)
        qb.whereConds = append(qb.whereConds, whereCondition{
            condition: condition,
            args:      []interface{}{storefrontID},
        })
        qb.args = append(qb.args, storefrontID)
    }
    return qb
}

func (qb *queryBuilder) Join(joinType, table, condition string) QueryBuilder {
    // Add schema prefix for joins too
    if qb.tenantCtx.TenantType == tenant.TenantTypeSchema {
        table = fmt.Sprintf("tenant_%s.%s", qb.tenantCtx.StorefrontID.String(), table)
    }
    
    qb.joins = append(qb.joins, joinClause{
        joinType:  joinType,
        table:     table,
        condition: condition,
    })
    return qb
}

func (qb *queryBuilder) OrderBy(column, direction string) QueryBuilder {
    qb.orderBy = append(qb.orderBy, orderClause{
        column:    column,
        direction: direction,
    })
    return qb
}

func (qb *queryBuilder) Limit(limit int) QueryBuilder {
    qb.limitVal = &limit
    return qb
}

func (qb *queryBuilder) Offset(offset int) QueryBuilder {
    qb.offsetVal = &offset
    return qb
}

func (qb *queryBuilder) Build() (string, []interface{}) {
    var query strings.Builder
    
    // SELECT clause
    query.WriteString("SELECT ")
    if len(qb.selectCols) > 0 {
        query.WriteString(strings.Join(qb.selectCols, ", "))
    } else {
        query.WriteString("*")
    }
    
    // FROM clause
    if qb.fromTable != "" {
        query.WriteString(" FROM ")
        query.WriteString(qb.fromTable)
    }
    
    // JOIN clauses
    for _, join := range qb.joins {
        query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.joinType, join.table, join.condition))
    }
    
    // WHERE clause
    if len(qb.whereConds) > 0 {
        query.WriteString(" WHERE ")
        conditions := make([]string, len(qb.whereConds))
        for i, cond := range qb.whereConds {
            conditions[i] = cond.condition
        }
        query.WriteString(strings.Join(conditions, " AND "))
    }
    
    // ORDER BY clause
    if len(qb.orderBy) > 0 {
        query.WriteString(" ORDER BY ")
        orderClauses := make([]string, len(qb.orderBy))
        for i, order := range qb.orderBy {
            orderClauses[i] = fmt.Sprintf("%s %s", order.column, order.direction)
        }
        query.WriteString(strings.Join(orderClauses, ", "))
    }
    
    // LIMIT clause
    if qb.limitVal != nil {
        qb.argCounter++
        query.WriteString(fmt.Sprintf(" LIMIT $%d", qb.argCounter))
        qb.args = append(qb.args, *qb.limitVal)
    }
    
    // OFFSET clause
    if qb.offsetVal != nil {
        qb.argCounter++
        query.WriteString(fmt.Sprintf(" OFFSET $%d", qb.argCounter))
        qb.args = append(qb.args, *qb.offsetVal)
    }
    
    return query.String(), qb.args
}

func (qb *queryBuilder) adjustPlaceholders(condition string, argCount int) string {
    result := condition
    for i := 1; i <= argCount; i++ {
        qb.argCounter++
        oldPlaceholder := fmt.Sprintf("$%d", i)
        newPlaceholder := fmt.Sprintf("$%d", qb.argCounter)
        result = strings.Replace(result, oldPlaceholder, newPlaceholder, 1)
    }
    return result
}

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
    db           *sql.DB
    tenantResolver tenant.TenantResolver
}

func NewBaseRepository(db *sql.DB, tenantResolver tenant.TenantResolver) *BaseRepository {
    return &BaseRepository{
        db:           db,
        tenantResolver: tenantResolver,
    }
}

func (br *BaseRepository) GetDB(ctx context.Context, storefrontID uuid.UUID) (*sql.DB, error) {
    return br.tenantResolver.GetDatabaseConnection(ctx, storefrontID)
}

func (br *BaseRepository) NewQueryBuilder(tenantCtx *tenant.TenantContext) QueryBuilder {
    return NewQueryBuilder(tenantCtx)
}
```

#### **Task 2.2: Customer Repository Implementation**

```go
// internal/infrastructure/repository/customer_repository.go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
    "github.com/google/uuid"
    "github.com/lib/pq"
)

type customerRepository struct {
    *BaseRepository
    cache tenant.TenantCache
}

func NewCustomerRepository(
    db *sql.DB,
    tenantResolver tenant.TenantResolver,
    cache tenant.TenantCache,
) repository.CustomerRepository {
    return &customerRepository{
        BaseRepository: NewBaseRepository(db, tenantResolver),
        cache:          cache,
    }
}

func (r *customerRepository) Create(ctx context.Context, customer *entity.Customer) error {
    // Validate before creation
    if err := customer.Validate(); err != nil {
        return err
    }
    
    // Get appropriate database connection
    db, err := r.GetDB(ctx, customer.StorefrontID)
    if err != nil {
        return fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Create tenant context for query building
    storefront, err := r.tenantResolver.GetStorefrontBySlug(ctx, "")
    if err != nil {
        // For direct customer creation, we might not have the slug
        // Use default shared context
        storefront = &entity.Storefront{ID: customer.StorefrontID}
    }
    tenantCtx := r.tenantResolver.CreateTenantContext(storefront)
    
    // Generate new ID if not set
    if customer.ID == uuid.Nil {
        customer.ID = uuid.New()
    }
    
    // Set timestamps
    now := time.Now()
    customer.CreatedAt = now
    customer.UpdatedAt = now
    
    // Build insert query using query builder
    query, args := r.NewQueryBuilder(tenantCtx).
        Select().
        From("customers").
        Build()
    
    // Manual query construction for INSERT (QueryBuilder is mainly for SELECT)
    insertQuery := `
        INSERT INTO customers (
            id, storefront_id, email, phone, password_hash, password_salt,
            first_name, last_name, date_of_birth, gender, profile_picture,
            email_verified_at, phone_verified_at, email_verification_token,
            phone_verification_token, status, total_orders, total_spent,
            average_order_value, lifetime_value, preferences, tags, notes,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
            $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25
        )
    `
    
    _, err = db.ExecContext(ctx, insertQuery,
        customer.ID, customer.StorefrontID, customer.Email, customer.Phone,
        customer.PasswordHash, customer.PasswordSalt, customer.FirstName,
        customer.LastName, customer.DateOfBirth, customer.Gender,
        customer.ProfilePicture, customer.EmailVerifiedAt,
        customer.PhoneVerifiedAt, customer.EmailVerificationToken,
        customer.PhoneVerificationToken, customer.Status,
        customer.TotalOrders, customer.TotalSpent,
        customer.AverageOrderValue, customer.LifetimeValue,
        customer.Preferences, pq.Array(customer.Tags), customer.Notes,
        customer.CreatedAt, customer.UpdatedAt,
    )
    
    if err != nil {
        if isUniqueViolationError(err) {
            return &domainErrors.DomainError{
                Code:    "EMAIL_ALREADY_EXISTS",
                Message: "Email already exists for this storefront",
            }
        }
        return fmt.Errorf("failed to create customer: %w", err)
    }
    
    return nil
}

func (r *customerRepository) GetByID(ctx context.Context, storefrontID, customerID uuid.UUID) (*entity.Customer, error) {
    // Check cache first
    if customer := r.cache.GetCustomer(storefrontID.String(), customerID.String()); customer != nil {
        return customer, nil
    }
    
    // Get database connection
    db, err := r.GetDB(ctx, storefrontID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Get tenant context
    storefront, err := r.tenantResolver.GetStorefrontBySlug(ctx, "")
    if err != nil {
        storefront = &entity.Storefront{ID: storefrontID}
    }
    tenantCtx := r.tenantResolver.CreateTenantContext(storefront)
    
    // Build query
    query, args := r.NewQueryBuilder(tenantCtx).
        Select(
            "id", "storefront_id", "email", "phone", "password_hash", "password_salt",
            "first_name", "last_name", "date_of_birth", "gender", "profile_picture",
            "email_verified_at", "phone_verified_at", "email_verification_token",
            "phone_verification_token", "status", "last_login_at",
            "total_orders", "total_spent", "average_order_value", "lifetime_value",
            "preferences", "tags", "notes", "created_at", "updated_at", "deleted_at",
        ).
        From("customers").
        TenantWhere(storefrontID).
        Where("id = $1 AND deleted_at IS NULL", customerID).
        Build()
    
    row := db.QueryRowContext(ctx, query, args...)
    
    customer := &entity.Customer{}
    err = row.Scan(
        &customer.ID, &customer.StorefrontID, &customer.Email, &customer.Phone,
        &customer.PasswordHash, &customer.PasswordSalt, &customer.FirstName,
        &customer.LastName, &customer.DateOfBirth, &customer.Gender,
        &customer.ProfilePicture, &customer.EmailVerifiedAt,
        &customer.PhoneVerifiedAt, &customer.EmailVerificationToken,
        &customer.PhoneVerificationToken, &customer.Status,
        &customer.LastLoginAt, &customer.TotalOrders, &customer.TotalSpent,
        &customer.AverageOrderValue, &customer.LifetimeValue,
        &customer.Preferences, pq.Array(&customer.Tags), &customer.Notes,
        &customer.CreatedAt, &customer.UpdatedAt, &customer.DeletedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, &domainErrors.DomainError{
                Code:    "CUSTOMER_NOT_FOUND",
                Message: "Customer not found",
            }
        }
        return nil, fmt.Errorf("failed to get customer: %w", err)
    }
    
    // Cache the customer for 15 minutes
    r.cache.SetCustomer(storefrontID.String(), customerID.String(), customer, time.Minute*15)
    
    return customer, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error) {
    // Get database connection
    db, err := r.GetDB(ctx, storefrontID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Get tenant context
    storefront, err := r.tenantResolver.GetStorefrontBySlug(ctx, "")
    if err != nil {
        storefront = &entity.Storefront{ID: storefrontID}
    }
    tenantCtx := r.tenantResolver.CreateTenantContext(storefront)
    
    // Normalize email
    normalizedEmail := strings.ToLower(strings.TrimSpace(email))
    
    // Build query
    query, args := r.NewQueryBuilder(tenantCtx).
        Select(
            "id", "storefront_id", "email", "phone", "password_hash", "password_salt",
            "first_name", "last_name", "date_of_birth", "gender", "profile_picture",
            "email_verified_at", "phone_verified_at", "email_verification_token",
            "phone_verification_token", "status", "last_login_at",
            "total_orders", "total_spent", "average_order_value", "lifetime_value",
            "preferences", "tags", "notes", "created_at", "updated_at", "deleted_at",
        ).
        From("customers").
        TenantWhere(storefrontID).
        Where("LOWER(email) = $1 AND deleted_at IS NULL", normalizedEmail).
        Build()
    
    row := db.QueryRowContext(ctx, query, args...)
    
    customer := &entity.Customer{}
    err = row.Scan(
        &customer.ID, &customer.StorefrontID, &customer.Email, &customer.Phone,
        &customer.PasswordHash, &customer.PasswordSalt, &customer.FirstName,
        &customer.LastName, &customer.DateOfBirth, &customer.Gender,
        &customer.ProfilePicture, &customer.EmailVerifiedAt,
        &customer.PhoneVerifiedAt, &customer.EmailVerificationToken,
        &customer.PhoneVerificationToken, &customer.Status,
        &customer.LastLoginAt, &customer.TotalOrders, &customer.TotalSpent,
        &customer.AverageOrderValue, &customer.LifetimeValue,
        &customer.Preferences, pq.Array(&customer.Tags), &customer.Notes,
        &customer.CreatedAt, &customer.UpdatedAt, &customer.DeletedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, &domainErrors.DomainError{
                Code:    "CUSTOMER_NOT_FOUND",
                Message: "Customer not found",
            }
        }
        return nil, fmt.Errorf("failed to get customer by email: %w", err)
    }
    
    return customer, nil
}

func (r *customerRepository) GetByStorefront(ctx context.Context, req *repository.GetCustomersRequest) (*repository.CustomerListResponse, error) {
    // Get database connection
    db, err := r.GetDB(ctx, req.StorefrontID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Get tenant context
    storefront, err := r.tenantResolver.GetStorefrontBySlug(ctx, "")
    if err != nil {
        storefront = &entity.Storefront{ID: req.StorefrontID}
    }
    tenantCtx := r.tenantResolver.CreateTenantContext(storefront)
    
    // Build base query
    queryBuilder := r.NewQueryBuilder(tenantCtx).
        Select(
            "id", "storefront_id", "email", "phone", "first_name", "last_name",
            "date_of_birth", "gender", "profile_picture", "email_verified_at",
            "phone_verified_at", "status", "last_login_at", "total_orders",
            "total_spent", "average_order_value", "lifetime_value",
            "preferences", "tags", "notes", "created_at", "updated_at",
        ).
        From("customers").
        TenantWhere(req.StorefrontID).
        Where("deleted_at IS NULL")
    
    // Add search filter
    if req.Search != "" {
        searchPattern := "%" + req.Search + "%"
        queryBuilder = queryBuilder.Where(
            "(LOWER(first_name) LIKE LOWER($1) OR LOWER(last_name) LIKE LOWER($2) OR LOWER(email) LIKE LOWER($3))",
            searchPattern, searchPattern, searchPattern,
        )
    }
    
    // Add status filter
    if req.Status != nil {
        queryBuilder = queryBuilder.Where("status = $1", *req.Status)
    }
    
    // Add ordering
    orderBy := "created_at"
    direction := "DESC"
    if req.OrderBy != "" {
        orderBy = req.OrderBy
    }
    if req.SortDesc {
        direction = "DESC"
    } else {
        direction = "ASC"
    }
    queryBuilder = queryBuilder.OrderBy(orderBy, direction)
    
    // Add pagination
    offset := (req.Page - 1) * req.PageSize
    queryBuilder = queryBuilder.Limit(req.PageSize).Offset(offset)
    
    // Execute query
    query, args := queryBuilder.Build()
    rows, err := db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query customers: %w", err)
    }
    defer rows.Close()
    
    customers := make([]*entity.Customer, 0)
    for rows.Next() {
        customer := &entity.Customer{}
        err := rows.Scan(
            &customer.ID, &customer.StorefrontID, &customer.Email, &customer.Phone,
            &customer.FirstName, &customer.LastName, &customer.DateOfBirth,
            &customer.Gender, &customer.ProfilePicture, &customer.EmailVerifiedAt,
            &customer.PhoneVerifiedAt, &customer.Status, &customer.LastLoginAt,
            &customer.TotalOrders, &customer.TotalSpent, &customer.AverageOrderValue,
            &customer.LifetimeValue, &customer.Preferences, pq.Array(&customer.Tags),
            &customer.Notes, &customer.CreatedAt, &customer.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan customer: %w", err)
        }
        customers = append(customers, customer)
    }
    
    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating customers: %w", err)
    }
    
    // Get total count
    countQuery, countArgs := r.NewQueryBuilder(tenantCtx).
        Select("COUNT(*)").
        From("customers").
        TenantWhere(req.StorefrontID).
        Where("deleted_at IS NULL").
        Build()
    
    if req.Search != "" {
        // Add same search filters to count query
        searchPattern := "%" + req.Search + "%"
        countQuery += " AND (LOWER(first_name) LIKE LOWER($2) OR LOWER(last_name) LIKE LOWER($3) OR LOWER(email) LIKE LOWER($4))"
        countArgs = append(countArgs, searchPattern, searchPattern, searchPattern)
    }
    
    if req.Status != nil {
        countQuery += " AND status = $" + fmt.Sprintf("%d", len(countArgs)+1)
        countArgs = append(countArgs, *req.Status)
    }
    
    var total int
    err = db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
    if err != nil {
        return nil, fmt.Errorf("failed to count customers: %w", err)
    }
    
    return &repository.CustomerListResponse{
        Customers:  customers,
        Total:      total,
        Page:       req.Page,
        PageSize:   req.PageSize,
        TotalPages: (total + req.PageSize - 1) / req.PageSize,
    }, nil
}

// Helper functions
func isUniqueViolationError(err error) bool {
    if pqErr, ok := err.(*pq.Error); ok {
        return pqErr.Code == "23505" // unique_violation
    }
    return false
}
```

**Success Criteria Day 8-9:**
- ✅ Abstract query builder handles different tenant types
- ✅ Customer repository with full multi-tenant support
- ✅ Caching integration for performance
- ✅ Error handling with domain-specific errors

### **Day 10-11: Storefront & Address Repositories**

#### **Task 2.3: Storefront Repository Implementation**

```go
// internal/infrastructure/repository/storefront_repository.go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "time"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
    "github.com/google/uuid"
)

type storefrontRepository struct {
    *BaseRepository
    cache tenant.TenantCache
}

func NewStorefrontRepository(
    db *sql.DB,
    tenantResolver tenant.TenantResolver,
    cache tenant.TenantCache,
) repository.StorefrontRepository {
    return &storefrontRepository{
        BaseRepository: NewBaseRepository(db, tenantResolver),
        cache:          cache,
    }
}

func (r *storefrontRepository) Create(ctx context.Context, storefront *entity.Storefront) error {
    if err := storefront.Validate(); err != nil {
        return err
    }
    
    // For storefronts, we always use the main database
    db := r.db
    
    // Generate ID if not set
    if storefront.ID == uuid.Nil {
        storefront.ID = uuid.New()
    }
    
    // Set timestamps
    now := time.Now()
    storefront.CreatedAt = now
    storefront.UpdatedAt = now
    
    // Normalize slug
    storefront.Slug = strings.ToLower(strings.TrimSpace(storefront.Slug))
    
    query := `
        INSERT INTO storefronts (
            id, seller_id, name, slug, description, domain, subdomain,
            status, currency, language, timezone, settings, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
    `
    
    _, err := db.ExecContext(ctx, query,
        storefront.ID, storefront.SellerID, storefront.Name, storefront.Slug,
        storefront.Description, storefront.Domain, storefront.Subdomain,
        storefront.Status, storefront.Currency, storefront.Language,
        storefront.Timezone, storefront.Settings, storefront.CreatedAt,
        storefront.UpdatedAt,
    )
    
    if err != nil {
        if isUniqueViolationError(err) {
            if strings.Contains(err.Error(), "slug") {
                return &domainErrors.DomainError{
                    Code:    "SLUG_ALREADY_EXISTS",
                    Message: "Storefront slug already exists",
                }
            }
            if strings.Contains(err.Error(), "domain") {
                return &domainErrors.DomainError{
                    Code:    "DOMAIN_ALREADY_EXISTS",
                    Message: "Domain already exists",
                }
            }
        }
        return fmt.Errorf("failed to create storefront: %w", err)
    }
    
    return nil
}

func (r *storefrontRepository) GetBySlug(ctx context.Context, slug string) (*entity.Storefront, error) {
    // Check cache first
    if storefront := r.cache.GetStorefront(slug); storefront != nil {
        return storefront, nil
    }
    
    normalizedSlug := strings.ToLower(strings.TrimSpace(slug))
    
    query := `
        SELECT id, seller_id, name, slug, description, domain, subdomain,
               status, currency, language, timezone, settings, 
               created_at, updated_at, deleted_at
        FROM storefronts
        WHERE slug = $1 AND deleted_at IS NULL
    `
    
    row := r.db.QueryRowContext(ctx, query, normalizedSlug)
    
    storefront := &entity.Storefront{}
    err := row.Scan(
        &storefront.ID, &storefront.SellerID, &storefront.Name,
        &storefront.Slug, &storefront.Description, &storefront.Domain,
        &storefront.Subdomain, &storefront.Status, &storefront.Currency,
        &storefront.Language, &storefront.Timezone, &storefront.Settings,
        &storefront.CreatedAt, &storefront.UpdatedAt, &storefront.DeletedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, &domainErrors.DomainError{
                Code:    "STOREFRONT_NOT_FOUND",
                Message: "Storefront not found",
            }
        }
        return nil, fmt.Errorf("failed to get storefront: %w", err)
    }
    
    // Cache for 1 hour
    r.cache.SetStorefront(slug, storefront, time.Hour)
    
    return storefront, nil
}

func (r *storefrontRepository) GetStorefrontStats(ctx context.Context, storefrontID uuid.UUID) (*repository.StorefrontStats, error) {
    // Get database connection for this storefront
    db, err := r.GetDB(ctx, storefrontID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Get tenant context
    storefront := &entity.Storefront{ID: storefrontID}
    tenantCtx := r.tenantResolver.CreateTenantContext(storefront)
    
    // Query stats based on tenant type
    var customerCountQuery, orderCountQuery, revenueQuery string
    var args []interface{}
    
    if tenantCtx.TenantType == tenant.TenantTypeShared {
        // Shared database - include storefront_id filter
        customerCountQuery = "SELECT COUNT(*) FROM customers WHERE storefront_id = $1 AND deleted_at IS NULL"
        orderCountQuery = "SELECT COUNT(*) FROM orders WHERE storefront_id = $1 AND deleted_at IS NULL"
        revenueQuery = "SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE storefront_id = $1 AND status = 'completed' AND deleted_at IS NULL"
        args = []interface{}{storefrontID}
    } else {
        // Schema or database per tenant - no storefront_id filter needed
        customerCountQuery = "SELECT COUNT(*) FROM customers WHERE deleted_at IS NULL"
        orderCountQuery = "SELECT COUNT(*) FROM orders WHERE deleted_at IS NULL"
        revenueQuery = "SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed' AND deleted_at IS NULL"
        args = []interface{}{}
    }
    
    stats := &repository.StorefrontStats{}
    
    // Customer count
    err = db.QueryRowContext(ctx, customerCountQuery, args...).Scan(&stats.CustomerCount)
    if err != nil {
        return nil, fmt.Errorf("failed to get customer count: %w", err)
    }
    
    // Order count (this would work when orders table exists)
    err = db.QueryRowContext(ctx, orderCountQuery, args...).Scan(&stats.OrderCount)
    if err != nil {
        // Orders table might not exist yet, set to 0
        stats.OrderCount = 0
    }
    
    // Revenue (this would work when orders table exists)
    err = db.QueryRowContext(ctx, revenueQuery, args...).Scan(&stats.TotalRevenue)
    if err != nil {
        // Orders table might not exist yet, set to 0
        stats.TotalRevenue = 0
    }
    
    // Measure average query time (simplified)
    start := time.Now()
    _, err = db.QueryContext(ctx, "SELECT 1")
    if err == nil {
        stats.AvgQueryTime = time.Since(start).Milliseconds()
    }
    
    return stats, nil
}
```

#### **Task 2.4: Customer Address Repository**

```go
// internal/infrastructure/repository/customer_address_repository.go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
    "github.com/google/uuid"
)

type customerAddressRepository struct {
    *BaseRepository
}

func NewCustomerAddressRepository(
    db *sql.DB,
    tenantResolver tenant.TenantResolver,
) repository.CustomerAddressRepository {
    return &customerAddressRepository{
        BaseRepository: NewBaseRepository(db, tenantResolver),
    }
}

func (r *customerAddressRepository) Create(ctx context.Context, address *entity.CustomerAddress) error {
    if err := address.Validate(); err != nil {
        return err
    }
    
    // Get customer first to determine storefront
    customer, err := r.getCustomerByID(ctx, address.CustomerID)
    if err != nil {
        return fmt.Errorf("failed to get customer: %w", err)
    }
    
    // Get database connection
    db, err := r.GetDB(ctx, customer.StorefrontID)
    if err != nil {
        return fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Generate ID if not set
    if address.ID == uuid.Nil {
        address.ID = uuid.New()
    }
    
    // Set timestamps
    now := time.Now()
    address.CreatedAt = now
    address.UpdatedAt = now
    
    query := `
        INSERT INTO customer_addresses (
            id, customer_id, type, label, first_name, last_name, company,
            address_line1, address_line2, city, province, postal_code,
            country, phone, is_default, latitude, longitude, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
        )
    `
    
    _, err = db.ExecContext(ctx, query,
        address.ID, address.CustomerID, address.Type, address.Label,
        address.FirstName, address.LastName, address.Company,
        address.AddressLine1, address.AddressLine2, address.City,
        address.Province, address.PostalCode, address.Country,
        address.Phone, address.IsDefault, address.Latitude,
        address.Longitude, address.CreatedAt, address.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create address: %w", err)
    }
    
    return nil
}

func (r *customerAddressRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error) {
    // Get customer first to determine storefront
    customer, err := r.getCustomerByID(ctx, customerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get customer: %w", err)
    }
    
    // Get database connection
    db, err := r.GetDB(ctx, customer.StorefrontID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
    }
    
    query := `
        SELECT id, customer_id, type, label, first_name, last_name, company,
               address_line1, address_line2, city, province, postal_code,
               country, phone, is_default, latitude, longitude, created_at, updated_at
        FROM customer_addresses
        WHERE customer_id = $1
        ORDER BY is_default DESC, created_at ASC
    `
    
    rows, err := db.QueryContext(ctx, query, customerID)
    if err != nil {
        return nil, fmt.Errorf("failed to query addresses: %w", err)
    }
    defer rows.Close()
    
    addresses := make([]*entity.CustomerAddress, 0)
    for rows.Next() {
        address := &entity.CustomerAddress{}
        err := rows.Scan(
            &address.ID, &address.CustomerID, &address.Type, &address.Label,
            &address.FirstName, &address.LastName, &address.Company,
            &address.AddressLine1, &address.AddressLine2, &address.City,
            &address.Province, &address.PostalCode, &address.Country,
            &address.Phone, &address.IsDefault, &address.Latitude,
            &address.Longitude, &address.CreatedAt, &address.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan address: %w", err)
        }
        addresses = append(addresses, address)
    }
    
    return addresses, nil
}

func (r *customerAddressRepository) SetAsDefault(ctx context.Context, customerID, addressID uuid.UUID) error {
    // Get customer first to determine storefront
    customer, err := r.getCustomerByID(ctx, customerID)
    if err != nil {
        return fmt.Errorf("failed to get customer: %w", err)
    }
    
    // Get database connection
    db, err := r.GetDB(ctx, customer.StorefrontID)
    if err != nil {
        return fmt.Errorf("failed to get database connection: %w", err)
    }
    
    // Start transaction
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Unset all other addresses as default
    _, err = tx.ExecContext(ctx,
        "UPDATE customer_addresses SET is_default = FALSE WHERE customer_id = $1",
        customerID,
    )
    if err != nil {
        return fmt.Errorf("failed to unset default addresses: %w", err)
    }
    
    // Set the specified address as default
    result, err := tx.ExecContext(ctx,
        "UPDATE customer_addresses SET is_default = TRUE WHERE id = $1 AND customer_id = $2",
        addressID, customerID,
    )
    if err != nil {
        return fmt.Errorf("failed to set default address: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return &domainErrors.DomainError{
            Code:    "ADDRESS_NOT_FOUND",
            Message: "Address not found",
        }
    }
    
    return tx.Commit()
}

// Helper method to get customer by ID (simplified version)
func (r *customerAddressRepository) getCustomerByID(ctx context.Context, customerID uuid.UUID) (*entity.Customer, error) {
    query := `
        SELECT id, storefront_id, email, first_name, last_name, created_at
        FROM customers
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    row := r.db.QueryRowContext(ctx, query, customerID)
    
    customer := &entity.Customer{}
    err := row.Scan(
        &customer.ID, &customer.StorefrontID, &customer.Email,
        &customer.FirstName, &customer.LastName, &customer.CreatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, &domainErrors.DomainError{
                Code:    "CUSTOMER_NOT_FOUND",
                Message: "Customer not found",
            }
        }
        return nil, err
    }
    
    return customer, nil
}
```

**Success Criteria Day 10-11:**
- ✅ Storefront repository with caching and stats
- ✅ Customer address repository with transaction support
- ✅ Multi-tenant query building working across all repositories
- ✅ Error handling and validation properly implemented

### **Day 12-14: Performance Monitoring & Testing**

#### **Task 2.5: Repository Performance Monitoring**

```go
// internal/infrastructure/metrics/repository_metrics.go
package metrics

import (
    "context"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    repositoryQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "repository_query_duration_seconds",
        Help: "Duration of repository queries",
        Buckets: prometheus.DefBuckets,
    }, []string{"repository", "method", "tenant_type"})
    
    repositoryQueryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "repository_query_total",
        Help: "Total number of repository queries",
    }, []string{"repository", "method", "tenant_type", "status"})
    
    tenantQueryLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name: "tenant_query_latency_seconds",
        Help: "Query latency by tenant",
        Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
    }, []string{"storefront_id", "tenant_type"})
)

// MetricsWrapper wraps repository methods with metrics collection
func WithMetrics(repository, method string) func(ctx context.Context, tenantType string, fn func() error) error {
    return func(ctx context.Context, tenantType string, fn func() error) error {
        start := time.Now()
        status := "success"
        
        err := fn()
        
        if err != nil {
            status = "error"
        }
        
        duration := time.Since(start).Seconds()
        
        repositoryQueryDuration.WithLabelValues(repository, method, tenantType).Observe(duration)
        repositoryQueryTotal.WithLabelValues(repository, method, tenantType, status).Inc()
        
        return err
    }
}

// RecordTenantLatency records query latency for a specific tenant
func RecordTenantLatency(storefrontID, tenantType string, duration time.Duration) {
    tenantQueryLatency.WithLabelValues(storefrontID, tenantType).Observe(duration.Seconds())
}
```

#### **Task 2.6: Integration Tests**

```go
// internal/infrastructure/repository/customer_repository_test.go
package repository_test

import (
    "context"
    "testing"
    "time"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCustomerRepository_Create(t *testing.T) {
    // Set up test database and dependencies
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    tenantResolver := setupTestTenantResolver(db)
    cache := tenant.NewInMemoryTenantCache()
    repo := repository.NewCustomerRepository(db, tenantResolver, cache)
    
    // Create test storefront
    storefront := &entity.Storefront{
        ID:       uuid.New(),
        SellerID: uuid.New(),
        Name:     "Test Store",
        Slug:     "test-store",
        Status:   entity.StorefrontStatusActive,
    }
    
    // Create customer
    customer := &entity.Customer{
        StorefrontID: storefront.ID,
        Email:        "test@example.com",
        FirstName:    "John",
        LastName:     "Doe",
        PasswordHash: "hashedpassword",
        PasswordSalt: "salt",
        Status:       entity.CustomerStatusActive,
    }
    
    err := repo.Create(context.Background(), customer)
    require.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, customer.ID)
    assert.False(t, customer.CreatedAt.IsZero())
    assert.False(t, customer.UpdatedAt.IsZero())
}

func TestCustomerRepository_GetByEmail_MultiTenant(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    tenantResolver := setupTestTenantResolver(db)
    cache := tenant.NewInMemoryTenantCache()
    repo := repository.NewCustomerRepository(db, tenantResolver, cache)
    
    // Create two different storefronts
    storefront1 := uuid.New()
    storefront2 := uuid.New()
    
    // Create customers with same email but different storefronts
    customer1 := &entity.Customer{
        StorefrontID: storefront1,
        Email:        "john@example.com",
        FirstName:    "John",
        LastName:     "Store1",
        PasswordHash: "hash1",
        PasswordSalt: "salt1",
        Status:       entity.CustomerStatusActive,
    }
    
    customer2 := &entity.Customer{
        StorefrontID: storefront2,
        Email:        "john@example.com", // Same email, different storefront
        FirstName:    "John",
        LastName:     "Store2",
        PasswordHash: "hash2",
        PasswordSalt: "salt2",
        Status:       entity.CustomerStatusActive,
    }
    
    // Create both customers
    err := repo.Create(context.Background(), customer1)
    require.NoError(t, err)
    
    err = repo.Create(context.Background(), customer2)
    require.NoError(t, err)
    
    // Test tenant isolation - should get different customers
    ctx := context.Background()
    
    found1, err := repo.GetByEmail(ctx, storefront1, "john@example.com")
    require.NoError(t, err)
    assert.Equal(t, customer1.ID, found1.ID)
    assert.Equal(t, "Store1", found1.LastName)
    
    found2, err := repo.GetByEmail(ctx, storefront2, "john@example.com")
    require.NoError(t, err)
    assert.Equal(t, customer2.ID, found2.ID)
    assert.Equal(t, "Store2", found2.LastName)
    
    // Verify they are different customers
    assert.NotEqual(t, found1.ID, found2.ID)
}

func TestCustomerRepository_Performance(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    tenantResolver := setupTestTenantResolver(db)
    cache := tenant.NewInMemoryTenantCache()
    repo := repository.NewCustomerRepository(db, tenantResolver, cache)
    
    storefront := uuid.New()
    
    // Create 1000 customers
    for i := 0; i < 1000; i++ {
        customer := &entity.Customer{
            StorefrontID: storefront,
            Email:        fmt.Sprintf("customer%d@example.com", i),
            FirstName:    "Customer",
            LastName:     fmt.Sprintf("User%d", i),
            PasswordHash: "hash",
            PasswordSalt: "salt",
            Status:       entity.CustomerStatusActive,
        }
        
        err := repo.Create(context.Background(), customer)
        require.NoError(t, err)
    }
    
    // Test query performance
    start := time.Now()
    
    req := &repository.GetCustomersRequest{
        StorefrontID: storefront,
        Page:         1,
        PageSize:     50,
        Search:       "Customer",
        OrderBy:      "created_at",
        SortDesc:     true,
    }
    
    response, err := repo.GetByStorefront(context.Background(), req)
    require.NoError(t, err)
    
    duration := time.Since(start)
    
    assert.Equal(t, 50, len(response.Customers))
    assert.Equal(t, 1000, response.Total)
    assert.Less(t, duration.Milliseconds(), int64(100)) // Should be under 100ms
    
    t.Logf("Query took %v to fetch %d customers from %d total", duration, len(response.Customers), response.Total)
}

// Test helper functions
func setupTestDB(t *testing.T) (*sql.DB, func()) {
    // Implementation would set up a test database
    // Return database connection and cleanup function
    panic("implement setupTestDB")
}

func setupTestTenantResolver(db *sql.DB) tenant.TenantResolver {
    config := &tenant.TenantConfig{
        DefaultTenantType:        tenant.TenantTypeShared,
        SharedDatabaseURL:       "test-db-url",
        MaxConnectionsPerTenant: 10,
        MigrationThresholds: tenant.MigrationThresholds{
            CustomerCount: 10000,
            OrderCount:    50000,
            AvgQueryTime:  100 * time.Millisecond,
        },
    }
    
    cache := tenant.NewInMemoryTenantCache()
    return tenant.NewTenantResolver(db, config, cache)
}
```

**Success Criteria Day 12-14:**
- ✅ Performance monitoring with Prometheus metrics
- ✅ Comprehensive integration tests
- ✅ Multi-tenant isolation tested
- ✅ Performance benchmarks established

---

## 🎯 **Phase 2 Success Criteria**

At the end of Phase 2, you should have:

### **✅ Repository Layer**
- Abstract query builder handling multiple tenant types
- Complete customer, storefront, and address repositories
- Caching integration for optimal performance
- Multi-tenant data isolation working correctly

### **✅ Performance Foundation**  
- Prometheus metrics for query monitoring
- Tenant-specific performance tracking
- Query optimization for large datasets
- Cache-first data access patterns

### **✅ Testing Infrastructure**
- Integration tests for all repositories
- Multi-tenant isolation verification
- Performance benchmarks established
- Test database setup and teardown

### **Ready for Phase 3**
- Use case layer implementation
- Business logic with customer authentication
- Registration and profile management flows
- JWT token handling with tenant scoping

---

## 💼 **Phase 3: Use Cases & Business Logic (7 days)**

### **Phase Goals**
- Implement customer registration and authentication use cases
- Build profile management and address handling business logic
- Create JWT token service with tenant-aware claims
- Implement password security and verification systems
- Set up email verification and password reset workflows

### **Day 15-16: Authentication Use Cases**

#### **Task 3.1: JWT Service with Tenant Scoping**

```go
// internal/application/service/jwt_service.go
package service

import (
    "context"
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/google/uuid"
)

type JWTClaims struct {
    jwt.RegisteredClaims
    CustomerID   string `json:"customer_id"`
    StorefrontID string `json:"storefront_id"`
    Email        string `json:"email"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Verified     bool   `json:"verified"`
    TokenType    string `json:"token_type"` // access, refresh, verification
}

type JWTService interface {
    GenerateAccessToken(ctx context.Context, customer *entity.Customer) (string, error)
    GenerateRefreshToken(ctx context.Context, customer *entity.Customer) (string, error)
    GenerateEmailVerificationToken(ctx context.Context, customer *entity.Customer) (string, error)
    GeneratePasswordResetToken(ctx context.Context, customer *entity.Customer) (string, error)
    ValidateToken(ctx context.Context, tokenString string, expectedType string) (*JWTClaims, error)
    RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, err error)
}

type jwtService struct {
    secretKey            []byte
    accessTokenDuration  time.Duration
    refreshTokenDuration time.Duration
    issuer              string
}

func NewJWTService(secretKey string, accessDuration, refreshDuration time.Duration, issuer string) JWTService {
    return &jwtService{
        secretKey:            []byte(secretKey),
        accessTokenDuration:  accessDuration,
        refreshTokenDuration: refreshDuration,
        issuer:              issuer,
    }
}

func (j *jwtService) GenerateAccessToken(ctx context.Context, customer *entity.Customer) (string, error) {
    now := time.Now()
    claims := &JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        uuid.New().String(),
            Subject:   customer.ID.String(),
            Issuer:    j.issuer,
            Audience:  []string{"storefront-" + customer.StorefrontID.String()},
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenDuration)),
        },
        CustomerID:   customer.ID.String(),
        StorefrontID: customer.StorefrontID.String(),
        Email:        customer.Email,
        FirstName:    customer.FirstName,
        LastName:     customer.LastName,
        Verified:     customer.IsEmailVerified(),
        TokenType:    "access",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(j.secretKey)
}

func (j *jwtService) GenerateRefreshToken(ctx context.Context, customer *entity.Customer) (string, error) {
    now := time.Now()
    claims := &JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        uuid.New().String(),
            Subject:   customer.ID.String(),
            Issuer:    j.issuer,
            Audience:  []string{"storefront-" + customer.StorefrontID.String()},
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenDuration)),
        },
        CustomerID:   customer.ID.String(),
        StorefrontID: customer.StorefrontID.String(),
        Email:        customer.Email,
        TokenType:    "refresh",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(j.secretKey)
}

func (j *jwtService) GenerateEmailVerificationToken(ctx context.Context, customer *entity.Customer) (string, error) {
    now := time.Now()
    claims := &JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        uuid.New().String(),
            Subject:   customer.ID.String(),
            Issuer:    j.issuer,
            Audience:  []string{"storefront-" + customer.StorefrontID.String()},
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 24 hours for email verification
        },
        CustomerID:   customer.ID.String(),
        StorefrontID: customer.StorefrontID.String(),
        Email:        customer.Email,
        TokenType:    "email_verification",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(j.secretKey)
}

func (j *jwtService) ValidateToken(ctx context.Context, tokenString string, expectedType string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return j.secretKey, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }
    
    claims, ok := token.Claims.(*JWTClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    
    // Validate token type if specified
    if expectedType != "" && claims.TokenType != expectedType {
        return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
    }
    
    return claims, nil
}

func (j *jwtService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
    claims, err := j.ValidateToken(ctx, refreshToken, "refresh")
    if err != nil {
        return "", "", fmt.Errorf("invalid refresh token: %w", err)
    }
    
    // Create a customer object from claims to generate new tokens
    customerID, err := uuid.Parse(claims.CustomerID)
    if err != nil {
        return "", "", fmt.Errorf("invalid customer ID in token: %w", err)
    }
    
    storefrontID, err := uuid.Parse(claims.StorefrontID)
    if err != nil {
        return "", "", fmt.Errorf("invalid storefront ID in token: %w", err)
    }
    
    customer := &entity.Customer{
        ID:           customerID,
        StorefrontID: storefrontID,
        Email:        claims.Email,
        FirstName:    claims.FirstName,
        LastName:     claims.LastName,
    }
    
    // Generate new tokens
    newAccessToken, err := j.GenerateAccessToken(ctx, customer)
    if err != nil {
        return "", "", fmt.Errorf("failed to generate access token: %w", err)
    }
    
    newRefreshToken, err := j.GenerateRefreshToken(ctx, customer)
    if err != nil {
        return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
    }
    
    return newAccessToken, newRefreshToken, nil
}
```

#### **Task 3.2: Customer Authentication Use Case**

```go
// internal/application/usecase/customer_auth_usecase.go
package usecase

import (
    "context"
    "crypto/rand"
    "crypto/subtle"
    "encoding/base64"
    "fmt"
    "time"
    "github.com/kirimku/smartseller-backend/internal/application/dto"
    "github.com/kirimku/smartseller-backend/internal/application/service"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    domainErrors "github.com/kirimku/smartseller-backend/internal/domain/errors"
    "github.com/kirimku/smartseller-backend/pkg/email"
    "github.com/google/uuid"
    "golang.org/x/crypto/scrypt"
)

type CustomerAuthUseCase interface {
    Register(ctx context.Context, req *dto.CustomerRegisterRequest) (*dto.CustomerRegisterResponse, error)
    Login(ctx context.Context, req *dto.CustomerLoginRequest) (*dto.CustomerLoginResponse, error)
    RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)
    VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) error
    RequestPasswordReset(ctx context.Context, req *dto.RequestPasswordResetRequest) error
    ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
    Logout(ctx context.Context, req *dto.LogoutRequest) error
}

type customerAuthUseCase struct {
    customerRepo      repository.CustomerRepository
    storefrontRepo    repository.StorefrontRepository
    jwtService        service.JWTService
    emailService      email.EmailService
    maxFailedAttempts int
    lockDuration      time.Duration
}

func NewCustomerAuthUseCase(
    customerRepo repository.CustomerRepository,
    storefrontRepo repository.StorefrontRepository,
    jwtService service.JWTService,
    emailService email.EmailService,
) CustomerAuthUseCase {
    return &customerAuthUseCase{
        customerRepo:      customerRepo,
        storefrontRepo:    storefrontRepo,
        jwtService:        jwtService,
        emailService:      emailService,
        maxFailedAttempts: 5,
        lockDuration:      15 * time.Minute,
    }
}

func (uc *customerAuthUseCase) Register(ctx context.Context, req *dto.CustomerRegisterRequest) (*dto.CustomerRegisterResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Get storefront to ensure it exists and is active
    storefront, err := uc.storefrontRepo.GetBySlug(ctx, req.StorefrontSlug)
    if err != nil {
        return nil, err
    }
    
    if storefront.Status != entity.StorefrontStatusActive {
        return nil, domainErrors.ErrStorefrontInactive
    }
    
    // Check if customer already exists
    existing, err := uc.customerRepo.GetByEmail(ctx, storefront.ID, req.Email)
    if err == nil && existing != nil {
        return nil, domainErrors.ErrEmailAlreadyExists
    }
    if err != nil && !domainErrors.IsCustomerNotFound(err) {
        return nil, fmt.Errorf("failed to check existing customer: %w", err)
    }
    
    // Validate password strength
    if err := uc.validatePasswordStrength(req.Password); err != nil {
        return nil, err
    }
    
    // Hash password
    passwordHash, passwordSalt, err := uc.hashPassword(req.Password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }
    
    // Create customer entity
    customer := &entity.Customer{
        StorefrontID:      storefront.ID,
        Email:            req.Email,
        Phone:            req.Phone,
        FirstName:        req.FirstName,
        LastName:         req.LastName,
        DateOfBirth:      req.DateOfBirth,
        Gender:           req.Gender,
        PasswordHash:     passwordHash,
        PasswordSalt:     passwordSalt,
        Status:           entity.CustomerStatusActive,
        Preferences:      dto.ToCustomerPreferences(req.Preferences),
    }
    
    // Set initial preferences if storefront requires email verification
    if storefront.Settings.RequireEmailVerification {
        customer.Status = entity.CustomerStatusInactive
        
        // Generate email verification token
        verificationToken, err := uc.jwtService.GenerateEmailVerificationToken(ctx, customer)
        if err != nil {
            return nil, fmt.Errorf("failed to generate verification token: %w", err)
        }
        customer.EmailVerificationToken = &verificationToken
    }
    
    // Create customer
    if err := uc.customerRepo.Create(ctx, customer); err != nil {
        return nil, fmt.Errorf("failed to create customer: %w", err)
    }
    
    // Send welcome/verification email
    if storefront.Settings.RequireEmailVerification {
        if err := uc.sendVerificationEmail(ctx, customer, storefront, *customer.EmailVerificationToken); err != nil {
            // Log error but don't fail registration
            fmt.Printf("Failed to send verification email: %v\n", err)
        }
    } else {
        if err := uc.sendWelcomeEmail(ctx, customer, storefront); err != nil {
            // Log error but don't fail registration
            fmt.Printf("Failed to send welcome email: %v\n", err)
        }
        
        // Mark email as verified
        now := time.Now()
        customer.EmailVerifiedAt = &now
        if err := uc.customerRepo.UpdateEmailVerification(ctx, customer.StorefrontID, customer.ID, true); err != nil {
            // Log error but don't fail registration
            fmt.Printf("Failed to mark email as verified: %v\n", err)
        }
    }
    
    // Generate tokens
    accessToken, err := uc.jwtService.GenerateAccessToken(ctx, customer)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }
    
    refreshToken, err := uc.jwtService.GenerateRefreshToken(ctx, customer)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }
    
    // Store refresh token
    expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
    if err := uc.customerRepo.UpdateRefreshToken(ctx, customer.StorefrontID, customer.ID, refreshToken, &expiresAt); err != nil {
        return nil, fmt.Errorf("failed to store refresh token: %w", err)
    }
    
    return &dto.CustomerRegisterResponse{
        Customer:     dto.ToCustomerDTO(customer),
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    expiresAt,
        RequiresEmailVerification: storefront.Settings.RequireEmailVerification,
    }, nil
}

func (uc *customerAuthUseCase) Login(ctx context.Context, req *dto.CustomerLoginRequest) (*dto.CustomerLoginResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Get storefront
    storefront, err := uc.storefrontRepo.GetBySlug(ctx, req.StorefrontSlug)
    if err != nil {
        return nil, err
    }
    
    if storefront.Status != entity.StorefrontStatusActive {
        return nil, domainErrors.ErrStorefrontInactive
    }
    
    // Get customer
    customer, err := uc.customerRepo.GetByEmail(ctx, storefront.ID, req.Email)
    if err != nil {
        if domainErrors.IsCustomerNotFound(err) {
            return nil, domainErrors.ErrInvalidCredentials
        }
        return nil, fmt.Errorf("failed to get customer: %w", err)
    }
    
    // Check if account is locked
    if customer.IsLocked() {
        return nil, domainErrors.ErrAccountLocked
    }
    
    // Check if account is suspended
    if customer.Status == entity.CustomerStatusSuspended {
        return nil, domainErrors.ErrAccountSuspended
    }
    
    // Verify password
    if !uc.verifyPassword(req.Password, customer.PasswordHash, customer.PasswordSalt) {
        // Increment failed attempts
        newAttempts := customer.FailedLoginAttempts + 1
        if err := uc.customerRepo.UpdateFailedLoginAttempts(ctx, customer.StorefrontID, customer.ID, newAttempts); err != nil {
            fmt.Printf("Failed to update login attempts: %v\n", err)
        }
        
        // Lock account if too many attempts
        if newAttempts >= uc.maxFailedAttempts {
            lockUntil := time.Now().Add(uc.lockDuration)
            if err := uc.customerRepo.LockAccount(ctx, customer.StorefrontID, customer.ID, &lockUntil); err != nil {
                fmt.Printf("Failed to lock account: %v\n", err)
            }
            return nil, domainErrors.ErrAccountLocked
        }
        
        return nil, domainErrors.ErrInvalidCredentials
    }
    
    // Check if email verification is required
    if storefront.Settings.RequireEmailVerification && !customer.IsEmailVerified() {
        return nil, domainErrors.ErrEmailNotVerified
    }
    
    // Reset failed attempts on successful login
    if customer.FailedLoginAttempts > 0 {
        if err := uc.customerRepo.UpdateFailedLoginAttempts(ctx, customer.StorefrontID, customer.ID, 0); err != nil {
            fmt.Printf("Failed to reset login attempts: %v\n", err)
        }
    }
    
    // Update last login
    if err := uc.customerRepo.UpdateLastLogin(ctx, customer.StorefrontID, customer.ID); err != nil {
        fmt.Printf("Failed to update last login: %v\n", err)
    }
    
    // Generate tokens
    accessToken, err := uc.jwtService.GenerateAccessToken(ctx, customer)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }
    
    refreshToken, err := uc.jwtService.GenerateRefreshToken(ctx, customer)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }
    
    // Store refresh token
    expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
    if err := uc.customerRepo.UpdateRefreshToken(ctx, customer.StorefrontID, customer.ID, refreshToken, &expiresAt); err != nil {
        return nil, fmt.Errorf("failed to store refresh token: %w", err)
    }
    
    return &dto.CustomerLoginResponse{
        Customer:     dto.ToCustomerDTO(customer),
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    expiresAt,
    }, nil
}

func (uc *customerAuthUseCase) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) error {
    // Validate token
    claims, err := uc.jwtService.ValidateToken(ctx, req.Token, "email_verification")
    if err != nil {
        return domainErrors.ErrInvalidEmailToken
    }
    
    // Get customer
    storefrontID, err := uuid.Parse(claims.StorefrontID)
    if err != nil {
        return domainErrors.ErrInvalidEmailToken
    }
    
    customerID, err := uuid.Parse(claims.CustomerID)
    if err != nil {
        return domainErrors.ErrInvalidEmailToken
    }
    
    customer, err := uc.customerRepo.GetByID(ctx, storefrontID, customerID)
    if err != nil {
        return err
    }
    
    // Mark email as verified
    if err := uc.customerRepo.UpdateEmailVerification(ctx, customer.StorefrontID, customer.ID, true); err != nil {
        return fmt.Errorf("failed to verify email: %w", err)
    }
    
    return nil
}

// Helper methods
func (uc *customerAuthUseCase) validatePasswordStrength(password string) error {
    if len(password) < 8 {
        return domainErrors.ErrWeakPassword
    }
    // Add more password validation rules as needed
    return nil
}

func (uc *customerAuthUseCase) hashPassword(password string) (string, string, error) {
    // Generate salt
    salt := make([]byte, 32)
    if _, err := rand.Read(salt); err != nil {
        return "", "", err
    }
    
    // Hash password with salt using scrypt
    hash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
    if err != nil {
        return "", "", err
    }
    
    return base64.StdEncoding.EncodeToString(hash), base64.StdEncoding.EncodeToString(salt), nil
}

func (uc *customerAuthUseCase) verifyPassword(password, hashedPassword, salt string) bool {
    // Decode salt
    saltBytes, err := base64.StdEncoding.DecodeString(salt)
    if err != nil {
        return false
    }
    
    // Hash provided password with salt
    hash, err := scrypt.Key([]byte(password), saltBytes, 32768, 8, 1, 32)
    if err != nil {
        return false
    }
    
    // Decode stored hash
    storedHash, err := base64.StdEncoding.DecodeString(hashedPassword)
    if err != nil {
        return false
    }
    
    // Compare hashes using constant-time comparison
    return subtle.ConstantTimeCompare(hash, storedHash) == 1
}

func (uc *customerAuthUseCase) sendVerificationEmail(ctx context.Context, customer *entity.Customer, storefront *entity.Storefront, token string) error {
    // Implementation would send verification email
    // This is a placeholder
    fmt.Printf("Sending verification email to %s for storefront %s with token %s\n", customer.Email, storefront.Name, token)
    return nil
}

func (uc *customerAuthUseCase) sendWelcomeEmail(ctx context.Context, customer *entity.Customer, storefront *entity.Storefront) error {
    // Implementation would send welcome email
    // This is a placeholder
    fmt.Printf("Sending welcome email to %s for storefront %s\n", customer.Email, storefront.Name)
    return nil
}
```

**Success Criteria Day 15-16:**
- ✅ JWT service with tenant-aware claims implemented
- ✅ Complete authentication use case with security measures
- ✅ Password hashing with scrypt and salt
- ✅ Email verification and password reset workflows
- ✅ Account locking after failed attempts

### **Day 17-18: Profile Management Use Cases**

#### **Task 3.3: Customer Profile Use Case**

```go
// internal/application/usecase/customer_profile_usecase.go
package usecase

import (
    "context"
    "fmt"
    "github.com/kirimku/smartseller-backend/internal/application/dto"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    domainErrors "github.com/kirimku/smartseller-backend/internal/domain/errors"
    "github.com/google/uuid"
)

type CustomerProfileUseCase interface {
    GetProfile(ctx context.Context, req *dto.GetCustomerProfileRequest) (*dto.CustomerProfileResponse, error)
    UpdateProfile(ctx context.Context, req *dto.UpdateCustomerProfileRequest) (*dto.CustomerProfileResponse, error)
    ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error
    UpdatePreferences(ctx context.Context, req *dto.UpdatePreferencesRequest) error
    GetCustomerStats(ctx context.Context, req *dto.GetCustomerStatsRequest) (*dto.CustomerStatsResponse, error)
}

type customerProfileUseCase struct {
    customerRepo repository.CustomerRepository
    authUseCase  CustomerAuthUseCase // For password operations
}

func NewCustomerProfileUseCase(
    customerRepo repository.CustomerRepository,
    authUseCase CustomerAuthUseCase,
) CustomerProfileUseCase {
    return &customerProfileUseCase{
        customerRepo: customerRepo,
        authUseCase:  authUseCase,
    }
}

func (uc *customerProfileUseCase) GetProfile(ctx context.Context, req *dto.GetCustomerProfileRequest) (*dto.CustomerProfileResponse, error) {
    // Get customer by ID
    customer, err := uc.customerRepo.GetByID(ctx, req.StorefrontID, req.CustomerID)
    if err != nil {
        return nil, err
    }
    
    // Get addresses
    addresses, err := uc.customerRepo.GetAddresses(ctx, req.CustomerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get customer addresses: %w", err)
    }
    
    return &dto.CustomerProfileResponse{
        Customer:  dto.ToCustomerDTO(customer),
        Addresses: dto.ToCustomerAddressDTOs(addresses),
    }, nil
}

func (uc *customerProfileUseCase) UpdateProfile(ctx context.Context, req *dto.UpdateCustomerProfileRequest) (*dto.CustomerProfileResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Get existing customer
    customer, err := uc.customerRepo.GetByID(ctx, req.StorefrontID, req.CustomerID)
    if err != nil {
        return nil, err
    }
    
    // Update fields
    customer.FirstName = req.FirstName
    customer.LastName = req.LastName
    customer.Phone = req.Phone
    customer.DateOfBirth = req.DateOfBirth
    customer.Gender = req.Gender
    customer.ProfilePicture = req.ProfilePicture
    
    // Update customer
    if err := uc.customerRepo.Update(ctx, customer); err != nil {
        return nil, fmt.Errorf("failed to update customer: %w", err)
    }
    
    return &dto.CustomerProfileResponse{
        Customer: dto.ToCustomerDTO(customer),
    }, nil
}

func (uc *customerProfileUseCase) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error {
    // Validate request
    if err := req.Validate(); err != nil {
        return err
    }
    
    // Get customer
    customer, err := uc.customerRepo.GetByID(ctx, req.StorefrontID, req.CustomerID)
    if err != nil {
        return err
    }
    
    // Verify current password
    if !uc.authUseCase.(*customerAuthUseCase).verifyPassword(req.CurrentPassword, customer.PasswordHash, customer.PasswordSalt) {
        return domainErrors.ErrInvalidCredentials
    }
    
    // Validate new password strength
    if err := uc.authUseCase.(*customerAuthUseCase).validatePasswordStrength(req.NewPassword); err != nil {
        return err
    }
    
    // Hash new password
    newHash, newSalt, err := uc.authUseCase.(*customerAuthUseCase).hashPassword(req.NewPassword)
    if err != nil {
        return fmt.Errorf("failed to hash new password: %w", err)
    }
    
    // Update password
    customer.PasswordHash = newHash
    customer.PasswordSalt = newSalt
    
    if err := uc.customerRepo.Update(ctx, customer); err != nil {
        return fmt.Errorf("failed to update password: %w", err)
    }
    
    // Clear all refresh tokens to force re-login on all devices
    if err := uc.customerRepo.ClearRefreshToken(ctx, customer.StorefrontID, customer.ID); err != nil {
        // Log error but don't fail the operation
        fmt.Printf("Failed to clear refresh tokens: %v\n", err)
    }
    
    return nil
}
```

#### **Task 3.4: Address Management Use Case**

```go
// internal/application/usecase/customer_address_usecase.go
package usecase

import (
    "context"
    "fmt"
    "github.com/kirimku/smartseller-backend/internal/application/dto"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    domainErrors "github.com/kirimku/smartseller-backend/internal/domain/errors"
    "github.com/google/uuid"
)

type CustomerAddressUseCase interface {
    GetAddresses(ctx context.Context, req *dto.GetCustomerAddressesRequest) (*dto.CustomerAddressesResponse, error)
    CreateAddress(ctx context.Context, req *dto.CreateCustomerAddressRequest) (*dto.CustomerAddressResponse, error)
    UpdateAddress(ctx context.Context, req *dto.UpdateCustomerAddressRequest) (*dto.CustomerAddressResponse, error)
    DeleteAddress(ctx context.Context, req *dto.DeleteCustomerAddressRequest) error
    SetDefaultAddress(ctx context.Context, req *dto.SetDefaultAddressRequest) error
}

type customerAddressUseCase struct {
    customerRepo repository.CustomerRepository
    addressRepo  repository.CustomerAddressRepository
}

func NewCustomerAddressUseCase(
    customerRepo repository.CustomerRepository,
    addressRepo repository.CustomerAddressRepository,
) CustomerAddressUseCase {
    return &customerAddressUseCase{
        customerRepo: customerRepo,
        addressRepo:  addressRepo,
    }
}

func (uc *customerAddressUseCase) CreateAddress(ctx context.Context, req *dto.CreateCustomerAddressRequest) (*dto.CustomerAddressResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Verify customer exists and belongs to storefront
    customer, err := uc.customerRepo.GetByID(ctx, req.StorefrontID, req.CustomerID)
    if err != nil {
        return nil, err
    }
    
    // Create address entity
    address := &entity.CustomerAddress{
        CustomerID:   customer.ID,
        Type:         entity.AddressType(req.Type),
        Label:        req.Label,
        FirstName:    req.FirstName,
        LastName:     req.LastName,
        Company:      req.Company,
        AddressLine1: req.AddressLine1,
        AddressLine2: req.AddressLine2,
        City:         req.City,
        Province:     req.Province,
        PostalCode:   req.PostalCode,
        Country:      req.Country,
        Phone:        req.Phone,
        IsDefault:    req.IsDefault,
        Latitude:     req.Latitude,
        Longitude:    req.Longitude,
    }
    
    // Create address
    if err := uc.addressRepo.Create(ctx, address); err != nil {
        return nil, fmt.Errorf("failed to create address: %w", err)
    }
    
    return &dto.CustomerAddressResponse{
        Address: dto.ToCustomerAddressDTO(address),
    }, nil
}

func (uc *customerAddressUseCase) UpdateAddress(ctx context.Context, req *dto.UpdateCustomerAddressRequest) (*dto.CustomerAddressResponse, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Get existing address
    address, err := uc.addressRepo.GetByID(ctx, req.AddressID)
    if err != nil {
        return nil, err
    }
    
    // Verify ownership through customer
    customer, err := uc.customerRepo.GetByID(ctx, req.StorefrontID, address.CustomerID)
    if err != nil {
        return nil, err
    }
    
    // Update address fields
    address.Type = entity.AddressType(req.Type)
    address.Label = req.Label
    address.FirstName = req.FirstName
    address.LastName = req.LastName
    address.Company = req.Company
    address.AddressLine1 = req.AddressLine1
    address.AddressLine2 = req.AddressLine2
    address.City = req.City
    address.Province = req.Province
    address.PostalCode = req.PostalCode
    address.Country = req.Country
    address.Phone = req.Phone
    address.Latitude = req.Latitude
    address.Longitude = req.Longitude
    
    // Update address
    if err := uc.addressRepo.Update(ctx, address); err != nil {
        return nil, fmt.Errorf("failed to update address: %w", err)
    }
    
    return &dto.CustomerAddressResponse{
        Address: dto.ToCustomerAddressDTO(address),
    }, nil
}

func (uc *customerAddressUseCase) DeleteAddress(ctx context.Context, req *dto.DeleteCustomerAddressRequest) error {
    // Get address
    address, err := uc.addressRepo.GetByID(ctx, req.AddressID)
    if err != nil {
        return err
    }
    
    // Verify ownership
    customer, err := uc.customerRepo.GetByID(ctx, req.StorefrontID, address.CustomerID)
    if err != nil {
        return err
    }
    
    // Don't allow deleting default address if there are other addresses
    if address.IsDefault {
        addresses, err := uc.addressRepo.GetByCustomerID(ctx, customer.ID)
        if err != nil {
            return fmt.Errorf("failed to check other addresses: %w", err)
        }
        
        if len(addresses) > 1 {
            return domainErrors.ErrCannotDeleteDefaultAddr
        }
    }
    
    // Delete address
    if err := uc.addressRepo.Delete(ctx, req.AddressID); err != nil {
        return fmt.Errorf("failed to delete address: %w", err)
    }
    
    return nil
}

func (uc *customerAddressUseCase) SetDefaultAddress(ctx context.Context, req *dto.SetDefaultAddressRequest) error {
    // Get address
    address, err := uc.addressRepo.GetByID(ctx, req.AddressID)
    if err != nil {
        return err
    }
    
    // Verify ownership
    customer, err := uc.customerRepo.GetByID(ctx, req.StorefrontID, address.CustomerID)
    if err != nil {
        return err
    }
    
    // Set as default
    if err := uc.addressRepo.SetAsDefault(ctx, customer.ID, req.AddressID); err != nil {
        return fmt.Errorf("failed to set default address: %w", err)
    }
    
    return nil
}
```

**Success Criteria Day 17-18:**
- ✅ Complete customer profile management use case
- ✅ Address management with full CRUD operations
- ✅ Password change with security validations
- ✅ Preference management system
- ✅ Multi-tenant data validation in all operations

### **Day 19-21: Business Logic & Validation**

#### **Task 3.5: Data Transfer Objects (DTOs)**

```go
// internal/application/dto/customer_auth_dto.go
package dto

import (
    "time"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/google/uuid"
)

// Registration DTOs
type CustomerRegisterRequest struct {
    StorefrontSlug string                     `json:"storefront_slug" validate:"required,min=2,max=100"`
    Email          string                     `json:"email" validate:"required,email"`
    Password       string                     `json:"password" validate:"required,min=8"`
    FirstName      string                     `json:"first_name" validate:"required,min=1,max=100"`
    LastName       string                     `json:"last_name" validate:"required,min=1,max=100"`
    Phone          *string                    `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
    DateOfBirth    *time.Time                 `json:"date_of_birth,omitempty"`
    Gender         *entity.Gender             `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
    Preferences    *CustomerPreferencesDTO    `json:"preferences,omitempty"`
    AcceptTerms    bool                       `json:"accept_terms" validate:"required,eq=true"`
    MarketingOptIn bool                       `json:"marketing_opt_in"`
}

type CustomerRegisterResponse struct {
    Customer                    *CustomerDTO `json:"customer"`
    AccessToken                 string       `json:"access_token"`
    RefreshToken               string       `json:"refresh_token"`
    ExpiresAt                  time.Time    `json:"expires_at"`
    RequiresEmailVerification  bool         `json:"requires_email_verification"`
}

// Login DTOs
type CustomerLoginRequest struct {
    StorefrontSlug string `json:"storefront_slug" validate:"required"`
    Email          string `json:"email" validate:"required,email"`
    Password       string `json:"password" validate:"required"`
    RememberMe     bool   `json:"remember_me"`
}

type CustomerLoginResponse struct {
    Customer     *CustomerDTO `json:"customer"`
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresAt    time.Time    `json:"expires_at"`
}

// Token DTOs
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
}

// Email verification DTOs
type VerifyEmailRequest struct {
    Token string `json:"token" validate:"required"`
}

type RequestPasswordResetRequest struct {
    StorefrontSlug string `json:"storefront_slug" validate:"required"`
    Email          string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
    Token       string `json:"token" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=8"`
}

// Logout DTO
type LogoutRequest struct {
    StorefrontID uuid.UUID `json:"storefront_id" validate:"required"`
    CustomerID   uuid.UUID `json:"customer_id" validate:"required"`
    RefreshToken string    `json:"refresh_token"`
}

// Customer DTO
type CustomerDTO struct {
    ID                  uuid.UUID               `json:"id"`
    Email               string                  `json:"email"`
    Phone               *string                 `json:"phone,omitempty"`
    FirstName           string                  `json:"first_name"`
    LastName            string                  `json:"last_name"`
    DateOfBirth         *time.Time              `json:"date_of_birth,omitempty"`
    Gender              *entity.Gender          `json:"gender,omitempty"`
    ProfilePicture      *string                 `json:"profile_picture,omitempty"`
    EmailVerified       bool                    `json:"email_verified"`
    PhoneVerified       bool                    `json:"phone_verified"`
    Status              entity.CustomerStatus   `json:"status"`
    LastLoginAt         *time.Time              `json:"last_login_at,omitempty"`
    TotalOrders         int                     `json:"total_orders"`
    TotalSpent          float64                 `json:"total_spent"`
    AverageOrderValue   float64                 `json:"average_order_value"`
    LifetimeValue       float64                 `json:"lifetime_value"`
    Preferences         CustomerPreferencesDTO  `json:"preferences"`
    Tags                []string               `json:"tags"`
    CreatedAt           time.Time              `json:"created_at"`
    UpdatedAt           time.Time              `json:"updated_at"`
}

type CustomerPreferencesDTO struct {
    Language             string `json:"language"`
    Currency             string `json:"currency"`
    EmailNotifications   bool   `json:"email_notifications"`
    SMSNotifications     bool   `json:"sms_notifications"`
    MarketingEmails      bool   `json:"marketing_emails"`
    OrderUpdates         bool   `json:"order_updates"`
    NewsletterSubscribed bool   `json:"newsletter_subscribed"`
}

// Validation methods
func (r *CustomerRegisterRequest) Validate() error {
    // Custom validation logic beyond struct tags
    return nil
}

func (r *CustomerLoginRequest) Validate() error {
    // Custom validation logic
    return nil
}

// Converter functions
func ToCustomerDTO(customer *entity.Customer) *CustomerDTO {
    if customer == nil {
        return nil
    }
    
    return &CustomerDTO{
        ID:                customer.ID,
        Email:             customer.Email,
        Phone:             customer.Phone,
        FirstName:         customer.FirstName,
        LastName:          customer.LastName,
        DateOfBirth:       customer.DateOfBirth,
        Gender:            customer.Gender,
        ProfilePicture:    customer.ProfilePicture,
        EmailVerified:     customer.IsEmailVerified(),
        PhoneVerified:     customer.IsPhoneVerified(),
        Status:            customer.Status,
        LastLoginAt:       customer.LastLoginAt,
        TotalOrders:       customer.TotalOrders,
        TotalSpent:        customer.TotalSpent,
        AverageOrderValue: customer.AverageOrderValue,
        LifetimeValue:     customer.LifetimeValue,
        Preferences:       ToCustomerPreferencesDTO(customer.Preferences),
        Tags:              customer.Tags,
        CreatedAt:         customer.CreatedAt,
        UpdatedAt:         customer.UpdatedAt,
    }
}

func ToCustomerPreferencesDTO(prefs entity.CustomerPreferences) CustomerPreferencesDTO {
    return CustomerPreferencesDTO{
        Language:             prefs.Language,
        Currency:             prefs.Currency,
        EmailNotifications:   prefs.EmailNotifications,
        SMSNotifications:     prefs.SMSNotifications,
        MarketingEmails:      prefs.MarketingEmails,
        OrderUpdates:         prefs.OrderUpdates,
        NewsletterSubscribed: prefs.NewsletterSubscribed,
    }
}

func ToCustomerPreferences(dto *CustomerPreferencesDTO) entity.CustomerPreferences {
    if dto == nil {
        return entity.CustomerPreferences{
            Language:             "en",
            Currency:             "IDR",
            EmailNotifications:   true,
            SMSNotifications:     false,
            MarketingEmails:      true,
            OrderUpdates:         true,
            NewsletterSubscribed: false,
        }
    }
    
    return entity.CustomerPreferences{
        Language:             dto.Language,
        Currency:             dto.Currency,
        EmailNotifications:   dto.EmailNotifications,
        SMSNotifications:     dto.SMSNotifications,
        MarketingEmails:      dto.MarketingEmails,
        OrderUpdates:         dto.OrderUpdates,
        NewsletterSubscribed: dto.NewsletterSubscribed,
    }
}
```

#### **Task 3.6: Business Logic Validation & Rules**

```go
// internal/application/service/customer_validation_service.go
package service

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "time"
    "github.com/kirimku/smartseller-backend/internal/domain/entity"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    domainErrors "github.com/kirimku/smartseller-backend/internal/domain/errors"
)

type CustomerValidationService interface {
    ValidateRegistration(ctx context.Context, customer *entity.Customer, storefront *entity.Storefront) error
    ValidateProfileUpdate(ctx context.Context, customer *entity.Customer, updates *entity.Customer) error
    ValidateAddressCreation(ctx context.Context, address *entity.CustomerAddress, customer *entity.Customer) error
    ValidateBusinessRules(ctx context.Context, customer *entity.Customer, storefront *entity.Storefront) error
}

type customerValidationService struct {
    customerRepo   repository.CustomerRepository
    storefrontRepo repository.StorefrontRepository
}

func NewCustomerValidationService(
    customerRepo repository.CustomerRepository,
    storefrontRepo repository.StorefrontRepository,
) CustomerValidationService {
    return &customerValidationService{
        customerRepo:   customerRepo,
        storefrontRepo: storefrontRepo,
    }
}

func (s *customerValidationService) ValidateRegistration(ctx context.Context, customer *entity.Customer, storefront *entity.Storefront) error {
    // Basic entity validation
    if err := customer.Validate(); err != nil {
        return err
    }
    
    // Business rule validations
    if err := s.ValidateBusinessRules(ctx, customer, storefront); err != nil {
        return err
    }
    
    // Check email uniqueness within storefront
    existing, err := s.customerRepo.GetByEmail(ctx, customer.StorefrontID, customer.Email)
    if err == nil && existing != nil {
        return domainErrors.ErrEmailAlreadyExists
    }
    if err != nil && !domainErrors.IsCustomerNotFound(err) {
        return fmt.Errorf("failed to check email uniqueness: %w", err)
    }
    
    // Check phone uniqueness if provided
    if customer.Phone != nil {
        existing, err := s.customerRepo.GetByPhone(ctx, customer.StorefrontID, *customer.Phone)
        if err == nil && existing != nil {
            return domainErrors.ErrPhoneAlreadyExists
        }
        if err != nil && !domainErrors.IsCustomerNotFound(err) {
            return fmt.Errorf("failed to check phone uniqueness: %w", err)
        }
    }
    
    return nil
}

func (s *customerValidationService) ValidateBusinessRules(ctx context.Context, customer *entity.Customer, storefront *entity.Storefront) error {
    // Age validation
    if customer.DateOfBirth != nil {
        age := time.Now().Year() - customer.DateOfBirth.Year()
        if age < 13 {
            return domainErrors.NewDomainError("CUSTOMER_TOO_YOUNG", "Customer must be at least 13 years old")
        }
        if age > 120 {
            return domainErrors.NewDomainError("INVALID_DATE_OF_BIRTH", "Invalid date of birth")
        }
    }
    
    // Email domain validation for specific storefronts
    if err := s.validateEmailDomain(customer.Email, storefront); err != nil {
        return err
    }
    
    // Phone number format validation
    if customer.Phone != nil {
        if err := s.validatePhoneNumber(*customer.Phone, storefront); err != nil {
            return err
        }
    }
    
    // Name validation
    if err := s.validateName(customer.FirstName, customer.LastName); err != nil {
        return err
    }
    
    return nil
}

func (s *customerValidationService) validateEmailDomain(email string, storefront *entity.Storefront) error {
    // Extract domain from email
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return domainErrors.NewDomainError("INVALID_EMAIL_FORMAT", "Invalid email format")
    }
    
    domain := strings.ToLower(parts[1])
    
    // Check for banned domains (implement business logic)
    bannedDomains := []string{"tempmail.com", "10minutemail.com", "guerrillamail.com"}
    for _, banned := range bannedDomains {
        if domain == banned {
            return domainErrors.NewDomainError("BANNED_EMAIL_DOMAIN", "Email domain is not allowed")
        }
    }
    
    // Business rule: Some storefronts might require specific email domains
    // This would be configurable per storefront
    if storefront.Settings.RequireVerifiedEmailDomain {
        allowedDomains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com"}
        allowed := false
        for _, allowedDomain := range allowedDomains {
            if domain == allowedDomain {
                allowed = true
                break
            }
        }
        if !allowed {
            return domainErrors.NewDomainError("EMAIL_DOMAIN_NOT_ALLOWED", "Please use a verified email provider")
        }
    }
    
    return nil
}

func (s *customerValidationService) validatePhoneNumber(phone string, storefront *entity.Storefront) error {
    // Indonesian phone number validation (example)
    if storefront.Currency == "IDR" {
        // Indonesian phone should start with +62 or 08
        phoneRegex := regexp.MustCompile(`^(\+62|62|08)[0-9]{8,11}$`)
        if !phoneRegex.MatchString(phone) {
            return domainErrors.NewDomainError("INVALID_PHONE_FORMAT", "Invalid Indonesian phone number format")
        }
    } else {
        // International phone validation
        phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
        if !phoneRegex.MatchString(phone) {
            return domainErrors.NewDomainError("INVALID_PHONE_FORMAT", "Invalid international phone number format")
        }
    }
    
    return nil
}

func (s *customerValidationService) validateName(firstName, lastName string) error {
    // Name should not contain numbers or special characters
    nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-'\.]+$`)
    
    if !nameRegex.MatchString(firstName) {
        return domainErrors.NewDomainError("INVALID_FIRST_NAME", "First name contains invalid characters")
    }
    
    if !nameRegex.MatchString(lastName) {
        return domainErrors.NewDomainError("INVALID_LAST_NAME", "Last name contains invalid characters")
    }
    
    // Check for suspicious patterns
    suspiciousPatterns := []string{"test", "admin", "null", "undefined", "fake"}
    firstNameLower := strings.ToLower(firstName)
    lastNameLower := strings.ToLower(lastName)
    
    for _, pattern := range suspiciousPatterns {
        if strings.Contains(firstNameLower, pattern) || strings.Contains(lastNameLower, pattern) {
            return domainErrors.NewDomainError("SUSPICIOUS_NAME", "Name appears to be fake or test data")
        }
    }
    
    return nil
}

func (s *customerValidationService) ValidateAddressCreation(ctx context.Context, address *entity.CustomerAddress, customer *entity.Customer) error {
    // Basic entity validation
    if err := address.Validate(); err != nil {
        return err
    }
    
    // Business rules for address
    if err := s.validateAddressBusinessRules(address, customer); err != nil {
        return err
    }
    
    return nil
}

func (s *customerValidationService) validateAddressBusinessRules(address *entity.CustomerAddress, customer *entity.Customer) error {
    // Postal code format validation (example for Indonesia)
    if address.Country == "Indonesia" {
        postalRegex := regexp.MustCompile(`^\d{5}$`)
        if !postalRegex.MatchString(address.PostalCode) {
            return domainErrors.NewDomainError("INVALID_POSTAL_CODE", "Invalid Indonesian postal code format")
        }
    }
    
    // Address line validation
    if len(address.AddressLine1) < 10 {
        return domainErrors.NewDomainError("ADDRESS_TOO_SHORT", "Address line 1 must be at least 10 characters")
    }
    
    // City validation - no numbers allowed
    cityRegex := regexp.MustCompile(`^[a-zA-Z\s\-'\.]+$`)
    if !cityRegex.MatchString(address.City) {
        return domainErrors.NewDomainError("INVALID_CITY_NAME", "City name contains invalid characters")
    }
    
    return nil
}
```

**Success Criteria Day 19-21:**
- ✅ Complete DTO system with validation
- ✅ Business logic validation service
- ✅ Email domain and phone number validation
- ✅ Address validation with country-specific rules
- ✅ Comprehensive error handling with domain errors

---

## 🎯 **Phase 3 Success Criteria**

At the end of Phase 3, you should have:

### **✅ Authentication System**
- JWT service with tenant-aware claims
- Complete registration and login flows
- Email verification and password reset
- Account security with lockout protection

### **✅ Business Logic Layer**
- Customer profile management use cases
- Address management with CRUD operations
- Data validation with business rules
- Multi-tenant security throughout

### **✅ DTO & Validation**
- Complete DTO system for all operations
- Request/response validation
- Business rule validation service
- Comprehensive error handling

### **Ready for Phase 4**
- API handlers and middleware
- HTTP endpoints with proper routing
- Authentication middleware
- Integration with existing seller system

---

## 🌐 **Phase 4: API Layer & Integration (7 days)**

### **Phase Goals**
- Implement HTTP handlers for customer endpoints
- Create authentication and authorization middleware
- Set up storefront-specific routing patterns
- Integrate with existing seller dashboard
- Add API documentation and testing endpoints

### **Day 22-23: API Handlers & Routing**

#### **Task 4.1: Customer Authentication Handlers**

```go
// internal/interfaces/api/handler/customer_auth_handler.go
package handler

import (
    "encoding/json"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/kirimku/smartseller-backend/internal/application/dto"
    "github.com/kirimku/smartseller-backend/internal/application/usecase"
    "github.com/kirimku/smartseller-backend/pkg/utils"
)

type CustomerAuthHandler struct {
    authUseCase usecase.CustomerAuthUseCase
}

func NewCustomerAuthHandler(authUseCase usecase.CustomerAuthUseCase) *CustomerAuthHandler {
    return &CustomerAuthHandler{
        authUseCase: authUseCase,
    }
}

// POST /api/storefront/{slug}/auth/register
func (h *CustomerAuthHandler) Register(c *gin.Context) {
    var req dto.CustomerRegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    // Get storefront slug from URL
    req.StorefrontSlug = c.Param("slug")
    
    // Validate request
    if err := req.Validate(); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", err)
        return
    }
    
    response, err := h.authUseCase.Register(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusCreated, "Registration successful", response)
}

// POST /api/storefront/{slug}/auth/login
func (h *CustomerAuthHandler) Login(c *gin.Context) {
    var req dto.CustomerLoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    // Get storefront slug from URL
    req.StorefrontSlug = c.Param("slug")
    
    response, err := h.authUseCase.Login(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// POST /api/storefront/{slug}/auth/refresh
func (h *CustomerAuthHandler) RefreshToken(c *gin.Context) {
    var req dto.RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    response, err := h.authUseCase.RefreshToken(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Token refreshed", response)
}

// POST /api/storefront/{slug}/auth/verify-email
func (h *CustomerAuthHandler) VerifyEmail(c *gin.Context) {
    var req dto.VerifyEmailRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    err := h.authUseCase.VerifyEmail(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Email verified successfully", nil)
}

// POST /api/storefront/{slug}/auth/forgot-password
func (h *CustomerAuthHandler) ForgotPassword(c *gin.Context) {
    var req dto.RequestPasswordResetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    req.StorefrontSlug = c.Param("slug")
    
    err := h.authUseCase.RequestPasswordReset(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Password reset email sent", nil)
}

// POST /api/storefront/{slug}/auth/reset-password
func (h *CustomerAuthHandler) ResetPassword(c *gin.Context) {
    var req dto.ResetPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    err := h.authUseCase.ResetPassword(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Password reset successful", nil)
}

// POST /api/storefront/{slug}/auth/logout
func (h *CustomerAuthHandler) Logout(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    req := &dto.LogoutRequest{
        StorefrontID: customerCtx.StorefrontID,
        CustomerID:   customerCtx.CustomerID,
    }
    
    // Get refresh token from request body if provided
    var body struct {
        RefreshToken string `json:"refresh_token"`
    }
    c.ShouldBindJSON(&body)
    req.RefreshToken = body.RefreshToken
    
    err := h.authUseCase.Logout(c.Request.Context(), req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Logged out successfully", nil)
}
```

#### **Task 4.2: Customer Profile Handlers**

```go
// internal/interfaces/api/handler/customer_profile_handler.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/kirimku/smartseller-backend/internal/application/dto"
    "github.com/kirimku/smartseller-backend/internal/application/usecase"
    "github.com/kirimku/smartseller-backend/pkg/utils"
    "github.com/google/uuid"
)

type CustomerProfileHandler struct {
    profileUseCase usecase.CustomerProfileUseCase
    addressUseCase usecase.CustomerAddressUseCase
}

func NewCustomerProfileHandler(
    profileUseCase usecase.CustomerProfileUseCase,
    addressUseCase usecase.CustomerAddressUseCase,
) *CustomerProfileHandler {
    return &CustomerProfileHandler{
        profileUseCase: profileUseCase,
        addressUseCase: addressUseCase,
    }
}

// GET /api/storefront/{slug}/profile
func (h *CustomerProfileHandler) GetProfile(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    req := &dto.GetCustomerProfileRequest{
        StorefrontID: customerCtx.StorefrontID,
        CustomerID:   customerCtx.CustomerID,
    }
    
    response, err := h.profileUseCase.GetProfile(c.Request.Context(), req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Profile retrieved", response)
}

// PUT /api/storefront/{slug}/profile
func (h *CustomerProfileHandler) UpdateProfile(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    var req dto.UpdateCustomerProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    req.StorefrontID = customerCtx.StorefrontID
    req.CustomerID = customerCtx.CustomerID
    
    response, err := h.profileUseCase.UpdateProfile(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Profile updated", response)
}

// PUT /api/storefront/{slug}/profile/password
func (h *CustomerProfileHandler) ChangePassword(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    var req dto.ChangePasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    req.StorefrontID = customerCtx.StorefrontID
    req.CustomerID = customerCtx.CustomerID
    
    err := h.profileUseCase.ChangePassword(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

// Address endpoints
// GET /api/storefront/{slug}/profile/addresses
func (h *CustomerProfileHandler) GetAddresses(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    req := &dto.GetCustomerAddressesRequest{
        StorefrontID: customerCtx.StorefrontID,
        CustomerID:   customerCtx.CustomerID,
    }
    
    response, err := h.addressUseCase.GetAddresses(c.Request.Context(), req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Addresses retrieved", response)
}

// POST /api/storefront/{slug}/profile/addresses
func (h *CustomerProfileHandler) CreateAddress(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    var req dto.CreateCustomerAddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    req.StorefrontID = customerCtx.StorefrontID
    req.CustomerID = customerCtx.CustomerID
    
    response, err := h.addressUseCase.CreateAddress(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusCreated, "Address created", response)
}

// PUT /api/storefront/{slug}/profile/addresses/{addressId}
func (h *CustomerProfileHandler) UpdateAddress(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    addressID, err := uuid.Parse(c.Param("addressId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_ADDRESS_ID", "Invalid address ID", err)
        return
    }
    
    var req dto.UpdateCustomerAddressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    req.StorefrontID = customerCtx.StorefrontID
    req.CustomerID = customerCtx.CustomerID
    req.AddressID = addressID
    
    response, err := h.addressUseCase.UpdateAddress(c.Request.Context(), &req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Address updated", response)
}

// DELETE /api/storefront/{slug}/profile/addresses/{addressId}
func (h *CustomerProfileHandler) DeleteAddress(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    addressID, err := uuid.Parse(c.Param("addressId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_ADDRESS_ID", "Invalid address ID", err)
        return
    }
    
    req := &dto.DeleteCustomerAddressRequest{
        StorefrontID: customerCtx.StorefrontID,
        CustomerID:   customerCtx.CustomerID,
        AddressID:    addressID,
    }
    
    err = h.addressUseCase.DeleteAddress(c.Request.Context(), req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Address deleted", nil)
}

// PUT /api/storefront/{slug}/profile/addresses/{addressId}/default
func (h *CustomerProfileHandler) SetDefaultAddress(c *gin.Context) {
    customerCtx := utils.GetCustomerFromContext(c)
    if customerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Customer not authenticated", nil)
        return
    }
    
    addressID, err := uuid.Parse(c.Param("addressId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_ADDRESS_ID", "Invalid address ID", err)
        return
    }
    
    req := &dto.SetDefaultAddressRequest{
        StorefrontID: customerCtx.StorefrontID,
        CustomerID:   customerCtx.CustomerID,
        AddressID:    addressID,
    }
    
    err = h.addressUseCase.SetDefaultAddress(c.Request.Context(), req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Default address set", nil)
}
```

### **Day 24-25: Middleware & Security**

#### **Task 4.3: Customer Authentication Middleware**

```go
// pkg/middleware/customer_auth_middleware.go
package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/kirimku/smartseller-backend/internal/application/service"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
    "github.com/kirimku/smartseller-backend/pkg/utils"
    "github.com/google/uuid"
)

type CustomerContext struct {
    CustomerID   uuid.UUID
    StorefrontID uuid.UUID
    Email        string
    FirstName    string
    LastName     string
    Verified     bool
}

type CustomerAuthMiddleware struct {
    jwtService      service.JWTService
    tenantResolver  tenant.TenantResolver
}

func NewCustomerAuthMiddleware(
    jwtService service.JWTService,
    tenantResolver tenant.TenantResolver,
) *CustomerAuthMiddleware {
    return &CustomerAuthMiddleware{
        jwtService:     jwtService,
        tenantResolver: tenantResolver,
    }
}

// CustomerAuth middleware for protecting customer endpoints
func (m *CustomerAuthMiddleware) CustomerAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            utils.ErrorResponse(c, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization token required", nil)
            c.Abort()
            return
        }
        
        // Check for Bearer token format
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Token must be in Bearer format", nil)
            c.Abort()
            return
        }
        
        token := tokenParts[1]
        
        // Validate token
        claims, err := m.jwtService.ValidateToken(c.Request.Context(), token, "access")
        if err != nil {
            utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", err)
            c.Abort()
            return
        }
        
        // Parse UUIDs
        customerID, err := uuid.Parse(claims.CustomerID)
        if err != nil {
            utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_CUSTOMER_ID", "Invalid customer ID in token", err)
            c.Abort()
            return
        }
        
        storefrontID, err := uuid.Parse(claims.StorefrontID)
        if err != nil {
            utils.ErrorResponse(c, http.StatusUnauthorized, "INVALID_STOREFRONT_ID", "Invalid storefront ID in token", err)
            c.Abort()
            return
        }
        
        // Verify storefront slug matches token (tenant isolation)
        storefrontSlug := c.Param("slug")
        if storefrontSlug != "" {
            storefront, err := m.tenantResolver.GetStorefrontBySlug(c.Request.Context(), storefrontSlug)
            if err != nil {
                utils.ErrorResponse(c, http.StatusNotFound, "STOREFRONT_NOT_FOUND", "Storefront not found", err)
                c.Abort()
                return
            }
            
            if storefront.ID != storefrontID {
                utils.ErrorResponse(c, http.StatusForbidden, "TENANT_MISMATCH", "Token does not match storefront", nil)
                c.Abort()
                return
            }
        }
        
        // Create customer context
        customerCtx := &CustomerContext{
            CustomerID:   customerID,
            StorefrontID: storefrontID,
            Email:        claims.Email,
            FirstName:    claims.FirstName,
            LastName:     claims.LastName,
            Verified:     claims.Verified,
        }
        
        // Store in context
        c.Set("customer", customerCtx)
        
        c.Next()
    }
}

// OptionalCustomerAuth middleware for endpoints that work with or without auth
func (m *CustomerAuthMiddleware) OptionalCustomerAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Next()
            return
        }
        
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            c.Next()
            return
        }
        
        token := tokenParts[1]
        claims, err := m.jwtService.ValidateToken(c.Request.Context(), token, "access")
        if err != nil {
            c.Next()
            return
        }
        
        customerID, err := uuid.Parse(claims.CustomerID)
        if err != nil {
            c.Next()
            return
        }
        
        storefrontID, err := uuid.Parse(claims.StorefrontID)
        if err != nil {
            c.Next()
            return
        }
        
        customerCtx := &CustomerContext{
            CustomerID:   customerID,
            StorefrontID: storefrontID,
            Email:        claims.Email,
            FirstName:    claims.FirstName,
            LastName:     claims.LastName,
            Verified:     claims.Verified,
        }
        
        c.Set("customer", customerCtx)
        c.Next()
    }
}

// StorefrontResolver middleware to resolve storefront from slug
func (m *CustomerAuthMiddleware) StorefrontResolver() gin.HandlerFunc {
    return func(c *gin.Context) {
        slug := c.Param("slug")
        if slug == "" {
            utils.ErrorResponse(c, http.StatusBadRequest, "MISSING_STOREFRONT_SLUG", "Storefront slug is required", nil)
            c.Abort()
            return
        }
        
        storefront, err := m.tenantResolver.GetStorefrontBySlug(c.Request.Context(), slug)
        if err != nil {
            utils.ErrorResponse(c, http.StatusNotFound, "STOREFRONT_NOT_FOUND", "Storefront not found", err)
            c.Abort()
            return
        }
        
        // Check if storefront is active
        if storefront.Status != "active" {
            utils.ErrorResponse(c, http.StatusServiceUnavailable, "STOREFRONT_INACTIVE", "Storefront is not active", nil)
            c.Abort()
            return
        }
        
        // Store storefront in context
        c.Set("storefront", storefront)
        
        c.Next()
    }
}
```

#### **Task 4.4: Rate Limiting & Security Headers**

```go
// pkg/middleware/rate_limit_middleware.go
package middleware

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/kirimku/smartseller-backend/pkg/utils"
)

// Rate limiter for customer authentication endpoints
type RateLimiter struct {
    requests map[string]*RequestData
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}

type RequestData struct {
    count     int
    resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        requests: make(map[string]*RequestData),
        limit:    limit,
        window:   window,
    }
    
    // Clean up expired entries every minute
    go rl.cleanup()
    
    return rl
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    now := time.Now()
    data, exists := rl.requests[key]
    
    if !exists || now.After(data.resetTime) {
        rl.requests[key] = &RequestData{
            count:     1,
            resetTime: now.Add(rl.window),
        }
        return true
    }
    
    if data.count >= rl.limit {
        return false
    }
    
    data.count++
    return true
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        rl.mutex.Lock()
        now := time.Now()
        for key, data := range rl.requests {
            if now.After(data.resetTime) {
                delete(rl.requests, key)
            }
        }
        rl.mutex.Unlock()
    }
}

// Rate limit middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Use IP address as key, but could be enhanced with user ID for authenticated requests
        key := c.ClientIP()
        
        if !rl.Allow(key) {
            utils.ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", nil)
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Security headers middleware
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Content-Security-Policy", "default-src 'self'")
        
        c.Next()
    }
}

// CORS middleware specifically for storefront API
func StorefrontCORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        // In production, you would validate against allowed storefront domains
        c.Header("Access-Control-Allow-Origin", origin)
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
        c.Header("Access-Control-Max-Age", "86400")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        
        c.Next()
    }
}
```

### **Day 26-27: Seller Dashboard Integration**

#### **Task 4.5: Seller Customer Management Handler**

```go
// internal/interfaces/api/handler/seller_customer_handler.go
package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/kirimku/smartseller-backend/internal/application/dto"
    "github.com/kirimku/smartseller-backend/internal/domain/repository"
    "github.com/kirimku/smartseller-backend/pkg/utils"
    "github.com/google/uuid"
)

type SellerCustomerHandler struct {
    customerRepo   repository.CustomerRepository
    storefrontRepo repository.StorefrontRepository
}

func NewSellerCustomerHandler(
    customerRepo repository.CustomerRepository,
    storefrontRepo repository.StorefrontRepository,
) *SellerCustomerHandler {
    return &SellerCustomerHandler{
        customerRepo:   customerRepo,
        storefrontRepo: storefrontRepo,
    }
}

// GET /api/v1/seller/storefronts/{storefrontId}/customers
func (h *SellerCustomerHandler) GetCustomers(c *gin.Context) {
    // Get authenticated seller from context (existing middleware)
    sellerCtx := utils.GetUserFromContext(c) // Existing seller auth
    if sellerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Seller not authenticated", nil)
        return
    }
    
    // Parse storefront ID
    storefrontID, err := uuid.Parse(c.Param("storefrontId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_STOREFRONT_ID", "Invalid storefront ID", err)
        return
    }
    
    // Verify seller owns the storefront
    storefront, err := h.storefrontRepo.GetByID(c.Request.Context(), storefrontID)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    if storefront.SellerID != sellerCtx.UserID {
        utils.ErrorResponse(c, http.StatusForbidden, "ACCESS_DENIED", "Not authorized to access this storefront", nil)
        return
    }
    
    // Parse query parameters
    page := 1
    if pageStr := c.Query("page"); pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = p
        }
    }
    
    pageSize := 20
    if sizeStr := c.Query("page_size"); sizeStr != "" {
        if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
            pageSize = s
        }
    }
    
    req := &repository.GetCustomersRequest{
        StorefrontID: storefrontID,
        Page:         page,
        PageSize:     pageSize,
        Search:       c.Query("search"),
        OrderBy:      c.Query("order_by"),
        SortDesc:     c.Query("sort") == "desc",
    }
    
    // Parse status filter
    if statusStr := c.Query("status"); statusStr != "" {
        status := entity.CustomerStatus(statusStr)
        if status.IsValid() {
            req.Status = &status
        }
    }
    
    response, err := h.customerRepo.GetByStorefront(c.Request.Context(), req)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Customers retrieved", response)
}

// GET /api/v1/seller/storefronts/{storefrontId}/customers/{customerId}
func (h *SellerCustomerHandler) GetCustomer(c *gin.Context) {
    sellerCtx := utils.GetUserFromContext(c)
    if sellerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Seller not authenticated", nil)
        return
    }
    
    storefrontID, err := uuid.Parse(c.Param("storefrontId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_STOREFRONT_ID", "Invalid storefront ID", err)
        return
    }
    
    customerID, err := uuid.Parse(c.Param("customerId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID", err)
        return
    }
    
    // Verify seller owns the storefront
    storefront, err := h.storefrontRepo.GetByID(c.Request.Context(), storefrontID)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    if storefront.SellerID != sellerCtx.UserID {
        utils.ErrorResponse(c, http.StatusForbidden, "ACCESS_DENIED", "Not authorized to access this storefront", nil)
        return
    }
    
    // Get customer
    customer, err := h.customerRepo.GetByID(c.Request.Context(), storefrontID, customerID)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Customer retrieved", dto.ToCustomerDTO(customer))
}

// GET /api/v1/seller/storefronts/{storefrontId}/customers/stats
func (h *SellerCustomerHandler) GetCustomerStats(c *gin.Context) {
    sellerCtx := utils.GetUserFromContext(c)
    if sellerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Seller not authenticated", nil)
        return
    }
    
    storefrontID, err := uuid.Parse(c.Param("storefrontId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_STOREFRONT_ID", "Invalid storefront ID", err)
        return
    }
    
    // Verify seller owns the storefront
    storefront, err := h.storefrontRepo.GetByID(c.Request.Context(), storefrontID)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    if storefront.SellerID != sellerCtx.UserID {
        utils.ErrorResponse(c, http.StatusForbidden, "ACCESS_DENIED", "Not authorized to access this storefront", nil)
        return
    }
    
    // Get customer stats
    stats, err := h.customerRepo.GetCustomerStats(c.Request.Context(), storefrontID)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Customer stats retrieved", stats)
}

// PUT /api/v1/seller/storefronts/{storefrontId}/customers/{customerId}/notes
func (h *SellerCustomerHandler) UpdateCustomerNotes(c *gin.Context) {
    sellerCtx := utils.GetUserFromContext(c)
    if sellerCtx == nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Seller not authenticated", nil)
        return
    }
    
    storefrontID, err := uuid.Parse(c.Param("storefrontId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_STOREFRONT_ID", "Invalid storefront ID", err)
        return
    }
    
    customerID, err := uuid.Parse(c.Param("customerId"))
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID", err)
        return
    }
    
    var req struct {
        Notes string   `json:"notes"`
        Tags  []string `json:"tags"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
        return
    }
    
    // Verify seller owns the storefront
    storefront, err := h.storefrontRepo.GetByID(c.Request.Context(), storefrontID)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    if storefront.SellerID != sellerCtx.UserID {
        utils.ErrorResponse(c, http.StatusForbidden, "ACCESS_DENIED", "Not authorized to access this storefront", nil)
        return
    }
    
    // Get customer
    customer, err := h.customerRepo.GetByID(c.Request.Context(), storefrontID, customerID)
    if err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    // Update notes and tags
    customer.Notes = &req.Notes
    customer.Tags = req.Tags
    
    if err := h.customerRepo.Update(c.Request.Context(), customer); err != nil {
        utils.HandleDomainError(c, err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Customer notes updated", dto.ToCustomerDTO(customer))
}
```

### **Day 28: Route Setup & Testing**

#### **Task 4.6: Complete Route Setup**

```go
// internal/interfaces/api/routes/customer_routes.go
package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
    "github.com/kirimku/smartseller-backend/pkg/middleware"
)

func SetupCustomerRoutes(
    router *gin.Engine,
    authHandler *handler.CustomerAuthHandler,
    profileHandler *handler.CustomerProfileHandler,
    sellerCustomerHandler *handler.SellerCustomerHandler,
    customerAuthMiddleware *middleware.CustomerAuthMiddleware,
    rateLimiter *middleware.RateLimiter,
) {
    // Storefront-specific customer API routes
    storefrontAPI := router.Group("/api/storefront/:slug")
    {
        storefrontAPI.Use(middleware.StorefrontCORS())
        storefrontAPI.Use(middleware.SecurityHeaders())
        storefrontAPI.Use(customerAuthMiddleware.StorefrontResolver())
        
        // Authentication endpoints (no auth required)
        auth := storefrontAPI.Group("/auth")
        {
            // Apply rate limiting to auth endpoints
            auth.Use(rateLimiter.Middleware())
            
            auth.POST("/register", authHandler.Register)
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.RefreshToken)
            auth.POST("/verify-email", authHandler.VerifyEmail)
            auth.POST("/forgot-password", authHandler.ForgotPassword)
            auth.POST("/reset-password", authHandler.ResetPassword)
        }
        
        // Protected customer endpoints
        protected := storefrontAPI.Group("/")
        {
            protected.Use(customerAuthMiddleware.CustomerAuth())
            
            // Profile management
            profile := protected.Group("/profile")
            {
                profile.GET("", profileHandler.GetProfile)
                profile.PUT("", profileHandler.UpdateProfile)
                profile.PUT("/password", profileHandler.ChangePassword)
                
                // Address management
                addresses := profile.Group("/addresses")
                {
                    addresses.GET("", profileHandler.GetAddresses)
                    addresses.POST("", profileHandler.CreateAddress)
                    addresses.PUT("/:addressId", profileHandler.UpdateAddress)
                    addresses.DELETE("/:addressId", profileHandler.DeleteAddress)
                    addresses.PUT("/:addressId/default", profileHandler.SetDefaultAddress)
                }
            }
            
            // Logout (requires auth)
            protected.POST("/auth/logout", authHandler.Logout)
        }
    }
    
    // Seller dashboard API routes (existing auth middleware)
    sellerAPI := router.Group("/api/v1/seller")
    {
        sellerAPI.Use(middleware.AuthMiddleware()) // Existing seller auth
        
        storefronts := sellerAPI.Group("/storefronts/:storefrontId")
        {
            customers := storefronts.Group("/customers")
            {
                customers.GET("", sellerCustomerHandler.GetCustomers)
                customers.GET("/stats", sellerCustomerHandler.GetCustomerStats)
                customers.GET("/:customerId", sellerCustomerHandler.GetCustomer)
                customers.PUT("/:customerId/notes", sellerCustomerHandler.UpdateCustomerNotes)
            }
        }
    }
}
```

#### **Task 4.7: Utility Functions Enhancement**

```go
// pkg/utils/response_helpers.go
package utils

import (
    "net/http"
    "github.com/gin-gonic/gin"
    domainErrors "github.com/kirimku/smartseller-backend/internal/domain/errors"
    "github.com/kirimku/smartseller-backend/pkg/middleware"
)

// Get customer context from Gin context
func GetCustomerFromContext(c *gin.Context) *middleware.CustomerContext {
    if customer, exists := c.Get("customer"); exists {
        if customerCtx, ok := customer.(*middleware.CustomerContext); ok {
            return customerCtx
        }
    }
    return nil
}

// Enhanced domain error handler for customer-specific errors
func HandleDomainError(c *gin.Context, err error) {
    if domainErr, ok := err.(*domainErrors.DomainError); ok {
        switch domainErr.Code {
        case "CUSTOMER_NOT_FOUND":
            ErrorResponse(c, http.StatusNotFound, domainErr.Code, domainErr.Message, err)
        case "EMAIL_ALREADY_EXISTS", "PHONE_ALREADY_EXISTS":
            ErrorResponse(c, http.StatusConflict, domainErr.Code, domainErr.Message, err)
        case "INVALID_CREDENTIALS", "INVALID_TOKEN", "INVALID_EMAIL_TOKEN", "INVALID_PASSWORD_TOKEN":
            ErrorResponse(c, http.StatusUnauthorized, domainErr.Code, domainErr.Message, err)
        case "ACCOUNT_SUSPENDED", "ACCOUNT_LOCKED", "EMAIL_NOT_VERIFIED":
            ErrorResponse(c, http.StatusForbidden, domainErr.Code, domainErr.Message, err)
        case "WEAK_PASSWORD":
            ErrorResponse(c, http.StatusBadRequest, domainErr.Code, domainErr.Message, err)
        case "STOREFRONT_NOT_FOUND":
            ErrorResponse(c, http.StatusNotFound, domainErr.Code, domainErr.Message, err)
        case "STOREFRONT_INACTIVE":
            ErrorResponse(c, http.StatusServiceUnavailable, domainErr.Code, domainErr.Message, err)
        case "TENANT_MISMATCH", "TENANT_ACCESS_DENIED":
            ErrorResponse(c, http.StatusForbidden, domainErr.Code, domainErr.Message, err)
        default:
            ErrorResponse(c, http.StatusBadRequest, domainErr.Code, domainErr.Message, err)
        }
        return
    }
    
    // Fallback for non-domain errors
    ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred", err)
}
```

**Success Criteria Day 22-28:**
- ✅ Complete API handlers for authentication and profile management
- ✅ Multi-tenant authentication middleware with JWT validation
- ✅ Rate limiting and security headers for customer endpoints
- ✅ Seller dashboard integration for customer management
- ✅ Proper route setup with tenant isolation
- ✅ Enhanced error handling for customer-specific errors

---

## 🎯 **Phase 4 Success Criteria**

At the end of Phase 4, you should have:

### **✅ API Layer**
- Complete HTTP handlers for all customer operations
- RESTful endpoints with proper HTTP status codes
- Request validation and response formatting
- Error handling with domain-specific responses

### **✅ Security & Middleware**
- JWT-based authentication with tenant scoping
- Rate limiting for authentication endpoints
- CORS configuration for storefront domains
- Security headers and input validation

### **✅ Integration**
- Seller dashboard endpoints for customer management
- Proper authorization checks across all endpoints
- Multi-tenant data isolation in API layer
- Enhanced utility functions for context handling

### **Ready for Testing & Deployment**
- Complete API documentation
- Performance monitoring setup
- Deployment procedures
- Comprehensive testing strategy

---

## 🧪 **Testing Strategy**

### **Unit Tests**
- Repository layer with mock databases
- Use case validation and business logic
- JWT service token generation and validation
- Entity validation and business rules

### **Integration Tests**  
- API endpoint testing with real database
- Multi-tenant isolation verification
- Authentication flow end-to-end
- Error handling and edge cases

### **Performance Tests**
- Load testing on authentication endpoints
- Database query performance with large datasets
- Tenant query isolation performance
- Cache hit/miss ratios

### **Security Tests**
- JWT token manipulation attempts
- Cross-tenant data access prevention
- Rate limiting effectiveness
- SQL injection and XSS prevention

---

## 🚀 **Deployment & Monitoring**

### **Database Deployment**
```bash
# Run migrations
make migrate-up

# Seed initial data
make seed-data
```

### **Environment Configuration**
```yaml
# config/config.yaml additions
customer:
  jwt:
    secret: ${CUSTOMER_JWT_SECRET}
    access_token_duration: 1h
    refresh_token_duration: 168h # 7 days
  rate_limiting:
    auth_requests_per_minute: 10
    api_requests_per_minute: 100
  multi_tenancy:
    default_type: "shared"
    migration_thresholds:
      customer_count: 10000
      order_count: 50000
      avg_query_time: 100ms
```

### **Monitoring Setup**
- Prometheus metrics for customer operations
- Tenant-specific performance tracking
- Authentication success/failure rates
- Multi-tenant query latency monitoring

### **Performance Optimization**
- Database connection pooling per tenant type
- Redis caching for storefront and customer data
- Query optimization with proper indexing
- Background job processing for email verification

---

## 📊 **Implementation Summary**

### **Total Development Time**: 4 weeks (28 days)
### **Team Size**: 3 backend developers
### **Key Deliverables**:

✅ **Phase 1 (Week 1)**: Database foundation with multi-tenant design  
✅ **Phase 2 (Week 2)**: Repository layer with tenant-aware queries  
✅ **Phase 3 (Week 3)**: Business logic and authentication use cases  
✅ **Phase 4 (Week 4)**: API layer with security and seller integration  

### **Architecture Benefits**:
- **Future-Proof**: Abstract repository pattern supports migration to database-per-tenant
- **Secure**: JWT tokens with tenant scoping prevent cross-tenant access
- **Scalable**: Caching and query optimization for high-performance
- **Maintainable**: Clean architecture with clear separation of concerns

### **Ready for Production**:
- Complete multi-tenant customer management system
- Secure authentication with industry best practices  
- Seller dashboard integration for customer insights
- Performance monitoring and alerting setup
- Comprehensive testing coverage

This implementation provides a solid foundation for the SmartSeller storefront customer management system while maintaining the flexibility to scale to dedicated databases per tenant as the platform grows.

Would you like me to continue with **Phase 5: Advanced Features** or focus on any specific aspect of the implementation?