# SmartSeller Storefront & Customer Management - PRD

## 📋 **Document Overview**

**Product**: SmartSeller Storefront & Customer Management System  
**Version**: 1.0  
**Created**: September 26, 2025  
**Status**: Draft  
**Owner**: SmartSeller Development Team  

---

## 🎯 **Executive Summary**

This Product Requirements Document (PRD) defines the requirements for implementing a comprehensive storefront and customer management system for the SmartSeller platform. This will enable SmartSeller users (sellers) to create customer-facing online stores where end customers can browse products, register accounts, and make purchases - similar to Shopify's storefront functionality.

### **Key Objectives**
- Enable sellers to create and customize their own storefronts
- Allow end customers to register, browse, and purchase products
- Provide a seamless B2B2C e-commerce experience
- Maintain multi-tenant architecture with proper data isolation
- Support scalable order processing and customer management

---

## 🏗️ **Current State Analysis**

### **Existing Foundation**
✅ **User Management**: Seller authentication and authorization system  
✅ **Product Management**: Complete product catalog with variants, categories, and inventory  
✅ **API Infrastructure**: RESTful API with authentication middleware  
✅ **Database**: PostgreSQL with proper entity relationships  
✅ **Clean Architecture**: Domain-driven design with repository pattern  

### **Gaps to Address**
❌ **Customer Entity**: Separate customer model for end-users  
❌ **Storefront Management**: Store configuration and customization  
❌ **Shopping Cart**: Session-based cart functionality  
❌ **Order Management**: Order creation, processing, and tracking  
❌ **Payment Integration**: Payment gateway integration  
❌ **Customer Authentication**: Separate auth system for customers  

---

## 🎯 **Core Features & Requirements**

### **F1: Multi-Tenant Storefront Management**
**Description**: Enable sellers to create and manage their own storefronts

**Requirements**:
- ✅ Each seller can create one or more storefronts
- ✅ Customizable storefront branding (logo, colors, theme)
- ✅ Custom domain and subdomain support
- ✅ SEO configuration (meta tags, sitemap, robots.txt)
- ✅ Store status management (active, inactive, maintenance)
- ✅ Multi-language support
- ✅ Mobile-responsive design

**Success Criteria**:
- Sellers can set up a storefront in under 10 minutes
- Storefronts load in under 3 seconds
- Support for unlimited products per storefront

### **F2: Customer Registration & Authentication**
**Description**: Separate authentication system for end customers

**Requirements**:
- ✅ Customer registration with email/phone verification
- ✅ Social login integration (Google, Facebook)
- ✅ Password reset and account recovery
- ✅ Customer profile management
- ✅ Address book management
- ✅ Order history and tracking
- ✅ Wishlist functionality

**Success Criteria**:
- Registration process completion rate >80%
- Social login integration working
- Customer data properly isolated per storefront

### **F3: Product Catalog Frontend**
**Description**: Customer-facing product browsing and search

**Requirements**:
- ✅ Product listing with pagination and filtering
- ✅ Product detail pages with variants and images
- ✅ Category-based navigation
- ✅ Search functionality with autocomplete
- ✅ Product comparison feature
- ✅ Recently viewed products
- ✅ Product recommendations

**Success Criteria**:
- Search results return in under 1 second
- Support for all product variants and options
- Mobile-optimized product pages

### **F4: Shopping Cart & Checkout**
**Description**: Complete shopping cart and checkout flow

**Requirements**:
- ✅ Session-based cart for guest users
- ✅ Persistent cart for registered users
- ✅ Cart management (add, remove, update quantities)
- ✅ Shipping calculation integration
- ✅ Tax calculation support
- ✅ Discount code application
- ✅ Multiple payment method support
- ✅ Guest checkout option

**Success Criteria**:
- Cart abandonment rate <70%
- Checkout completion in under 3 minutes
- Support for multiple shipping options

### **F5: Order Management**
**Description**: Complete order lifecycle management

**Requirements**:
- ✅ Order creation and validation
- ✅ Inventory deduction and reservation
- ✅ Order status tracking
- ✅ Email notifications at each stage
- ✅ Invoice generation
- ✅ Return and refund processing
- ✅ Order history for customers

**Success Criteria**:
- Orders processed within 1 minute of payment
- Real-time inventory updates
- Automated email notifications

### **F6: Customer Management Dashboard**
**Description**: Seller-facing customer management tools

