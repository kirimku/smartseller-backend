# SmartSeller Storefront & Customer Management - Technical Architecture

## 📋 **Document Overview**

**Document**: Technical Architecture Specification  
**Product**: SmartSeller Storefront & Customer Management System  
**Version**: 1.0  
**Created**: September 26, 2025  
**Status**: Draft  
**Owner**: SmartSeller Development Team  

---

## 🏗️ **Architecture Overview**

### **System Context**
The SmartSeller platform operates as a B2B2C system where:
- **B2B**: SmartSeller provides tools to sellers (businesses)
- **B2C**: Sellers use these tools to serve their customers
- **Multi-Tenant**: Each seller can have multiple storefronts, each serving different customer bases

### **Architectural Principles**
- **Domain-Driven Design**: Clear domain boundaries and ubiquitous language
- **Clean Architecture**: Separation of concerns with dependency inversion
- **Multi-Tenant**: Complete data isolation between different sellers and storefronts
- **API-First**: All functionality exposed through well-designed APIs
- **Event-Driven**: Asynchronous processing for scalability and resilience

---

## 🎯 **Domain Model**

### **Core Domains**

#### **1. Seller Domain** (Existing)
```
Responsibilities:
- Seller authentication and authorization
- Business profile management
- Subscription and billing management
- Business analytics and reporting

Key Entities:
- User (Seller)
- UserRole, UserTier, UserType
- Business Profile
- Subscription
```

#### **2. Product Domain** (Existing)
```
Responsibilities:
- Product catalog management
- Inventory tracking and management
- Pricing and discount management
- Product categorization and variants

Key Entities:
- Product
- ProductVariant, ProductVariantOption
- ProductCategory
- ProductImage
- Inventory
```

#### **3. Storefront Domain** (New)
```
Responsibilities:
- Storefront configuration and branding
- Theme and layout management
- Domain and subdomain management
- SEO and marketing settings
- Multi-language and currency support

Key Entities:
- Storefront
- StorefrontConfig
- StorefrontTheme
- StorefrontDomain
- StorefrontAnalytics
```

#### **4. Customer Domain** (New)
```
Responsibilities:
- Customer authentication and authorization
- Profile and preference management
- Address and contact information
- Customer segmentation and analytics
- Communication preferences

Key Entities:
- Customer
- CustomerProfile
- CustomerAddress
- CustomerPreferences
- CustomerSegment
```

#### **5. Commerce Domain** (New)
```
Responsibilities:
- Shopping cart management
- Order processing and fulfillment
- Payment processing integration
- Shipping and logistics integration
- Return and refund management

Key Entities:
- ShoppingCart
- CartItem
- Order
- OrderItem
- Payment
- Shipment
- Return
```

#### **6. Communication Domain** (New)
```
Responsibilities:
- Email and SMS notifications
- Marketing campaign management
- Customer support ticketing
- Review and feedback collection
- Real-time messaging

Key Entities:
- EmailTemplate
- NotificationQueue
- CampaignMessage
- SupportTicket
- CustomerReview
```

---

## 🗄️ **Detailed Data Model**

### **Storefront Management**

#### **Storefront Entity**
```go
type Storefront struct {
    ID           uuid.UUID          `json:"id" db:"id"`
    SellerID     uuid.UUID          `json:"seller_id" db:"seller_id"`
    Name         string             `json:"name" db:"name"`
    Slug         string             `json:"slug" db:"slug"`
    Description  *string            `json:"description" db:"description"`
    Domain       *string            `json:"domain" db:"domain"`
    Subdomain    *string            `json:"subdomain" db:"subdomain"`
    Status       StorefrontStatus   `json:"status" db:"status"`
    Currency     string             `json:"currency" db:"currency"`
    Language     string             `json:"language" db:"language"`
    Timezone     string             `json:"timezone" db:"timezone"`
    Settings     StorefrontSettings `json:"settings" db:"settings"`
    CreatedAt    time.Time          `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time          `json:"updated_at" db:"updated_at"`
    DeletedAt    *time.Time         `json:"deleted_at,omitempty" db:"deleted_at"`
}

type StorefrontStatus string
const (
    StorefrontStatusActive      StorefrontStatus = "active"
    StorefrontStatusInactive    StorefrontStatus = "inactive"
    StorefrontStatusMaintenance StorefrontStatus = "maintenance"
    StorefrontStatusSuspended   StorefrontStatus = "suspended"
)

