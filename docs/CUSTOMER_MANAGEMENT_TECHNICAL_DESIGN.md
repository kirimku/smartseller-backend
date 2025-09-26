# SmartSeller Multi-Tenant Customer Management - Technical Design

## 📋 **Document Overview**

**Document**: Multi-Tenant Customer Management & Authentication  
**Focus**: Customer Domain Architecture & Implementation Strategy  
**Version**: 1.0  
**Created**: September 26, 2025  
**Status**: Technical Design  
**Owner**: SmartSeller Development Team  

---

## 🎯 **Multi-Tenancy Architecture**

### **What is Multi-Tenancy?**

**Multi-tenancy** is an architectural pattern where a single application instance serves multiple tenants (customers/organizations), with complete data isolation between them.

In SmartSeller's case:
- **Tenant** = Each SmartSeller user (seller) with their storefront(s)
- **End Users** = Customers who shop on individual storefronts
- **Isolation** = Customer data is completely separated per storefront

```
SmartSeller Platform
├── Seller A (Tenant 1)
│   ├── Storefront "Fashion Store"
│   │   ├── Customer 1, Customer 2, Customer 3...
│   │   └── Orders, Carts, Profiles...
│   └── Storefront "Electronics Store" 
│       ├── Customer 4, Customer 5, Customer 6...
│       └── Orders, Carts, Profiles...
└── Seller B (Tenant 2)
    └── Storefront "Book Store"
        ├── Customer 7, Customer 8, Customer 9...
        └── Orders, Carts, Profiles...
```

### **Why Multi-Tenancy for Customer Management?**

✅ **Data Isolation**: Each seller's customers are completely separate  
✅ **Scalability**: Single database serves thousands of storefronts  
✅ **Cost Efficiency**: Shared infrastructure reduces operational costs  
✅ **Customization**: Each storefront can have unique settings  
✅ **Compliance**: Easy to manage data privacy per tenant  

---

## 🏗️ **Customer Management Architecture**

### **Tenant Identification Strategy**

We'll use **Storefront Slug** as the primary tenant identifier:

```
URL Structure:
- Seller API: /api/v1/... (authenticated with seller JWT)
- Customer API: /api/storefront/{slug}/... (tenant identified by slug)

Examples:
- https://api.smartseller.com/api/storefront/fashion-boutique/auth/login
- https://api.smartseller.com/api/storefront/tech-gadgets/profile
- https://api.smartseller.com/api/storefront/book-corner/cart
```

### **Data Isolation Model**

#### **Database-Level Isolation**
```sql
-- Every customer table includes storefront_id for tenant isolation
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    email VARCHAR(255) NOT NULL,
    -- ... other fields
    UNIQUE(storefront_id, email) -- Ensure email uniqueness per tenant
);

-- Row Level Security (RLS) for automatic filtering
CREATE POLICY customer_tenant_isolation ON customers 
    FOR ALL USING (storefront_id = current_setting('app.current_storefront_id')::UUID);

ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
```

#### **Application-Level Isolation**
```go
// Context-based tenant isolation
type TenantContext struct {
    StorefrontID uuid.UUID
    StorefrontSlug string
    SellerID     uuid.UUID
}

// All customer operations include tenant context
func (r *customerRepository) GetByEmail(ctx context.Context, tenantCtx *TenantContext, email string) (*entity.Customer, error) {
    query := `
        SELECT * FROM customers 
        WHERE storefront_id = $1 AND email = $2 AND deleted_at IS NULL
    `
    // Always filter by storefront_id
    return r.db.QueryRow(query, tenantCtx.StorefrontID, email)
}
```

---

## 👥 **Customer Domain Design**

### **Core Entities**