**Requirements**:
- ✅ Customer listing with search and filters
- ✅ Customer profile views with order history
- ✅ Customer segmentation tools
- ✅ Communication history tracking
- ✅ Customer analytics and insights
- ✅ Bulk email and SMS marketing
- ✅ Customer support ticket system

**Success Criteria**:
- 360-degree customer view for sellers
- Segmentation and targeting capabilities
- Integrated communication tools

---

## 🏛️ **System Architecture**

### **Multi-Tenant Data Model**

```
SmartSeller User (Seller)
├── Storefront (1:N)
│   ├── StorefrontConfig (1:1)
│   ├── StorefrontTheme (1:1)
│   └── Customer (1:N)
│       ├── CustomerProfile (1:1)
│       ├── CustomerAddress (1:N)
│       ├── ShoppingCart (1:1)
│       └── Order (1:N)
│           ├── OrderItem (1:N)
│           ├── OrderShipping (1:1)
│           └── OrderPayment (1:N)
└── Product (1:N) [Shared across storefronts]
```

### **Domain Boundaries**

#### **Seller Domain** (Existing)
- User authentication and management
- Product catalog management
- Business analytics and reporting

#### **Storefront Domain** (New)
- Storefront configuration and branding
- Theme and layout management
- Domain and SEO management

#### **Customer Domain** (New)
- Customer authentication and profiles
- Address and preference management
- Wishlist and favorites

#### **Commerce Domain** (New)
- Shopping cart management
- Order processing and fulfillment
- Payment and shipping integration

#### **Communication Domain** (New)
- Email and SMS notifications
- Customer support ticketing
- Marketing campaigns

---

## 🗄️ **Database Schema Design**

### **New Entities**

#### **Storefront**
```sql
CREATE TABLE storefronts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    domain VARCHAR(255),
    subdomain VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    currency VARCHAR(3) DEFAULT 'IDR',
    language VARCHAR(5) DEFAULT 'id',
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

#### **StorefrontConfig**
```sql
CREATE TABLE storefront_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    logo_url VARCHAR(500),
    favicon_url VARCHAR(500),
    primary_color VARCHAR(7),
    secondary_color VARCHAR(7),
    font_family VARCHAR(100),
    meta_title VARCHAR(255),
    meta_description TEXT,
    social_media_links JSONB,
    contact_info JSONB,
    business_hours JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### **Customer**
```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    password_hash VARCHAR(255),
    password_salt VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    date_of_birth DATE,
    gender VARCHAR(10),
    email_verified_at TIMESTAMPTZ,
    phone_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    customer_group VARCHAR(50) DEFAULT 'regular',
    total_orders INTEGER DEFAULT 0,
    total_spent DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(storefront_id, email)
);
```

#### **CustomerAddress**
```sql
CREATE TABLE customer_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    label VARCHAR(50) NOT NULL,
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
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### **ShoppingCart**
```sql
CREATE TABLE shopping_carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    customer_id UUID REFERENCES customers(id),
    session_id VARCHAR(255),
    currency VARCHAR(3) DEFAULT 'IDR',
    subtotal DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    shipping_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### **CartItem**
```sql
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES shopping_carts(id),
    product_id UUID NOT NULL REFERENCES products(id),
    product_variant_id UUID REFERENCES product_variants(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    total_price DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(cart_id, product_id, product_variant_id)
);
```

#### **Order**
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    customer_id UUID REFERENCES customers(id),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    payment_status VARCHAR(50) DEFAULT 'pending',
    fulfillment_status VARCHAR(50) DEFAULT 'unfulfilled',
    
    -- Customer information (snapshot)
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(20),
    customer_first_name VARCHAR(100),
    customer_last_name VARCHAR(100),
    
    -- Billing address (snapshot)
    billing_address JSONB NOT NULL,
    
    -- Shipping address (snapshot)
    shipping_address JSONB NOT NULL,
    
    -- Financial information
    currency VARCHAR(3) DEFAULT 'IDR',
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    shipping_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    
    -- Additional information
    notes TEXT,
    shipping_method VARCHAR(100),
    tracking_number VARCHAR(100),
    
    -- Timestamps
    ordered_at TIMESTAMPTZ DEFAULT NOW(),
    shipped_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### **OrderItem**
```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    product_id UUID NOT NULL REFERENCES products(id),
    product_variant_id UUID REFERENCES product_variants(id),
    
    -- Product information (snapshot)
    product_name VARCHAR(255) NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    variant_name VARCHAR(255),
    
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    total_price DECIMAL(15,2) NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### **Indexes and Performance**
```sql
-- Storefront indexes
CREATE INDEX idx_storefronts_seller_id ON storefronts(seller_id);
CREATE INDEX idx_storefronts_slug ON storefronts(slug);
CREATE INDEX idx_storefronts_status ON storefronts(status);