type StorefrontSettings struct {
    EnableGuestCheckout    bool                   `json:"enable_guest_checkout"`
    RequireEmailVerification bool               `json:"require_email_verification"`
    AllowReviews          bool                   `json:"allow_reviews"`
    EnableWishlist        bool                   `json:"enable_wishlist"`
    MinOrderAmount        *decimal.Decimal       `json:"min_order_amount"`
    MaxOrderAmount        *decimal.Decimal       `json:"max_order_amount"`
    TaxSettings           TaxSettings            `json:"tax_settings"`
    ShippingSettings      ShippingSettings       `json:"shipping_settings"`
    PaymentMethods        []PaymentMethodConfig  `json:"payment_methods"`
}
```

#### **StorefrontConfig Entity**
```go
type StorefrontConfig struct {
    ID                uuid.UUID    `json:"id" db:"id"`
    StorefrontID      uuid.UUID    `json:"storefront_id" db:"storefront_id"`
    LogoURL           *string      `json:"logo_url" db:"logo_url"`
    FaviconURL        *string      `json:"favicon_url" db:"favicon_url"`
    PrimaryColor      string       `json:"primary_color" db:"primary_color"`
    SecondaryColor    string       `json:"secondary_color" db:"secondary_color"`
    AccentColor       *string      `json:"accent_color" db:"accent_color"`
    FontFamily        string       `json:"font_family" db:"font_family"`
    MetaTitle         *string      `json:"meta_title" db:"meta_title"`
    MetaDescription   *string      `json:"meta_description" db:"meta_description"`
    SocialMediaLinks  SocialLinks  `json:"social_media_links" db:"social_media_links"`
    ContactInfo       ContactInfo  `json:"contact_info" db:"contact_info"`
    BusinessHours     BusinessHours `json:"business_hours" db:"business_hours"`
    CustomCSS         *string      `json:"custom_css" db:"custom_css"`
    CustomJS          *string      `json:"custom_js" db:"custom_js"`
    GoogleAnalyticsID *string      `json:"google_analytics_id" db:"google_analytics_id"`
    FacebookPixelID   *string      `json:"facebook_pixel_id" db:"facebook_pixel_id"`
    CreatedAt         time.Time    `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time    `json:"updated_at" db:"updated_at"`
}

type SocialLinks struct {
    Facebook  *string `json:"facebook"`
    Instagram *string `json:"instagram"`
    Twitter   *string `json:"twitter"`
    YouTube   *string `json:"youtube"`
    LinkedIn  *string `json:"linkedin"`
    TikTok    *string `json:"tiktok"`
}

type ContactInfo struct {
    Phone       *string `json:"phone"`
    Email       *string `json:"email"`
    Address     *string `json:"address"`
    WhatsApp    *string `json:"whatsapp"`
    Telegram    *string `json:"telegram"`
}

type BusinessHours struct {
    Monday    DaySchedule `json:"monday"`
    Tuesday   DaySchedule `json:"tuesday"`
    Wednesday DaySchedule `json:"wednesday"`
    Thursday  DaySchedule `json:"thursday"`
    Friday    DaySchedule `json:"friday"`
    Saturday  DaySchedule `json:"saturday"`
    Sunday    DaySchedule `json:"sunday"`
}

type DaySchedule struct {
    IsOpen    bool   `json:"is_open"`
    OpenTime  string `json:"open_time"`  // HH:MM format
    CloseTime string `json:"close_time"` // HH:MM format
}
```

### **Customer Management**

#### **Customer Entity**
```go
type Customer struct {
    ID                uuid.UUID         `json:"id" db:"id"`
    StorefrontID      uuid.UUID         `json:"storefront_id" db:"storefront_id"`
    Email             string            `json:"email" db:"email"`
    Phone             *string           `json:"phone" db:"phone"`
    PasswordHash      string            `json:"-" db:"password_hash"`
    PasswordSalt      string            `json:"-" db:"password_salt"`
    FirstName         string            `json:"first_name" db:"first_name"`
    LastName          string            `json:"last_name" db:"last_name"`
    DateOfBirth       *time.Time        `json:"date_of_birth" db:"date_of_birth"`
    Gender            *CustomerGender   `json:"gender" db:"gender"`
    ProfilePicture    *string           `json:"profile_picture" db:"profile_picture"`
    EmailVerifiedAt   *time.Time        `json:"email_verified_at" db:"email_verified_at"`
    PhoneVerifiedAt   *time.Time        `json:"phone_verified_at" db:"phone_verified_at"`
    LastLoginAt       *time.Time        `json:"last_login_at" db:"last_login_at"`
    Status            CustomerStatus    `json:"status" db:"status"`
    CustomerGroup     CustomerGroup     `json:"customer_group" db:"customer_group"`
    TotalOrders       int               `json:"total_orders" db:"total_orders"`
    TotalSpent        decimal.Decimal   `json:"total_spent" db:"total_spent"`
    AverageOrderValue decimal.Decimal   `json:"average_order_value" db:"average_order_value"`
    LifetimeValue     decimal.Decimal   `json:"lifetime_value" db:"lifetime_value"`
    Preferences       CustomerPrefs     `json:"preferences" db:"preferences"`
    Tags              []string          `json:"tags" db:"tags"`
    Notes             *string           `json:"notes" db:"notes"`
    RefreshToken      string            `json:"-" db:"refresh_token"`
    CreatedAt         time.Time         `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
    DeletedAt         *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CustomerGender string
const (
    GenderMale   CustomerGender = "male"
    GenderFemale CustomerGender = "female"
    GenderOther  CustomerGender = "other"
)

type CustomerStatus string
const (
    CustomerStatusActive    CustomerStatus = "active"
    CustomerStatusInactive  CustomerStatus = "inactive"
    CustomerStatusSuspended CustomerStatus = "suspended"
    CustomerStatusBlocked   CustomerStatus = "blocked"
)

type CustomerGroup string
const (
    CustomerGroupRegular CustomerGroup = "regular"
    CustomerGroupVIP     CustomerGroup = "vip"
    CustomerGroupBronze  CustomerGroup = "bronze"
    CustomerGroupSilver  CustomerGroup = "silver"
    CustomerGroupGold    CustomerGroup = "gold"
)

type CustomerPrefs struct {
    Language            string `json:"language"`
    Currency            string `json:"currency"`
    EmailNotifications  bool   `json:"email_notifications"`
    SMSNotifications    bool   `json:"sms_notifications"`
    MarketingEmails     bool   `json:"marketing_emails"`
    OrderUpdates        bool   `json:"order_updates"`
    NewsletterSubscribed bool  `json:"newsletter_subscribed"`
}
```

#### **CustomerAddress Entity**
```go
type CustomerAddress struct {
    ID           uuid.UUID `json:"id" db:"id"`
    CustomerID   uuid.UUID `json:"customer_id" db:"customer_id"`
    Type         AddressType `json:"type" db:"type"`
    Label        string    `json:"label" db:"label"`
    FirstName    string    `json:"first_name" db:"first_name"`
    LastName     string    `json:"last_name" db:"last_name"`
    Company      *string   `json:"company" db:"company"`
    AddressLine1 string    `json:"address_line1" db:"address_line1"`
    AddressLine2 *string   `json:"address_line2" db:"address_line2"`
    City         string    `json:"city" db:"city"`
    Province     string    `json:"province" db:"province"`
    PostalCode   string    `json:"postal_code" db:"postal_code"`
    Country      string    `json:"country" db:"country"`
    Phone        *string   `json:"phone" db:"phone"`
    IsDefault    bool      `json:"is_default" db:"is_default"`
    Coordinates  *GeoPoint `json:"coordinates" db:"coordinates"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type AddressType string
const (
    AddressTypeBilling  AddressType = "billing"
    AddressTypeShipping AddressType = "shipping"
    AddressTypeBoth     AddressType = "both"
)

type GeoPoint struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}
```

### **Commerce Management**

#### **ShoppingCart Entity**
```go
type ShoppingCart struct {
    ID             uuid.UUID       `json:"id" db:"id"`
    StorefrontID   uuid.UUID       `json:"storefront_id" db:"storefront_id"`
    CustomerID     *uuid.UUID      `json:"customer_id" db:"customer_id"`
    SessionID      *string         `json:"session_id" db:"session_id"`
    Currency       string          `json:"currency" db:"currency"`
    ItemCount      int             `json:"item_count" db:"item_count"`
    Subtotal       decimal.Decimal `json:"subtotal" db:"subtotal"`
    TaxAmount      decimal.Decimal `json:"tax_amount" db:"tax_amount"`
    ShippingAmount decimal.Decimal `json:"shipping_amount" db:"shipping_amount"`
    DiscountAmount decimal.Decimal `json:"discount_amount" db:"discount_amount"`
    TotalAmount    decimal.Decimal `json:"total_amount" db:"total_amount"`
    DiscountCode   *string         `json:"discount_code" db:"discount_code"`
    ShippingMethod *string         `json:"shipping_method" db:"shipping_method"`
    Notes          *string         `json:"notes" db:"notes"`
    ExpiresAt      *time.Time      `json:"expires_at" db:"expires_at"`
    CreatedAt      time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
    
    // Related entities (loaded separately)
    Items []CartItem `json:"items,omitempty" db:"-"`
}
```

#### **CartItem Entity**
```go
type CartItem struct {
    ID               uuid.UUID       `json:"id" db:"id"`
    CartID           uuid.UUID       `json:"cart_id" db:"cart_id"`
    ProductID        uuid.UUID       `json:"product_id" db:"product_id"`
    ProductVariantID *uuid.UUID      `json:"product_variant_id" db:"product_variant_id"`
    Quantity         int             `json:"quantity" db:"quantity"`
    UnitPrice        decimal.Decimal `json:"unit_price" db:"unit_price"`
    TotalPrice       decimal.Decimal `json:"total_price" db:"total_price"`
    
    // Product information snapshot
    ProductName    string  `json:"product_name" db:"product_name"`
    ProductSKU     string  `json:"product_sku" db:"product_sku"`
    ProductImage   *string `json:"product_image" db:"product_image"`
    VariantOptions *string `json:"variant_options" db:"variant_options"`
    
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

#### **Order Entity**
```go
type Order struct {
    ID                 uuid.UUID         `json:"id" db:"id"`
    StorefrontID       uuid.UUID         `json:"storefront_id" db:"storefront_id"`
    CustomerID         *uuid.UUID        `json:"customer_id" db:"customer_id"`
    OrderNumber        string            `json:"order_number" db:"order_number"`
    Status             OrderStatus       `json:"status" db:"status"`
    PaymentStatus      PaymentStatus     `json:"payment_status" db:"payment_status"`
    FulfillmentStatus  FulfillmentStatus `json:"fulfillment_status" db:"fulfillment_status"`
    
    // Customer information (snapshot)
    CustomerEmail     string  `json:"customer_email" db:"customer_email"`
    CustomerPhone     *string `json:"customer_phone" db:"customer_phone"`
    CustomerFirstName string  `json:"customer_first_name" db:"customer_first_name"`
    CustomerLastName  string  `json:"customer_last_name" db:"customer_last_name"`
    
    // Financial information
    Currency       string          `json:"currency" db:"currency"`
    ItemCount      int             `json:"item_count" db:"item_count"`
    Subtotal       decimal.Decimal `json:"subtotal" db:"subtotal"`
    TaxAmount      decimal.Decimal `json:"tax_amount" db:"tax_amount"`
    ShippingAmount decimal.Decimal `json:"shipping_amount" db:"shipping_amount"`
    DiscountAmount decimal.Decimal `json:"discount_amount" db:"discount_amount"`
    TotalAmount    decimal.Decimal `json:"total_amount" db:"total_amount"`
    PaidAmount     decimal.Decimal `json:"paid_amount" db:"paid_amount"`
    RefundedAmount decimal.Decimal `json:"refunded_amount" db:"refunded_amount"`
    
    // Addresses (as JSON)
    BillingAddress  Address `json:"billing_address" db:"billing_address"`
    ShippingAddress Address `json:"shipping_address" db:"shipping_address"`
    
    // Additional information
    Notes           *string `json:"notes" db:"notes"`
    InternalNotes   *string `json:"internal_notes" db:"internal_notes"`
    DiscountCode    *string `json:"discount_code" db:"discount_code"`
    ShippingMethod  *string `json:"shipping_method" db:"shipping_method"`
    TrackingNumber  *string `json:"tracking_number" db:"tracking_number"`
    TrackingURL     *string `json:"tracking_url" db:"tracking_url"`
    
    // Timestamps
    OrderedAt   time.Time  `json:"ordered_at" db:"ordered_at"`
    ProcessedAt *time.Time `json:"processed_at" db:"processed_at"`
    ShippedAt   *time.Time `json:"shipped_at" db:"shipped_at"`
    DeliveredAt *time.Time `json:"delivered_at" db:"delivered_at"`
    CancelledAt *time.Time `json:"cancelled_at" db:"cancelled_at"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
    
    // Related entities (loaded separately)
    Items    []OrderItem `json:"items,omitempty" db:"-"`
    Payments []Payment   `json:"payments,omitempty" db:"-"`
    Shipments []Shipment `json:"shipments,omitempty" db:"-"`
}

type OrderStatus string
const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusConfirmed OrderStatus = "confirmed"
    OrderStatusProcessing OrderStatus = "processing"
    OrderStatusShipped   OrderStatus = "shipped"
    OrderStatusDelivered OrderStatus = "delivered"
    OrderStatusCompleted OrderStatus = "completed"
    OrderStatusCancelled OrderStatus = "cancelled"
    OrderStatusRefunded  OrderStatus = "refunded"
)

type PaymentStatus string
const (
    PaymentStatusPending   PaymentStatus = "pending"
    PaymentStatusPaid      PaymentStatus = "paid"
    PaymentStatusFailed    PaymentStatus = "failed"
    PaymentStatusRefunded  PaymentStatus = "refunded"
    PaymentStatusPartial   PaymentStatus = "partial"
)

type FulfillmentStatus string
const (
    FulfillmentStatusUnfulfilled FulfillmentStatus = "unfulfilled"
    FulfillmentStatusPartial     FulfillmentStatus = "partial"
    FulfillmentStatusFulfilled   FulfillmentStatus = "fulfilled"
    FulfillmentStatusShipped     FulfillmentStatus = "shipped"
    FulfillmentStatusDelivered   FulfillmentStatus = "delivered"
)
```

#### **OrderItem Entity**
```go
type OrderItem struct {
    ID               uuid.UUID       `json:"id" db:"id"`
    OrderID          uuid.UUID       `json:"order_id" db:"order_id"`
    ProductID        uuid.UUID       `json:"product_id" db:"product_id"`
    ProductVariantID *uuid.UUID      `json:"product_variant_id" db:"product_variant_id"`
    
    // Product information (snapshot)
    ProductName    string  `json:"product_name" db:"product_name"`
    ProductSKU     string  `json:"product_sku" db:"product_sku"`
    VariantName    *string `json:"variant_name" db:"variant_name"`
    VariantSKU     *string `json:"variant_sku" db:"variant_sku"`
    VariantOptions *string `json:"variant_options" db:"variant_options"`
    ProductImage   *string `json:"product_image" db:"product_image"`
    
    Quantity         int             `json:"quantity" db:"quantity"`
    UnitPrice        decimal.Decimal `json:"unit_price" db:"unit_price"`
    TotalPrice       decimal.Decimal `json:"total_price" db:"total_price"`
    FulfillmentStatus FulfillmentStatus `json:"fulfillment_status" db:"fulfillment_status"`
    
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

---

## 🔧 **Service Layer Architecture**

### **Repository Pattern Implementation**

#### **Storefront Repository**
```go
type StorefrontRepository interface {
    // Core CRUD
    Create(ctx context.Context, storefront *entity.Storefront) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Storefront, error)
    GetBySlug(ctx context.Context, slug string) (*entity.Storefront, error)
    GetBySellerID(ctx context.Context, sellerID uuid.UUID) ([]*entity.Storefront, error)
    Update(ctx context.Context, storefront *entity.Storefront) error
    Delete(ctx context.Context, id uuid.UUID) error
    
    // Business queries
    GetActiveStorefronts(ctx context.Context) ([]*entity.Storefront, error)
    GetByDomain(ctx context.Context, domain string) (*entity.Storefront, error)
    Search(ctx context.Context, req *dto.SearchStorefrontsRequest) (*dto.StorefrontSearchResult, error)
}

type StorefrontConfigRepository interface {
    GetByStorefrontID(ctx context.Context, storefrontID uuid.UUID) (*entity.StorefrontConfig, error)
    Upsert(ctx context.Context, config *entity.StorefrontConfig) error
    Delete(ctx context.Context, storefrontID uuid.UUID) error
}
```

#### **Customer Repository**
```go
type CustomerRepository interface {
    // Core CRUD
    Create(ctx context.Context, customer *entity.Customer) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
    GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error)
    GetByPhone(ctx context.Context, storefrontID uuid.UUID, phone string) (*entity.Customer, error)
    Update(ctx context.Context, customer *entity.Customer) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    
    // Business queries
    GetByStorefront(ctx context.Context, req *dto.GetCustomersRequest) (*dto.CustomerListResponse, error)
    Search(ctx context.Context, req *dto.SearchCustomersRequest) (*dto.CustomerSearchResult, error)
    GetTopCustomers(ctx context.Context, storefrontID uuid.UUID, limit int) ([]*entity.Customer, error)
    GetCustomerStats(ctx context.Context, storefrontID uuid.UUID) (*dto.CustomerStats, error)
    
    // Authentication
    UpdateLastLogin(ctx context.Context, customerID uuid.UUID) error
    UpdateRefreshToken(ctx context.Context, customerID uuid.UUID, token string) error
}

type CustomerAddressRepository interface {
    Create(ctx context.Context, address *entity.CustomerAddress) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.CustomerAddress, error)
    GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error)
    Update(ctx context.Context, address *entity.CustomerAddress) error
    Delete(ctx context.Context, id uuid.UUID) error
    SetDefault(ctx context.Context, customerID, addressID uuid.UUID) error
}
```

#### **Commerce Repository**
```go
type ShoppingCartRepository interface {
    // Cart management
    Create(ctx context.Context, cart *entity.ShoppingCart) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.ShoppingCart, error)
    GetByCustomerID(ctx context.Context, storefrontID, customerID uuid.UUID) (*entity.ShoppingCart, error)
    GetBySessionID(ctx context.Context, storefrontID uuid.UUID, sessionID string) (*entity.ShoppingCart, error)
    Update(ctx context.Context, cart *entity.ShoppingCart) error
    Delete(ctx context.Context, id uuid.UUID) error
    
    // Cart merging and cleanup
    MergeCart(ctx context.Context, fromCartID, toCartID uuid.UUID) error
    CleanupExpiredCarts(ctx context.Context) error
}