#### **Customer Entity**
```go
type Customer struct {
    // Primary identification
    ID           uuid.UUID `json:"id" db:"id"`
    StorefrontID uuid.UUID `json:"storefront_id" db:"storefront_id"` // TENANT KEY
    
    // Authentication
    Email        string `json:"email" db:"email"`
    Phone        *string `json:"phone" db:"phone"`
    PasswordHash string `json:"-" db:"password_hash"`
    PasswordSalt string `json:"-" db:"password_salt"`
    
    // Personal Information
    FirstName   string     `json:"first_name" db:"first_name"`
    LastName    string     `json:"last_name" db:"last_name"`
    DateOfBirth *time.Time `json:"date_of_birth" db:"date_of_birth"`
    Gender      *Gender    `json:"gender" db:"gender"`
    
    // Verification Status
    EmailVerifiedAt *time.Time `json:"email_verified_at" db:"email_verified_at"`
    PhoneVerifiedAt *time.Time `json:"phone_verified_at" db:"phone_verified_at"`
    
    // Account Status
    Status       CustomerStatus `json:"status" db:"status"`
    LastLoginAt  *time.Time     `json:"last_login_at" db:"last_login_at"`
    
    // Business Intelligence
    TotalOrders       int             `json:"total_orders" db:"total_orders"`
    TotalSpent        decimal.Decimal `json:"total_spent" db:"total_spent"`
    AverageOrderValue decimal.Decimal `json:"average_order_value" db:"average_order_value"`
    
    // Preferences & Settings
    Preferences CustomerPreferences `json:"preferences" db:"preferences"`
    
    // Security
    RefreshToken string `json:"-" db:"refresh_token"`
    
    // Audit
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CustomerStatus string
const (
    CustomerStatusActive    CustomerStatus = "active"
    CustomerStatusInactive  CustomerStatus = "inactive"
    CustomerStatusSuspended CustomerStatus = "suspended"
    CustomerStatusBlocked   CustomerStatus = "blocked"
)

type Gender string
const (
    GenderMale   Gender = "male"
    GenderFemale Gender = "female"
    GenderOther  Gender = "other"
)

type CustomerPreferences struct {
    Language           string `json:"language"`
    Currency           string `json:"currency"`
    EmailNotifications bool   `json:"email_notifications"`
    SMSNotifications   bool   `json:"sms_notifications"`
    MarketingEmails    bool   `json:"marketing_emails"`
}
```

#### **Customer Address Entity**
```go
type CustomerAddress struct {
    ID         uuid.UUID   `json:"id" db:"id"`
    CustomerID uuid.UUID   `json:"customer_id" db:"customer_id"`
    Type       AddressType `json:"type" db:"type"`
    
    // Address Information
    Label        string  `json:"label" db:"label"`
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
    
    // Geo-location (optional for delivery optimization)
    Coordinates *GeoPoint `json:"coordinates" db:"coordinates"`
    
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type AddressType string
const (
    AddressTypeBilling  AddressType = "billing"
    AddressTypeShipping AddressType = "shipping"
    AddressTypeBoth     AddressType = "both"
)
```

---

## 🔐 **Multi-Tenant Authentication System**

### **Authentication Flow Design**

#### **1. Customer Registration**
```go
// Registration Request
type CustomerRegisterRequest struct {
    Email       string `json:"email" validate:"required,email"`
    Password    string `json:"password" validate:"required,min=8"`
    FirstName   string `json:"first_name" validate:"required"`
    LastName    string `json:"last_name" validate:"required"`
    Phone       string `json:"phone" validate:"omitempty,e164"`
    AcceptTerms bool   `json:"accept_terms" validate:"required"`
}

// Registration Process
func (uc *customerUseCase) RegisterCustomer(ctx context.Context, storefrontSlug string, req *CustomerRegisterRequest) (*CustomerAuthResponse, error) {
    // 1. Get storefront by slug to establish tenant context
    storefront, err := uc.storefrontRepo.GetBySlug(ctx, storefrontSlug)
    if err != nil {
        return nil, ErrStorefrontNotFound
    }
    
    // 2. Check if email already exists for this tenant
    existingCustomer, _ := uc.customerRepo.GetByEmail(ctx, storefront.ID, req.Email)
    if existingCustomer != nil {
        return nil, ErrEmailAlreadyExists
    }
    
    // 3. Hash password with salt
    hashedPassword, salt, err := uc.passwordService.HashPassword(req.Password)
    if err != nil {
        return nil, err
    }
    
    // 4. Create customer entity
    customer := &entity.Customer{
        ID:           uuid.New(),
        StorefrontID: storefront.ID, // CRITICAL: Set tenant ID
        Email:        strings.ToLower(req.Email),
        PasswordHash: hashedPassword,
        PasswordSalt: salt,
        FirstName:    req.FirstName,
        LastName:     req.LastName,
        Phone:        &req.Phone,
        Status:       entity.CustomerStatusActive,
        Preferences:  getDefaultPreferences(storefront),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    // 5. Save customer to database
    if err := uc.customerRepo.Create(ctx, customer); err != nil {
        return nil, err
    }
    
    // 6. Generate JWT tokens
    accessToken, refreshToken, err := uc.generateCustomerTokens(customer, storefront)
    if err != nil {
        return nil, err
    }
    
    // 7. Send welcome email (async)
    go uc.emailService.SendWelcomeEmail(customer, storefront)
    
    return &CustomerAuthResponse{
        Customer:     convertToCustomerDTO(customer),
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    3600, // 1 hour
    }, nil
}
```

