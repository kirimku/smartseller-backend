# SmartSeller Product Management System - PRD

## 📋 Product Requirements Document

### 🎯 **Overview**
SmartSeller Product Management System is a comprehensive e-commerce solution that enables sellers to manage products with advanced variant support, dynamic pricing, inventory tracking, and multi-media capabilities.

---

## 🏗️ **Technical Architecture Analysis**

### **Database Schema Overview**
- **Primary Tables**: 5 core tables with advanced relationships
- **Advanced Features**: JSONB variant options, automatic triggers, validation functions
- **Scalability**: UUID primary keys, GIN indexes, soft deletes

#### **Table Structure**
1. **`products`** - Core product information
2. **`product_categories`** - Hierarchical categorization  
3. **`product_images`** - Multi-media management
4. **`product_variant_options`** - Dynamic option definitions
5. **`product_variants`** - Specific variant combinations

---

## 🎯 **Core Features & Requirements**

### **F1: Product Management**
**Description**: Complete CRUD operations for products with business rules validation

**Requirements**:
- ✅ Create products with SKU auto-generation
- ✅ Update product information with audit trail
- ✅ Soft delete with restoration capability
- ✅ Bulk operations (import/export)
- ✅ Status workflow (draft → active → inactive → archived)

### **F2: Dynamic Variant System**
**Description**: Flexible product variant management with dynamic options

**Requirements**:
- ✅ Define variant options per product (Color, Size, Material, etc.)
- ✅ Create variants with specific combinations
- ✅ Auto-generate variant names from options
- ✅ Validate variant option combinations
- ✅ Individual pricing and inventory per variant

### **F3: Hierarchical Categories**
**Description**: Multi-level category system with inheritance

**Requirements**:
- ✅ Create parent/child category relationships
- ✅ Category-based product organization
- ✅ SEO-friendly slugs
- ✅ Category activation/deactivation

### **F4: Inventory Management**
**Description**: Real-time inventory tracking with alerts

**Requirements**:
- ✅ Stock quantity tracking per product/variant
- ✅ Low stock threshold alerts
- ✅ Inventory adjustments with reason codes
- ✅ Stock movement history

### **F5: Pricing Management**
**Description**: Flexible pricing with cost tracking and sales

**Requirements**:
- ✅ Base price, sale price, cost price tracking
- ✅ Variant-specific pricing overrides
- ✅ Profit margin calculations
- ✅ Price history tracking

### **F6: Multi-Media Management**
**Description**: Product image and media handling

**Requirements**:
- ✅ Multiple images per product
- ✅ Primary image designation
- ✅ Alt text for accessibility
- ✅ Image sorting and organization
- ✅ Variant-specific images

### **F7: SEO & Marketing**
**Description**: Search engine optimization and marketing features

**Requirements**:
- ✅ SEO-friendly slugs
- ✅ Meta titles and descriptions
- ✅ Tag-based categorization
- ✅ Product search optimization

---

## 🔧 **Technical Implementation Breakdown**

### **Phase 1: Core Entities & Repository Layer**
**Duration**: 2-3 days  
**Dependencies**: Database migrations completed

#### **Task 1.1: Product Entity Development**
**File**: `internal/domain/entity/product.go`
```go
// Key structures to implement:
- Product struct with all fields
- ProductStatus enum (draft, active, inactive, archived)
- Validation methods
- Business logic methods
```

**Acceptance Criteria**:
- ✅ Complete Product entity with all database fields
- ✅ Validation rules for SKU format, pricing, dimensions
- ✅ Status transition logic
- ✅ Soft delete support

#### **Task 1.2: Product Category Entity**
**File**: `internal/domain/entity/product_category.go`
```go
// Key structures:
- ProductCategory struct
- Hierarchical relationship methods
- Slug generation and validation
```

#### **Task 1.3: Product Variant System Entities**
**Files**: 
- `internal/domain/entity/product_variant.go`
- `internal/domain/entity/product_variant_option.go`