type CartItemRepository interface {
    Create(ctx context.Context, item *entity.CartItem) error
    GetByCartID(ctx context.Context, cartID uuid.UUID) ([]*entity.CartItem, error)
    Update(ctx context.Context, item *entity.CartItem) error
    Delete(ctx context.Context, id uuid.UUID) error
    DeleteByCartID(ctx context.Context, cartID uuid.UUID) error
    
    // Business operations
    UpsertItem(ctx context.Context, item *entity.CartItem) error
    GetCartSummary(ctx context.Context, cartID uuid.UUID) (*dto.CartSummary, error)
}

type OrderRepository interface {
    // Core CRUD
    Create(ctx context.Context, order *entity.Order) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
    GetByOrderNumber(ctx context.Context, orderNumber string) (*entity.Order, error)
    Update(ctx context.Context, order *entity.Order) error
    
    // Business queries
    GetByCustomer(ctx context.Context, req *dto.GetOrdersRequest) (*dto.OrderListResponse, error)
    GetByStorefront(ctx context.Context, req *dto.GetStorefrontOrdersRequest) (*dto.OrderListResponse, error)
    Search(ctx context.Context, req *dto.SearchOrdersRequest) (*dto.OrderSearchResult, error)
    
    // Analytics
    GetOrderStats(ctx context.Context, storefrontID uuid.UUID, period string) (*dto.OrderStats, error)
    GetRevenueStats(ctx context.Context, storefrontID uuid.UUID, period string) (*dto.RevenueStats, error)
}
```

### **Use Case Layer**

#### **Storefront Use Cases**
```go
type StorefrontUseCase interface {
    // Storefront management
    CreateStorefront(ctx context.Context, req *dto.CreateStorefrontRequest) (*dto.StorefrontResponse, error)
    UpdateStorefront(ctx context.Context, id uuid.UUID, req *dto.UpdateStorefrontRequest) (*dto.StorefrontResponse, error)
    GetStorefront(ctx context.Context, id uuid.UUID) (*dto.StorefrontResponse, error)
    GetStorefrontBySlug(ctx context.Context, slug string) (*dto.StorefrontResponse, error)
    ListStorefronts(ctx context.Context, sellerID uuid.UUID) ([]*dto.StorefrontResponse, error)
    DeleteStorefront(ctx context.Context, id uuid.UUID) error
    
    // Configuration
    UpdateStorefrontConfig(ctx context.Context, storefrontID uuid.UUID, req *dto.UpdateStorefrontConfigRequest) error
    GetStorefrontConfig(ctx context.Context, storefrontID uuid.UUID) (*dto.StorefrontConfigResponse, error)
    
    // Domain management
    SetCustomDomain(ctx context.Context, storefrontID uuid.UUID, domain string) error
    VerifyDomain(ctx context.Context, storefrontID uuid.UUID) error
}
```

#### **Customer Use Cases**
```go
type CustomerUseCase interface {
    // Authentication
    RegisterCustomer(ctx context.Context, req *dto.CustomerRegisterRequest) (*dto.CustomerAuthResponse, error)
    LoginCustomer(ctx context.Context, req *dto.CustomerLoginRequest) (*dto.CustomerAuthResponse, error)
    RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.CustomerAuthResponse, error)
    LogoutCustomer(ctx context.Context, customerID uuid.UUID) error
    
    // Profile management
    GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*dto.CustomerProfileResponse, error)
    UpdateCustomerProfile(ctx context.Context, customerID uuid.UUID, req *dto.UpdateCustomerProfileRequest) error
    ChangePassword(ctx context.Context, customerID uuid.UUID, req *dto.ChangePasswordRequest) error
    
    // Address management
    AddAddress(ctx context.Context, customerID uuid.UUID, req *dto.AddAddressRequest) (*dto.CustomerAddressResponse, error)
    UpdateAddress(ctx context.Context, addressID uuid.UUID, req *dto.UpdateAddressRequest) error
    DeleteAddress(ctx context.Context, addressID uuid.UUID) error
    SetDefaultAddress(ctx context.Context, customerID, addressID uuid.UUID) error
    
    // For sellers
    GetStorefrontCustomers(ctx context.Context, req *dto.GetCustomersRequest) (*dto.CustomerListResponse, error)
    GetCustomerAnalytics(ctx context.Context, storefrontID uuid.UUID) (*dto.CustomerAnalytics, error)
}
```

#### **Commerce Use Cases**
```go
type CommerceUseCase interface {
    // Cart management
    GetCart(ctx context.Context, req *dto.GetCartRequest) (*dto.CartResponse, error)
    AddToCart(ctx context.Context, req *dto.AddToCartRequest) (*dto.CartResponse, error)
    UpdateCartItem(ctx context.Context, req *dto.UpdateCartItemRequest) (*dto.CartResponse, error)
    RemoveFromCart(ctx context.Context, req *dto.RemoveFromCartRequest) (*dto.CartResponse, error)
    ClearCart(ctx context.Context, req *dto.ClearCartRequest) error
    
    // Checkout
    InitiateCheckout(ctx context.Context, req *dto.CheckoutRequest) (*dto.CheckoutResponse, error)
    ApplyDiscount(ctx context.Context, req *dto.ApplyDiscountRequest) (*dto.CartResponse, error)
    CalculateShipping(ctx context.Context, req *dto.CalculateShippingRequest) (*dto.ShippingOptionsResponse, error)
    
    // Order management
    CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
    GetOrder(ctx context.Context, orderID uuid.UUID) (*dto.OrderResponse, error)
    UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, req *dto.UpdateOrderStatusRequest) error
    CancelOrder(ctx context.Context, orderID uuid.UUID, req *dto.CancelOrderRequest) error
    
    // Customer orders
    GetCustomerOrders(ctx context.Context, req *dto.GetOrdersRequest) (*dto.OrderListResponse, error)
    TrackOrder(ctx context.Context, orderNumber string) (*dto.OrderTrackingResponse, error)
}
```

---

## 🌐 **API Design**

### **Multi-Tenant API Structure**

#### **Seller APIs** (Authenticated with Bearer Token)
```
Base URL: /api/v1/