#### **2. Customer Login**
```go
// Login Request
type CustomerLoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// Login Process
func (uc *customerUseCase) LoginCustomer(ctx context.Context, storefrontSlug string, req *CustomerLoginRequest) (*CustomerAuthResponse, error) {
    // 1. Get storefront context
    storefront, err := uc.storefrontRepo.GetBySlug(ctx, storefrontSlug)
    if err != nil {
        return nil, ErrStorefrontNotFound
    }
    
    // 2. Find customer by email within tenant
    customer, err := uc.customerRepo.GetByEmail(ctx, storefront.ID, strings.ToLower(req.Email))
    if err != nil {
        return nil, ErrInvalidCredentials // Don't reveal if email exists
    }
    
    // 3. Verify account status
    if customer.Status != entity.CustomerStatusActive {
        return nil, ErrAccountSuspended
    }
    
    // 4. Verify password
    if !uc.passwordService.VerifyPassword(req.Password, customer.PasswordHash, customer.PasswordSalt) {
        return nil, ErrInvalidCredentials
    }
    
    // 5. Update last login timestamp
    if err := uc.customerRepo.UpdateLastLogin(ctx, customer.ID); err != nil {
        // Log error but don't fail login
        log.Printf("Failed to update last login for customer %s: %v", customer.ID, err)
    }
    
    // 6. Generate JWT tokens
    accessToken, refreshToken, err := uc.generateCustomerTokens(customer, storefront)
    if err != nil {
        return nil, err
    }
    
    // 7. Store refresh token
    if err := uc.customerRepo.UpdateRefreshToken(ctx, customer.ID, refreshToken); err != nil {
        return nil, err
    }
    
    return &CustomerAuthResponse{
        Customer:     convertToCustomerDTO(customer),
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    3600,
    }, nil
}
```

### **JWT Token Design**

#### **Customer JWT Claims**
```go
type CustomerClaims struct {
    CustomerID   string `json:"customer_id"`
    StorefrontID string `json:"storefront_id"` // CRITICAL: Tenant identifier
    Email        string `json:"email"`
    FirstName    string `json:"first_name"`
    Status       string `json:"status"`
    Scope        []string `json:"scope"` // ["customer:read", "customer:write"]
    TokenType    string `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}