```go
// Dynamic variant system:
- ProductVariantOption (defines available options)
- ProductVariant (specific combinations)
- JSONB option validation
- Auto-naming logic
```

#### **Task 1.4: Product Image Entity**
**File**: `internal/domain/entity/product_image.go`
```go
// Media management:
- ProductImage struct
- Image sorting and primary designation
- URL validation
```

#### **Task 1.5: Repository Interfaces**
**File**: `internal/domain/repository/product_repository.go`
```go
// Repository contracts:
- ProductRepository interface
- ProductCategoryRepository interface  
- ProductVariantRepository interface
- ProductImageRepository interface
```

---

### **Phase 2: Repository Implementation**
**Duration**: 3-4 days  
**Dependencies**: Phase 1 completed

#### **Task 2.1: Product Repository Implementation**
**File**: `internal/infrastructure/repository/product_repository.go`

**Key Methods**:
```go
// Core CRUD
- CreateProduct(ctx context.Context, product *entity.Product) error
- GetProductByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
- GetProductBySKU(ctx context.Context, sku string) (*entity.Product, error)
- UpdateProduct(ctx context.Context, product *entity.Product) error
- DeleteProduct(ctx context.Context, id uuid.UUID) error
- RestoreProduct(ctx context.Context, id uuid.UUID) error

// Business queries
- GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*entity.Product, error)
- GetProductsByStatus(ctx context.Context, status entity.ProductStatus, limit, offset int) ([]*entity.Product, error)
- GetProductsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Product, error)
- SearchProducts(ctx context.Context, query string, filters ProductFilters) ([]*entity.Product, error)
- GetLowStockProducts(ctx context.Context, userID uuid.UUID) ([]*entity.Product, error)

// Bulk operations
- BulkCreateProducts(ctx context.Context, products []*entity.Product) error
- BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status entity.ProductStatus) error
```

#### **Task 2.2: Product Category Repository**
**File**: `internal/infrastructure/repository/product_category_repository.go`

**Key Methods**:
```go
// Hierarchical operations
- CreateCategory(ctx context.Context, category *entity.ProductCategory) error
- GetCategoryHierarchy(ctx context.Context, parentID *uuid.UUID) ([]*entity.ProductCategory, error)
- GetCategoryPath(ctx context.Context, categoryID uuid.UUID) ([]*entity.ProductCategory, error)
- MoveCategoryToParent(ctx context.Context, categoryID, newParentID uuid.UUID) error
```

#### **Task 2.3: Product Variant Repository**
**File**: `internal/infrastructure/repository/product_variant_repository.go`

**Key Methods**:
```go
// Variant management
- CreateVariantOptions(ctx context.Context, productID uuid.UUID, options []*entity.ProductVariantOption) error
- CreateVariant(ctx context.Context, variant *entity.ProductVariant) error
- GetProductVariants(ctx context.Context, productID uuid.UUID) ([]*entity.ProductVariant, error)
- GetVariantBySKU(ctx context.Context, sku string) (*entity.ProductVariant, error)
- ValidateVariantOptions(ctx context.Context, productID uuid.UUID, options map[string]string) error
```

---

### **Phase 3: Use Case Layer**
**Duration**: 2-3 days  
**Dependencies**: Phase 2 completed

#### **Task 3.1: Product Use Cases**
**File**: `internal/application/usecase/product_usecase.go`