-- Customer indexes
CREATE INDEX idx_customers_storefront_id ON customers(storefront_id);
CREATE INDEX idx_customers_email ON customers(storefront_id, email);
CREATE INDEX idx_customers_phone ON customers(storefront_id, phone);
CREATE INDEX idx_customers_status ON customers(status);

-- Cart indexes
CREATE INDEX idx_shopping_carts_storefront_id ON shopping_carts(storefront_id);
CREATE INDEX idx_shopping_carts_customer_id ON shopping_carts(customer_id);
CREATE INDEX idx_shopping_carts_session_id ON shopping_carts(session_id);

-- Order indexes
CREATE INDEX idx_orders_storefront_id ON orders(storefront_id);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
```

---

## 🔧 **API Design**

### **Storefront Management API**

#### **Seller-facing Endpoints**
```
GET    /api/v1/storefronts              # List seller's storefronts
POST   /api/v1/storefronts              # Create new storefront
GET    /api/v1/storefronts/{id}         # Get storefront details
PUT    /api/v1/storefronts/{id}         # Update storefront
DELETE /api/v1/storefronts/{id}         # Delete storefront

GET    /api/v1/storefronts/{id}/config  # Get storefront configuration
PUT    /api/v1/storefronts/{id}/config  # Update storefront configuration