// Token Generation
func (s *jwtService) GenerateCustomerTokens(customer *entity.Customer, storefront *entity.Storefront) (string, string, error) {
    now := time.Now()
    
    // Access Token (short-lived: 1 hour)
    accessClaims := &CustomerClaims{
        CustomerID:   customer.ID.String(),
        StorefrontID: storefront.ID.String(),
        Email:        customer.Email,
        FirstName:    customer.FirstName,
        Status:       string(customer.Status),
        Scope:        []string{"customer:read", "customer:write"},
        TokenType:    "access",
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    "smartseller-storefront",
            Subject:   customer.ID.String(),
            Audience:  []string{storefront.Slug},
            ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
            NotBefore: jwt.NewNumericDate(now),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }
    
    // Refresh Token (long-lived: 30 days)
    refreshClaims := &CustomerClaims{
        CustomerID:   customer.ID.String(),
        StorefrontID: storefront.ID.String(),
        TokenType:    "refresh",
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    "smartseller-storefront",
            Subject:   customer.ID.String(),
            Audience:  []string{storefront.Slug},
            ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }
    
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    
    accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", "", err
    }
    
    refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", "", err
    }
    
    return accessTokenString, refreshTokenString, nil
}
```

---

## 🛡️ **Multi-Tenant Security Middleware**

### **Storefront Context Middleware**
```go
// Extract storefront from URL slug and set context
func StorefrontContextMiddleware(storefrontRepo repository.StorefrontRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        slug := c.Param("slug")
        if slug == "" {
            c.JSON(400, gin.H{"error": "Storefront slug is required"})
            c.Abort()
            return
        }
        
        // Get storefront from cache or database
        storefront, err := getStorefrontBySlug(slug, storefrontRepo)
        if err != nil {
            c.JSON(404, gin.H{"error": "Storefront not found"})
            c.Abort()
            return
        }
        
        // Check if storefront is active
        if storefront.Status != entity.StorefrontStatusActive {
            c.JSON(503, gin.H{"error": "Storefront is currently unavailable"})
            c.Abort()
            return
        }
        
        // Set tenant context
        tenantCtx := &TenantContext{
            StorefrontID:   storefront.ID,
            StorefrontSlug: storefront.Slug,
            SellerID:       storefront.SellerID,
        }
        
        c.Set("tenant_context", tenantCtx)
        c.Set("storefront", storefront)
        c.Next()
    }
}

// Cache-enabled storefront lookup
func getStorefrontBySlug(slug string, repo repository.StorefrontRepository) (*entity.Storefront, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("storefront:%s", slug)
    if cachedStorefront := cache.Get(cacheKey); cachedStorefront != nil {
        return cachedStorefront.(*entity.Storefront), nil
    }
    
    // Fallback to database
    storefront, err := repo.GetBySlug(context.Background(), slug)
    if err != nil {
        return nil, err
    }
    
    // Cache for 1 hour
    cache.Set(cacheKey, storefront, time.Hour)
    return storefront, nil
}
```

### **Customer Authentication Middleware**
```go
// Validate customer JWT and ensure tenant matching
func CustomerAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get tenant context (must be set by StorefrontContextMiddleware)
        tenantCtx, exists := c.Get("tenant_context")
        if !exists {
            c.JSON(500, gin.H{"error": "Tenant context not found"})
            c.Abort()
            return
        }
        
        tenant := tenantCtx.(*TenantContext)
        
        // Extract JWT token
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Parse and validate token
        claims, err := parseCustomerToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // CRITICAL: Verify token belongs to current tenant
        if claims.StorefrontID != tenant.StorefrontID.String() {
            c.JSON(403, gin.H{"error": "Token not valid for this storefront"})
            c.Abort()
            return
        }
        
        // Verify token type
        if claims.TokenType != "access" {
            c.JSON(401, gin.H{"error": "Invalid token type"})
            c.Abort()
            return
        }
        
        // Set customer context
        c.Set("customer_id", claims.CustomerID)
        c.Set("customer_claims", claims)
        c.Next()
    }
}

// Optional authentication for guest users
func OptionalCustomerAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            // Guest user - continue without authentication
            c.Next()
            return
        }
        
        // Try to authenticate, but don't fail if invalid
        if strings.HasPrefix(authHeader, "Bearer ") {
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if claims, err := parseCustomerToken(tokenString); err == nil {
                // Verify tenant matching
                tenantCtx, _ := c.Get("tenant_context")
                if tenant := tenantCtx.(*TenantContext); tenant != nil {
                    if claims.StorefrontID == tenant.StorefrontID.String() && claims.TokenType == "access" {
                        c.Set("customer_id", claims.CustomerID)
                        c.Set("customer_claims", claims)
                    }
                }
            }
        }
        
        c.Next()
    }
}
```

---

## 🗄️ **Repository Implementation Strategy**

### **Customer Repository with Tenant Isolation**
```go
type CustomerRepository interface {
    // Core CRUD operations - all include tenant isolation
    Create(ctx context.Context, customer *entity.Customer) error
    GetByID(ctx context.Context, storefrontID, customerID uuid.UUID) (*entity.Customer, error)
    GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error)
    GetByPhone(ctx context.Context, storefrontID uuid.UUID, phone string) (*entity.Customer, error)
    Update(ctx context.Context, customer *entity.Customer) error
    SoftDelete(ctx context.Context, storefrontID, customerID uuid.UUID) error
    
    // Business queries
    GetByStorefront(ctx context.Context, req *GetCustomersRequest) (*CustomerListResponse, error)
    Search(ctx context.Context, storefrontID uuid.UUID, req *SearchCustomersRequest) (*CustomerSearchResult, error)
    GetTopCustomers(ctx context.Context, storefrontID uuid.UUID, limit int) ([]*entity.Customer, error)
    GetCustomerStats(ctx context.Context, storefrontID uuid.UUID) (*CustomerStats, error)
    
    // Authentication-specific
    UpdateLastLogin(ctx context.Context, storefrontID, customerID uuid.UUID) error
    UpdateRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string) error
    ValidateRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string) (bool, error)
}