Storefront Management:
GET    /storefronts                    # List seller's storefronts
POST   /storefronts                    # Create new storefront
GET    /storefronts/{id}               # Get storefront details
PUT    /storefronts/{id}               # Update storefront
DELETE /storefronts/{id}               # Delete storefront

Storefront Configuration:
GET    /storefronts/{id}/config        # Get configuration
PUT    /storefronts/{id}/config        # Update configuration
POST   /storefronts/{id}/domain        # Set custom domain
POST   /storefronts/{id}/verify-domain # Verify domain

Customer Management:
GET    /storefronts/{id}/customers     # List customers
GET    /storefronts/{id}/customers/{customer_id} # Get customer details
GET    /storefronts/{id}/analytics/customers # Customer analytics

Order Management:
GET    /storefronts/{id}/orders        # List orders
GET    /storefronts/{id}/orders/{order_id} # Get order details
PUT    /storefronts/{id}/orders/{order_id}/status # Update order status
GET    /storefronts/{id}/analytics/orders # Order analytics
```

#### **Customer APIs** (Public + Customer Token)
```
Base URL: /api/storefront/{slug}/

Authentication:
POST   /auth/register                  # Customer registration
POST   /auth/login                     # Customer login
POST   /auth/logout                    # Customer logout
POST   /auth/refresh                   # Refresh token
POST   /auth/forgot-password           # Password reset request
POST   /auth/reset-password            # Password reset