**Key Use Cases**:
```go
// Product lifecycle
- CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
- UpdateProduct(ctx context.Context, id uuid.UUID, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
- GetProduct(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error)
- ListProducts(ctx context.Context, req *dto.ListProductsRequest) (*dto.ListProductsResponse, error)
- DeleteProduct(ctx context.Context, id uuid.UUID) error

// Business operations
- ActivateProduct(ctx context.Context, id uuid.UUID) error
- DeactivateProduct(ctx context.Context, id uuid.UUID) error
- ArchiveProduct(ctx context.Context, id uuid.UUID) error
- DuplicateProduct(ctx context.Context, id uuid.UUID, newSKU string) (*dto.ProductResponse, error)

// Inventory operations
- UpdateStock(ctx context.Context, id uuid.UUID, quantity int, reason string) error
- GetLowStockAlerts(ctx context.Context, userID uuid.UUID) ([]*dto.LowStockAlert, error)
```

#### **Task 3.2: Product Variant Use Cases**
**File**: `internal/application/usecase/product_variant_usecase.go`

#### **Task 3.3: Product Category Use Cases**
**File**: `internal/application/usecase/product_category_usecase.go`

---

### **Phase 4: DTOs & Validation**
**Duration**: 1-2 days  
**Dependencies**: Phase 3 completed

#### **Task 4.1: Product DTOs**
**File**: `internal/application/dto/product_dto.go`

**Key DTOs**:
```go
// Request DTOs
- CreateProductRequest
- UpdateProductRequest  
- ListProductsRequest
- ProductFilters

// Response DTOs
- ProductResponse
- ProductListResponse
- ProductSummary
```

#### **Task 4.2: Validation Rules**
**File**: `internal/application/dto/product_validation.go`

```go
// Validation functions
- ValidateCreateProductRequest(req *CreateProductRequest) *ValidationErrors
- ValidateSKUFormat(sku string) error
- ValidatePricing(basePrice, salePrice, costPrice decimal.Decimal) error
- ValidateDimensions(length, width, height decimal.Decimal) error
```

---

### **Phase 5: HTTP API Layer**
**Duration**: 2-3 days  
**Dependencies**: Phase 4 completed

#### **Task 5.1: Product Handlers**
**File**: `internal/interfaces/api/handler/product_handler.go`

**Endpoints**:
```go
// Product CRUD
- CreateProduct(c *gin.Context)      // POST /api/v1/products
- GetProduct(c *gin.Context)         // GET /api/v1/products/:id
- UpdateProduct(c *gin.Context)      // PUT /api/v1/products/:id
- DeleteProduct(c *gin.Context)      // DELETE /api/v1/products/:id
- ListProducts(c *gin.Context)       // GET /api/v1/products

// Product operations
- ActivateProduct(c *gin.Context)    // PUT /api/v1/products/:id/activate
- DeactivateProduct(c *gin.Context)  // PUT /api/v1/products/:id/deactivate
- ArchiveProduct(c *gin.Context)     // PUT /api/v1/products/:id/archive
- DuplicateProduct(c *gin.Context)   // POST /api/v1/products/:id/duplicate

// Inventory operations
- UpdateStock(c *gin.Context)        // PUT /api/v1/products/:id/stock
- GetLowStock(c *gin.Context)        // GET /api/v1/products/low-stock
```

#### **Task 5.2: Product Category Handlers**
**File**: `internal/interfaces/api/handler/product_category_handler.go`

#### **Task 5.3: Product Variant Handlers**
**File**: `internal/interfaces/api/handler/product_variant_handler.go`

---

### **Phase 6: API Routes & Integration**
**Duration**: 1 day  
**Dependencies**: Phase 5 completed

#### **Task 6.1: Route Registration**
**File**: `internal/interfaces/api/router/router.go`