// PostgreSQL Implementation
type customerRepository struct {
    db *sql.DB
}

func (r *customerRepository) GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error) {
    query := `
        SELECT id, storefront_id, email, password_hash, password_salt, 
               first_name, last_name, phone, date_of_birth, gender,
               email_verified_at, phone_verified_at, status, last_login_at,
               total_orders, total_spent, average_order_value,
               preferences, refresh_token, created_at, updated_at
        FROM customers 
        WHERE storefront_id = $1 AND email = $2 AND deleted_at IS NULL
    `
    
    row := r.db.QueryRowContext(ctx, query, storefrontID, strings.ToLower(email))
    
    customer := &entity.Customer{}
    var preferences []byte
    
    err := row.Scan(
        &customer.ID, &customer.StorefrontID, &customer.Email,
        &customer.PasswordHash, &customer.PasswordSalt,
        &customer.FirstName, &customer.LastName, &customer.Phone,
        &customer.DateOfBirth, &customer.Gender,
        &customer.EmailVerifiedAt, &customer.PhoneVerifiedAt,
        &customer.Status, &customer.LastLoginAt,
        &customer.TotalOrders, &customer.TotalSpent, &customer.AverageOrderValue,
        &preferences, &customer.RefreshToken,
        &customer.CreatedAt, &customer.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrCustomerNotFound
        }
        return nil, err
    }
    
    // Unmarshal preferences JSON
    if len(preferences) > 0 {
        if err := json.Unmarshal(preferences, &customer.Preferences); err != nil {
            return nil, fmt.Errorf("failed to unmarshal customer preferences: %w", err)
        }
    }
    
    return customer, nil
}

func (r *customerRepository) Create(ctx context.Context, customer *entity.Customer) error {
    query := `
        INSERT INTO customers (
            id, storefront_id, email, password_hash, password_salt,
            first_name, last_name, phone, date_of_birth, gender,
            status, preferences, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
    `
    
    preferencesJSON, err := json.Marshal(customer.Preferences)
    if err != nil {
        return fmt.Errorf("failed to marshal customer preferences: %w", err)
    }
    
    _, err = r.db.ExecContext(ctx, query,
        customer.ID, customer.StorefrontID, customer.Email,
        customer.PasswordHash, customer.PasswordSalt,
        customer.FirstName, customer.LastName, customer.Phone,
        customer.DateOfBirth, customer.Gender,
        customer.Status, preferencesJSON,
        customer.CreatedAt, customer.UpdatedAt,
    )
    
    if err != nil {
        if isUniqueConstraintError(err, "customers_storefront_id_email_key") {
            return ErrEmailAlreadyExists
        }
        return err
    }
    
    return nil
}
```

### **Database Indexes for Performance**
```sql
-- Critical indexes for multi-tenant customer queries
CREATE INDEX CONCURRENTLY idx_customers_storefront_email 
    ON customers(storefront_id, email) 
    WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY idx_customers_storefront_phone 
    ON customers(storefront_id, phone) 
    WHERE deleted_at IS NULL AND phone IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_customers_storefront_status 
    ON customers(storefront_id, status, created_at DESC) 
    WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY idx_customers_storefront_search 
    ON customers USING gin(to_tsvector('english', first_name || ' ' || last_name || ' ' || email)) 
    WHERE deleted_at IS NULL;

-- Address indexes
CREATE INDEX CONCURRENTLY idx_customer_addresses_customer_id 
    ON customer_addresses(customer_id);