Profile Management:
GET    /profile                        # Get customer profile
PUT    /profile                        # Update customer profile
POST   /profile/change-password        # Change password

Address Management:
GET    /addresses                      # List customer addresses
POST   /addresses                      # Add new address
PUT    /addresses/{id}                 # Update address
DELETE /addresses/{id}                 # Delete address
POST   /addresses/{id}/default         # Set as default

Product Catalog:
GET    /products                       # List products with filters
GET    /products/{id}                  # Get product details
GET    /categories                     # List categories
GET    /search                         # Search products

Shopping Cart:
GET    /cart                          # Get cart contents
POST   /cart/items                    # Add item to cart
PUT    /cart/items/{id}               # Update cart item
DELETE /cart/items/{id}               # Remove cart item
DELETE /cart                          # Clear cart

Checkout & Orders:
POST   /checkout                      # Initiate checkout
GET    /shipping-methods              # Get shipping options
POST   /orders                        # Create order
GET    /orders                        # List customer orders
GET    /orders/{id}                   # Get order details
POST   /orders/{id}/cancel            # Cancel order
GET    /orders/{number}/track         # Track order
```

### **Response Format Standards**

#### **Success Response**
```go
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Message   string      `json:"message,omitempty"`
    RequestID string      `json:"request_id"`
    Timestamp time.Time   `json:"timestamp"`
}
```

#### **Error Response**
```go
type APIError struct {
    Success      bool                   `json:"success"`
    Error        ErrorDetail            `json:"error"`
    ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
    RequestID    string                 `json:"request_id"`
    Timestamp    time.Time              `json:"timestamp"`
}

type ErrorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

#### **Pagination Response**
```go
type PaginatedResponse struct {
    Success    bool        `json:"success"`
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
    RequestID  string      `json:"request_id"`
    Timestamp  time.Time   `json:"timestamp"`
}

type Pagination struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalItems int `json:"total_items"`
    TotalPages int `json:"total_pages"`
    HasNext    bool `json:"has_next"`
    HasPrev    bool `json:"has_prev"`
}
```

---

## 🔒 **Security Architecture**

### **Authentication & Authorization**

#### **JWT Token Structure**

**Seller Token (Existing)**:
```go
type SellerClaims struct {
    UserID   string     `json:"user_id"`
    UserRole entity.UserRole `json:"user_role"`
    UserTier entity.UserTier `json:"user_tier"`
    Scope    []string   `json:"scope"` // ["seller:read", "seller:write"]
    jwt.RegisteredClaims
}
```

**Customer Token (New)**:
```go
type CustomerClaims struct {
    CustomerID   string `json:"customer_id"`
    StorefrontID string `json:"storefront_id"`
    Email        string `json:"email"`
    Scope        []string `json:"scope"` // ["customer:read", "customer:write"]
    jwt.RegisteredClaims
}
```

#### **Multi-Tenant Security Middleware**

```go
// Middleware for storefront-scoped operations
func StorefrontMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        slug := c.Param("slug")
        storefront, err := storefrontService.GetBySlug(c, slug)
        if err != nil {
            c.JSON(404, gin.H{"error": "Storefront not found"})
            c.Abort()
            return
        }
        c.Set("storefront", storefront)
        c.Next()
    }
}

// Middleware for customer authentication
func CustomerAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c.GetHeader("Authorization"))
        claims, err := validateCustomerToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Verify customer belongs to this storefront
        storefront := c.MustGet("storefront").(*entity.Storefront)
        if claims.StorefrontID != storefront.ID.String() {
            c.JSON(403, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        c.Set("customer_id", claims.CustomerID)
        c.Set("customer_claims", claims)
        c.Next()
    }
}

// Optional customer auth for guest users
func OptionalCustomerAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c.GetHeader("Authorization"))
        if token != "" {
            if claims, err := validateCustomerToken(token); err == nil {
                storefront := c.MustGet("storefront").(*entity.Storefront)
                if claims.StorefrontID == storefront.ID.String() {
                    c.Set("customer_id", claims.CustomerID)
                    c.Set("customer_claims", claims)
                }
            }
        }
        c.Next()
    }
}
```