GET    /api/v1/storefronts/{id}/customers    # List storefront customers
GET    /api/v1/storefronts/{id}/orders       # List storefront orders
GET    /api/v1/storefronts/{id}/analytics    # Get storefront analytics
```

### **Customer-facing API (Public)**

#### **Authentication**
```
POST   /api/storefront/{slug}/auth/register     # Customer registration
POST   /api/storefront/{slug}/auth/login        # Customer login
POST   /api/storefront/{slug}/auth/logout       # Customer logout
POST   /api/storefront/{slug}/auth/refresh      # Refresh token
POST   /api/storefront/{slug}/auth/forgot-password  # Password reset
```

#### **Customer Profile**
```
GET    /api/storefront/{slug}/profile           # Get customer profile
PUT    /api/storefront/{slug}/profile           # Update customer profile
GET    /api/storefront/{slug}/addresses         # List customer addresses
POST   /api/storefront/{slug}/addresses         # Add new address
PUT    /api/storefront/{slug}/addresses/{id}    # Update address
DELETE /api/storefront/{slug}/addresses/{id}    # Delete address
```

#### **Product Catalog**
```
GET    /api/storefront/{slug}/products          # List products
GET    /api/storefront/{slug}/products/{id}     # Get product details
GET    /api/storefront/{slug}/categories        # List categories
GET    /api/storefront/{slug}/search            # Search products
```

#### **Shopping Cart**
```
GET    /api/storefront/{slug}/cart              # Get cart contents
POST   /api/storefront/{slug}/cart/items        # Add item to cart
PUT    /api/storefront/{slug}/cart/items/{id}   # Update cart item
DELETE /api/storefront/{slug}/cart/items/{id}   # Remove cart item
POST   /api/storefront/{slug}/cart/clear        # Clear entire cart
```

#### **Checkout & Orders**
```
POST   /api/storefront/{slug}/checkout          # Initiate checkout
GET    /api/storefront/{slug}/shipping-methods  # Get shipping options
POST   /api/storefront/{slug}/orders            # Create order
GET    /api/storefront/{slug}/orders            # List customer orders
GET    /api/storefront/{slug}/orders/{id}       # Get order details
POST   /api/storefront/{slug}/orders/{id}/cancel  # Cancel order
```

---

## 🔒 **Security & Authentication**

### **Multi-Tenant Security**
- **Data Isolation**: All customer data scoped to specific storefronts
- **JWT Tokens**: Separate token scopes for sellers vs customers
- **API Rate Limiting**: Per-storefront and per-customer limits
- **CORS Configuration**: Storefront-specific domain restrictions

### **Customer Authentication**
```go
// JWT Claims for customers
type CustomerClaims struct {
    CustomerID   string `json:"customer_id"`
    StorefrontID string `json:"storefront_id"`
    Email        string `json:"email"`
    Role         string `json:"role"` // "customer"
    jwt.RegisteredClaims
}
```

### **Authorization Middleware**
```go
// Middleware for storefront-scoped operations
func StorefrontAuthMiddleware() gin.HandlerFunc
func CustomerAuthMiddleware() gin.HandlerFunc
func OptionalCustomerAuthMiddleware() gin.HandlerFunc // For guest users
```

---

## 📊 **Implementation Timeline**

### **Phase 1: Foundation (Week 1-2)**
**Duration**: 10 days  
**Dependencies**: Existing product management system  

**Deliverables**:
- Database schema and migrations
- Core entity models (Storefront, Customer)
- Basic repository implementations
- Multi-tenant security middleware

**Success Criteria**:
- All database tables created and indexed
- Basic CRUD operations working
- Seller can create storefronts

### **Phase 2: Storefront Management (Week 3)**  
**Duration**: 7 days  
**Dependencies**: Phase 1  

**Deliverables**:
- Storefront configuration management
- Theme and branding system
- Seller-facing storefront APIs
- Admin dashboard integration

**Success Criteria**:
- Sellers can configure storefronts
- Storefront branding working
- API endpoints fully tested

### **Phase 3: Customer Authentication (Week 4)**
**Duration**: 7 days  
**Dependencies**: Phase 1  

**Deliverables**:
- Customer registration and login
- Profile and address management
- Email verification system
- Customer-facing APIs

**Success Criteria**:
- Customer registration flow working
- Authentication system secure
- Profile management functional

### **Phase 4: Shopping Cart (Week 5)**
**Duration**: 7 days  
**Dependencies**: Phase 2, 3  

**Deliverables**:
- Cart entity and management
- Session-based and persistent carts
- Cart synchronization logic
- Cart API endpoints

**Success Criteria**:
- Add/remove items working
- Cart persistence across sessions
- Guest and customer carts merged

### **Phase 5: Order Processing (Week 6-7)**
**Duration**: 10 days  
**Dependencies**: Phase 4  

**Deliverables**:
- Order creation and management
- Checkout flow implementation
- Inventory integration
- Email notification system

**Success Criteria**:
- Complete checkout flow
- Orders properly created
- Inventory automatically updated

### **Phase 6: Integration & Testing (Week 8)**
**Duration**: 7 days  
**Dependencies**: All previous phases  

**Deliverables**:
- End-to-end testing
- Performance optimization
- Documentation updates
- Deployment preparation

**Success Criteria**:
- All tests passing
- Performance meets SLAs
- Ready for production deployment

---

## 🎯 **Success Metrics**

### **Technical Metrics**
- **API Response Time**: <200ms for 95th percentile
- **Database Query Performance**: <50ms average
- **Storefront Load Time**: <3 seconds
- **System Uptime**: 99.9% availability

### **Business Metrics**
- **Storefront Creation Rate**: Sellers can create storefronts in <10 minutes
- **Customer Registration Rate**: >80% completion rate
- **Cart Abandonment Rate**: <70% (industry average)
- **Order Processing Time**: <1 minute from payment to confirmation

### **User Experience Metrics**
- **Mobile Responsiveness**: All pages mobile-optimized
- **Search Performance**: Results in <1 second
- **Checkout Flow**: Completion in <3 minutes
- **Customer Satisfaction**: >4.0/5.0 rating

---

## 🔄 **Future Enhancements**

### **Phase 2 Features**
- Advanced analytics and reporting
- Customer segmentation and marketing tools
- Multi-language and multi-currency support
- Advanced shipping integrations

### **Phase 3 Features**
- Mobile app for customers
- Loyalty program integration
- Social commerce features
- AI-powered recommendations

### **Phase 4 Features**
- Marketplace integrations
- Advanced customization tools
- Third-party app ecosystem
- White-label solutions

---

## 📝 **Conclusion**

This PRD provides a comprehensive roadmap for implementing storefront and customer management capabilities in the SmartSeller platform. The proposed solution leverages the existing product management foundation while introducing new customer-facing capabilities that will enable sellers to create professional e-commerce storefronts similar to Shopify.

The multi-tenant architecture ensures data isolation and scalability, while the phased implementation approach allows for incremental delivery and testing. The end result will be a powerful B2B2C platform that empowers sellers to grow their online businesses effectively.

**Next Steps**:
1. Review and approve this PRD
2. Create technical architecture document
3. Begin Phase 1 implementation
4. Set up testing and monitoring infrastructure
5. Plan frontend development roadmap

---

**Document Status**: Draft - Pending Review  
**Last Updated**: September 26, 2025  
**Next Review**: TBD