CREATE INDEX CONCURRENTLY idx_customer_addresses_default 
    ON customer_addresses(customer_id, is_default) 
    WHERE is_default = true;
```

---

## 🔄 **API Design for Customer Management**

### **Customer-Facing APIs**
```
Base URL: /api/storefront/{slug}/

Authentication Endpoints:
POST   /auth/register           # Customer registration
POST   /auth/login              # Customer login
POST   /auth/logout             # Customer logout
POST   /auth/refresh            # Refresh access token
POST   /auth/forgot-password    # Request password reset
POST   /auth/reset-password     # Reset password with token
POST   /auth/verify-email       # Verify email address

Profile Management:
GET    /profile                 # Get customer profile
PUT    /profile                 # Update customer profile
POST   /profile/change-password # Change password
POST   /profile/upload-avatar   # Upload profile picture

Address Management:
GET    /addresses               # List customer addresses
POST   /addresses               # Add new address
GET    /addresses/{id}          # Get specific address
PUT    /addresses/{id}          # Update address
DELETE /addresses/{id}          # Delete address
POST   /addresses/{id}/default  # Set as default address
```

### **Seller-Facing Customer Management APIs**
```
Base URL: /api/v1/storefronts/{storefront_id}/

Customer Management:
GET    /customers               # List customers with pagination/filtering
GET    /customers/{id}          # Get customer details
PUT    /customers/{id}          # Update customer (limited fields)
DELETE /customers/{id}          # Soft delete customer
POST   /customers/{id}/suspend  # Suspend customer account
POST   /customers/{id}/activate # Activate customer account