### **Data Isolation Strategy**

#### **Database Level Isolation**
```sql
-- All customer data includes storefront_id for isolation
CREATE POLICY customer_isolation ON customers 
    FOR ALL USING (storefront_id = current_setting('app.current_storefront_id')::UUID);

-- All orders include storefront_id for isolation
CREATE POLICY order_isolation ON orders 
    FOR ALL USING (storefront_id = current_setting('app.current_storefront_id')::UUID);

-- Enable RLS on all multi-tenant tables
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE shopping_carts ENABLE ROW LEVEL SECURITY;
```

#### **Application Level Isolation**
```go
// Repository methods always include storefront filtering
func (r *customerRepository) GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error) {
    query := `
        SELECT * FROM customers 
        WHERE storefront_id = $1 AND email = $2 AND deleted_at IS NULL
    `
    // Implementation...
}

// Use case methods validate storefront access
func (uc *customerUseCase) GetCustomer(ctx context.Context, customerID uuid.UUID) (*dto.CustomerResponse, error) {
    // Validate storefront access from context
    storefrontID := uc.getStorefrontIDFromContext(ctx)
    customer, err := uc.customerRepo.GetByID(ctx, customerID)
    if err != nil {
        return nil, err
    }
    if customer.StorefrontID != storefrontID {
        return nil, ErrAccessDenied
    }
    // Return customer data
}
```

---

## 📊 **Performance & Scalability**

### **Caching Strategy**

#### **Redis Cache Structure**
```
Storefront Data:
- storefront:{slug} -> StorefrontDetails (TTL: 1 hour)
- storefront_config:{storefront_id} -> StorefrontConfig (TTL: 1 hour)

Customer Sessions:
- customer_session:{token} -> CustomerClaims (TTL: 24 hours)
- customer_cart:{storefront_id}:{customer_id} -> CartDetails (TTL: 7 days)
- guest_cart:{session_id} -> CartDetails (TTL: 24 hours)

Product Catalog:
- storefront_products:{storefront_id}:{page}:{filters} -> ProductList (TTL: 30 minutes)
- product_details:{product_id} -> ProductDetails (TTL: 1 hour)

Search Results:
- search:{storefront_id}:{query}:{filters} -> SearchResults (TTL: 15 minutes)
```