```go
// Add to setupAPIRoutes()
products := v1.Group("/products")
products.Use(middleware.AuthMiddleware())
{
    // Product CRUD routes
    products.POST("", productHandler.CreateProduct)
    products.GET("", productHandler.ListProducts)
    products.GET("/:id", productHandler.GetProduct)
    products.PUT("/:id", productHandler.UpdateProduct)
    products.DELETE("/:id", productHandler.DeleteProduct)
    
    // Product operations
    products.PUT("/:id/activate", productHandler.ActivateProduct)
    products.PUT("/:id/deactivate", productHandler.DeactivateProduct)
    products.PUT("/:id/archive", productHandler.ArchiveProduct)
    products.POST("/:id/duplicate", productHandler.DuplicateProduct)
    
    // Inventory management
    products.PUT("/:id/stock", productHandler.UpdateStock)
    products.GET("/low-stock", productHandler.GetLowStock)
    
    // Categories
    categories := products.Group("/categories")
    {
        categories.POST("", categoryHandler.CreateCategory)
        categories.GET("", categoryHandler.ListCategories)
        categories.GET("/:id", categoryHandler.GetCategory)
        categories.PUT("/:id", categoryHandler.UpdateCategory)
        categories.DELETE("/:id", categoryHandler.DeleteCategory)
    }
    
    // Variants
    variants := products.Group("/:product_id/variants")
    {
        variants.POST("/options", variantHandler.CreateVariantOptions)
        variants.POST("", variantHandler.CreateVariant)
        variants.GET("", variantHandler.ListVariants)
        variants.PUT("/:variant_id", variantHandler.UpdateVariant)
        variants.DELETE("/:variant_id", variantHandler.DeleteVariant)
    }
}
```

---

### **Phase 7: Testing & Documentation**
**Duration**: 2-3 days  
**Dependencies**: Phase 6 completed

#### **Task 7.1: Unit Tests**
**Files**: 
- `internal/application/usecase/product_usecase_test.go`
- `internal/infrastructure/repository/product_repository_test.go`
- `internal/interfaces/api/handler/product_handler_test.go`

#### **Task 7.2: Integration Tests**
**File**: `tests/integration/product_api_test.go`

#### **Task 7.3: API Documentation**
**File**: `docs/PRODUCT_API_DOCUMENTATION.md`

---

## 📊 **Implementation Timeline**

| Phase | Duration | Start After | Key Deliverables |
|-------|----------|-------------|------------------|
| **Phase 1** | 2-3 days | DB Migration | Core entities & interfaces |
| **Phase 2** | 3-4 days | Phase 1 | Repository implementations |
| **Phase 3** | 2-3 days | Phase 2 | Business logic & use cases |
| **Phase 4** | 1-2 days | Phase 3 | DTOs & validation |
| **Phase 5** | 2-3 days | Phase 4 | HTTP handlers & API |
| **Phase 6** | 1 day | Phase 5 | Route integration |
| **Phase 7** | 2-3 days | Phase 6 | Testing & documentation |

**Total Estimated Duration**: 13-19 days

---

## 🎯 **Success Criteria**

### **Technical Requirements**
- ✅ All database triggers and functions working correctly
- ✅ Full CRUD operations for all entities
- ✅ Variant system with dynamic option validation
- ✅ Comprehensive error handling and validation
- ✅ Performance optimized with proper indexing

### **API Requirements**
- ✅ RESTful API endpoints following consistent patterns
- ✅ Proper HTTP status codes and error responses
- ✅ Request/response validation
- ✅ Authentication and authorization
- ✅ Rate limiting and security measures

### **Business Requirements**
- ✅ Product lifecycle management (draft → active → archived)
- ✅ Inventory tracking with low stock alerts
- ✅ Flexible variant system supporting any option types
- ✅ Hierarchical category system
- ✅ Multi-image support per product
- ✅ SEO-friendly URLs and metadata

---

## 📝 **Next Steps**

1. **Review and approve** this PRD
2. **Set up development environment** with updated dependencies
3. **Start with Phase 1** - Core entity development
4. **Implement phases sequentially** with testing at each step
5. **Conduct code reviews** after each phase
6. **Deploy to staging** for integration testing
7. **Document API endpoints** for frontend integration

---

**Note**: This PRD provides a comprehensive roadmap for implementing the complete product management system. Each phase builds upon the previous one, ensuring a solid foundation and maintainable codebase.