Customer Analytics:
GET    /analytics/customers     # Customer acquisition metrics
GET    /customers/{id}/orders   # Customer order history
GET    /customers/{id}/analytics # Individual customer analytics
```

### **Sample API Responses**

#### **Customer Registration Response**
```json
{
    "success": true,
    "data": {
        "customer": {
            "id": "123e4567-e89b-12d3-a456-426614174000",
            "email": "john@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "phone": "+628123456789",
            "email_verified": false,
            "status": "active",
            "total_orders": 0,
            "total_spent": "0.00",
            "created_at": "2025-09-26T10:00:00Z"
        },
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_in": 3600
    },
    "message": "Registration successful. Please check your email to verify your account.",
    "request_id": "req_123456789",
    "timestamp": "2025-09-26T10:00:00Z"
}
```

#### **Customer Profile Response**
```json
{
    "success": true,
    "data": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "phone": "+628123456789",
        "date_of_birth": "1990-01-15",
        "gender": "male",
        "email_verified": true,
        "phone_verified": false,
        "status": "active",
        "total_orders": 5,
        "total_spent": "1,250.00",
        "average_order_value": "250.00",
        "preferences": {
            "language": "en",
            "currency": "IDR",
            "email_notifications": true,
            "sms_notifications": false,
            "marketing_emails": true
        },
        "addresses": [
            {
                "id": "addr_123456789",
                "type": "both",
                "label": "Home",
                "first_name": "John",
                "last_name": "Doe",
                "address_line1": "Jl. Sudirman No. 123",
                "city": "Jakarta",
                "province": "DKI Jakarta",
                "postal_code": "10110",
                "country": "Indonesia",
                "phone": "+628123456789",
                "is_default": true
            }
        ],
        "created_at": "2025-08-01T10:00:00Z",
        "updated_at": "2025-09-20T15:30:00Z"
    },
    "request_id": "req_987654321",
    "timestamp": "2025-09-26T10:00:00Z"
}
```

---

## 📊 **Best Practices & Recommendations**

### **Security Best Practices**

1. **Always Validate Tenant Context**
   ```go
   // WRONG - No tenant validation
   func GetCustomer(customerID uuid.UUID) (*Customer, error) {
       return repo.GetByID(customerID) // ❌ Cross-tenant access possible
   }
   
   // CORRECT - Tenant validation
   func GetCustomer(storefrontID, customerID uuid.UUID) (*Customer, error) {
       return repo.GetByID(storefrontID, customerID) // ✅ Tenant isolated
   }
   ```

2. **Use Row-Level Security (RLS)**
   ```sql
   -- Automatic tenant filtering at database level
   CREATE POLICY customer_isolation ON customers 
       FOR ALL USING (storefront_id = current_setting('app.current_storefront_id')::UUID);
   ```

3. **JWT Token Validation**
   - Always verify `storefront_id` in token matches current request
   - Use short-lived access tokens (1 hour)
   - Implement proper refresh token rotation

### **Performance Optimization**

1. **Caching Strategy**
   ```go
   // Cache frequently accessed data
   cacheKeys := map[string]time.Duration{
       "storefront:{slug}":                    time.Hour,        // Storefront details
       "customer:{storefront_id}:{customer_id}": 15 * time.Minute, // Customer profile
       "customer_addresses:{customer_id}":     30 * time.Minute, // Customer addresses
   }
   ```

2. **Database Query Optimization**
   - Use composite indexes on `(storefront_id, *)` columns
   - Implement pagination for customer lists
   - Use prepared statements for frequent queries

3. **Connection Pooling**
   ```go
   // Configure database connection pool
   config := pgxpool.Config{
       MaxConns:        30,
       MinConns:        5,
       MaxConnLifetime: time.Hour,
       MaxConnIdleTime: time.Minute * 30,
   }
   ```

### **Monitoring & Observability**

1. **Multi-Tenant Metrics**
   ```go
   // Track metrics per tenant
   metrics := map[string]prometheus.Counter{
       "customer_registrations_total": prometheus.NewCounterVec(
           prometheus.CounterOpts{Name: "customer_registrations_total"},
           []string{"storefront_id", "storefront_slug"},
       ),
       "customer_logins_total": prometheus.NewCounterVec(
           prometheus.CounterOpts{Name: "customer_logins_total"},
           []string{"storefront_id", "status"}, // success/failure
       ),
   }
   ```

2. **Audit Logging**
   ```go
   // Log all customer authentication events
   type AuditLog struct {
       StorefrontID string    `json:"storefront_id"`
       CustomerID   string    `json:"customer_id"`
       Action       string    `json:"action"`
       IPAddress    string    `json:"ip_address"`
       UserAgent    string    `json:"user_agent"`
       Timestamp    time.Time `json:"timestamp"`
       Success      bool      `json:"success"`
       Details      string    `json:"details"`
   }
   ```

---

## 🚀 **Implementation Priority**

### **Phase 1: Core Customer Management (Week 1-2)**
1. **Database schema** and migrations
2. **Customer entity** with validation rules
3. **Repository layer** with tenant isolation
4. **Basic CRUD operations** with proper security

### **Phase 2: Authentication System (Week 3)**
1. **JWT service** with multi-tenant claims
2. **Registration and login** endpoints
3. **Password management** (reset, change)
4. **Email verification** system

### **Phase 3: Profile & Address Management (Week 4)**
1. **Customer profile** management APIs
2. **Address management** system
3. **Customer preferences** handling
4. **Profile picture upload**

### **Phase 4: Advanced Features (Week 5)**
1. **Customer analytics** for sellers
2. **Search and filtering** capabilities
3. **Customer segmentation** tools
4. **Performance optimization** and caching

---

## ✅ **Success Criteria**

### **Technical Requirements**
- ✅ Complete data isolation between storefronts
- ✅ JWT authentication with proper tenant validation
- ✅ API response times < 200ms for customer operations
- ✅ Support for 10,000+ customers per storefront
- ✅ Email verification and password reset working

### **Security Requirements**
- ✅ No cross-tenant data leakage possible
- ✅ Secure password handling with salt
- ✅ JWT tokens properly scoped to storefronts
- ✅ Rate limiting prevents brute force attacks
- ✅ All sensitive operations logged for audit

### **Business Requirements**
- ✅ Customers can register in under 2 minutes
- ✅ Login process completes in under 30 seconds
- ✅ Profile updates reflect immediately
- ✅ Sellers can view customer analytics
- ✅ Support for guest checkout (future)

---

**Document Status**: Ready for Implementation  
**Next Step**: Begin database schema design for customer tables  
**Priority**: High - Foundation for entire storefront system  

This multi-tenant customer management system provides the secure, scalable foundation needed for SmartSeller's B2B2C platform, ensuring complete data isolation while maintaining performance and ease of use.