#### **Database Optimization**

**Indexes Strategy**:
```sql
-- Storefront indexes
CREATE INDEX CONCURRENTLY idx_storefronts_slug ON storefronts(slug) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY idx_storefronts_seller_status ON storefronts(seller_id, status) WHERE deleted_at IS NULL;

-- Customer indexes  
CREATE INDEX CONCURRENTLY idx_customers_storefront_email ON customers(storefront_id, email) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY idx_customers_storefront_phone ON customers(storefront_id, phone) WHERE deleted_at IS NULL;

-- Order indexes
CREATE INDEX CONCURRENTLY idx_orders_storefront_status ON orders(storefront_id, status, created_at DESC);
CREATE INDEX CONCURRENTLY idx_orders_customer_created ON orders(customer_id, created_at DESC);
CREATE INDEX CONCURRENTLY idx_orders_order_number ON orders(order_number) WHERE order_number IS NOT NULL;

-- Shopping cart indexes
CREATE INDEX CONCURRENTLY idx_shopping_carts_customer ON shopping_carts(storefront_id, customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_shopping_carts_session ON shopping_carts(storefront_id, session_id) WHERE session_id IS NOT NULL;
```

**Query Optimization**:
```sql
-- Optimized customer order history query
SELECT o.*, 
       COUNT(*) OVER() as total_count,
       jsonb_agg(
           jsonb_build_object(
               'product_name', oi.product_name,
               'quantity', oi.quantity,
               'unit_price', oi.unit_price
           )
       ) as items
FROM orders o
LEFT JOIN order_items oi ON oi.order_id = o.id
WHERE o.storefront_id = $1 AND o.customer_id = $2
GROUP BY o.id
ORDER BY o.created_at DESC
LIMIT $3 OFFSET $4;
```

### **API Rate Limiting**

```go
// Rate limiting configuration
type RateLimitConfig struct {
    PublicAPI     RateLimit // Guest users: 100 req/minute
    CustomerAPI   RateLimit // Authenticated customers: 300 req/minute  
    SellerAPI     RateLimit // Sellers: 1000 req/minute
    CheckoutAPI   RateLimit // Checkout flow: 10 req/minute per customer
}

// Implementation using Redis
func RateLimitMiddleware(config RateLimit) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := generateRateLimitKey(c)
        allowed, err := checkRateLimit(key, config)
        if err != nil || !allowed {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## 🚀 **Deployment Architecture**

### **Container Architecture**
```yaml
# docker-compose.yml
services:
  smartseller-api:
    image: smartseller/backend:latest
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
    ports:
      - "8080:8080"
    
  postgresql:
    image: postgres:15
    environment:
      POSTGRES_DB: smartseller
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  redis:
    image: redis:7
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
```

### **Kubernetes Deployment**
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: smartseller-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: smartseller-backend
  template:
    metadata:
      labels:
        app: smartseller-backend
    spec:
      containers:
      - name: backend
        image: smartseller/backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: smartseller-secrets
              key: database-url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

## 📈 **Monitoring & Observability**

### **Metrics Collection**
```go
// Custom metrics for storefront operations
var (
    storefrontRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "storefront_requests_total",
            Help: "Total number of storefront requests",
        },
        []string{"storefront_id", "method", "endpoint", "status"},
    )
    
    orderProcessingTime = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "order_processing_duration_seconds",
            Help: "Time taken to process orders",
        },
        []string{"storefront_id", "payment_method"},
    )
    
    cartOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cart_operations_total",
            Help: "Total cart operations",
        },
        []string{"storefront_id", "operation", "customer_type"}, // guest vs registered
    )
)
```

### **Health Checks**
```go
type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

func (h *HealthChecker) CheckHealth() HealthStatus {
    status := HealthStatus{
        Status: "healthy",
        Timestamp: time.Now(),
        Services: make(map[string]ServiceHealth),
    }
    
    // Check database
    if err := h.db.Ping(); err != nil {
        status.Services["database"] = ServiceHealth{Status: "unhealthy", Error: err.Error()}
        status.Status = "unhealthy"
    } else {
        status.Services["database"] = ServiceHealth{Status: "healthy"}
    }
    
    // Check Redis
    if _, err := h.redis.Ping(context.Background()).Result(); err != nil {
        status.Services["redis"] = ServiceHealth{Status: "unhealthy", Error: err.Error()}
        status.Status = "degraded" // Redis failure is not critical
    } else {
        status.Services["redis"] = ServiceHealth{Status: "healthy"}
    }
    
    return status
}
```

---

## 📝 **Next Steps**

### **Development Priorities**

1. **Phase 1**: Database schema and core entities
2. **Phase 2**: Repository and use case implementations
3. **Phase 3**: API handlers and middleware
4. **Phase 4**: Authentication and security
5. **Phase 5**: Integration testing and optimization
6. **Phase 6**: Documentation and deployment

### **Integration Points**

- **Frontend Development**: React/Next.js storefront templates
- **Payment Gateways**: Midtrans, Xendit, Stripe integration
- **Shipping Providers**: JNE, J&T, SiCepat integration
- **Email Service**: Template system for order confirmations
- **SMS Service**: Order status notifications
- **Analytics**: Customer behavior and sales analytics

---

**Document Status**: Draft - Technical Review Required  
**Last Updated**: September 26, 2025  
**Next Review**: Implementation Phase 1 Completion