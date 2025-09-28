# Frontend Admin Product Management Implementation Guide

## Overview

This guide provides comprehensive documentation for implementing product management functionality in the admin frontend. It covers the three core operations: **Create**, **Get**, and **List** products, along with detailed API specifications, data structures, and implementation flows.

## Table of Contents

1. [Authentication Requirements](#authentication-requirements)
2. [API Endpoints Overview](#api-endpoints-overview)
3. [Data Structures](#data-structures)
4. [Product Creation Flow](#product-creation-flow)
5. [Product Retrieval (Get)](#product-retrieval-get)
6. [Product Listing](#product-listing)
7. [Error Handling](#error-handling)
8. [Implementation Examples](#implementation-examples)
9. [Best Practices](#best-practices)

---

## Authentication Requirements

All product management endpoints require authentication using Bearer token:

```javascript
const headers = {
  'Authorization': `Bearer ${accessToken}`,
  'Content-Type': 'application/json'
}
```

**Important Notes:**
- Ensure the user has admin privileges
- Handle token expiration gracefully
- Implement token refresh mechanism

---

## API Endpoints Overview

| Operation | Method | Endpoint | Description |
|-----------|--------|----------|-------------|
| Create Product | `POST` | `/api/v1/products` | Create a new product |
| Get Product | `GET` | `/api/v1/products/{id}` | Retrieve a specific product |
| List Products | `GET` | `/api/v1/products` | Get paginated list of products |
| Update Product | `PUT` | `/api/v1/products/{id}` | Update existing product |
| Delete Product | `DELETE` | `/api/v1/products/{id}` | Soft delete a product |

---

## Data Structures

### CreateProductRequest

```typescript
interface CreateProductRequest {
  // Required fields
  name: string;                    // 1-255 characters
  sku: string;                     // 3-100 characters, alphanumeric with underscore/hyphen
  base_price: number;              // Must be > 0
  
  // Optional fields
  description?: string;            // Max 2000 characters
  category_id?: string;            // UUID format
  brand?: string;                  // Max 100 characters
  tags?: string[];                 // Array of strings, max 20 items, each max 50 chars
  sale_price?: number;             // Must be >= 0
  cost_price?: number;             // Must be >= 0
  stock_quantity?: number;         // Default: 0, min: 0
  low_stock_threshold?: number;    // Default: 10, min: 0
  track_inventory?: boolean;       // Default: true
  
  // Physical attributes
  weight?: number;                 // In kg, min: 0
  dimensions_length?: number;      // In cm, min: 0
  dimensions_width?: number;       // In cm, min: 0
  dimensions_height?: number;      // In cm, min: 0
  
  // SEO fields
  meta_title?: string;             // Max 160 characters
  meta_description?: string;       // Max 320 characters
  slug?: string;                   // URL-friendly, lowercase with hyphens
}
```

### ProductResponse

```typescript
interface ProductResponse {
  id: string;                      // UUID
  sku: string;
  name: string;
  description?: string;
  category_id?: string;
  brand?: string;
  tags: string[];
  
  // Pricing
  base_price: number;
  sale_price?: number;
  cost_price?: number;
  effective_price: number;         // Computed field (sale_price || base_price)
  profit_margin?: number;          // Computed field
  
  // Inventory
  track_inventory: boolean;
  stock_quantity: number;
  low_stock_threshold?: number;
  is_low_stock: boolean;           // Computed field
  
  // Status
  status: 'draft' | 'active' | 'inactive' | 'archived';
  
  // Physical attributes
  weight?: number;
  dimensions_length?: number;
  dimensions_width?: number;
  dimensions_height?: number;
  
  // SEO
  meta_title?: string;
  meta_description?: string;
  slug?: string;
  
  // Metadata
  created_by: string;              // UUID
  created_at: string;              // ISO 8601 timestamp
  updated_at: string;              // ISO 8601 timestamp
}
```

### ProductListResponse

```typescript
interface ProductListResponse {
  products: ProductResponse[];
  pagination: {
    page: number;
    page_size: number;
    total_items: number;
    total_pages: number;
    has_next: boolean;
    has_previous: boolean;
  };
  filters_applied: {
    search?: string;
    category_id?: string;
    brand?: string;
    status?: string;
    min_price?: number;
    max_price?: number;
    low_stock?: boolean;
  };
}
```

---

## Product Creation Flow

### Step-by-Step Implementation

#### 1. Form Validation (Frontend)

```typescript
const validateProductForm = (data: CreateProductRequest): ValidationErrors => {
  const errors: ValidationErrors = {};
  
  // Required field validation
  if (!data.name?.trim()) {
    errors.name = 'Product name is required';
  } else if (data.name.length > 255) {
    errors.name = 'Product name must be less than 255 characters';
  }
  
  if (!data.sku?.trim()) {
    errors.sku = 'SKU is required';
  } else if (!/^[A-Z0-9_-]+$/.test(data.sku)) {
    errors.sku = 'SKU must contain only uppercase letters, numbers, underscores, and hyphens';
  }
  
  if (!data.base_price || data.base_price <= 0) {
    errors.base_price = 'Base price must be greater than 0';
  }
  
  // Optional field validation
  if (data.sale_price && data.sale_price < 0) {
    errors.sale_price = 'Sale price cannot be negative';
  }
  
  if (data.sale_price && data.base_price && data.sale_price > data.base_price) {
    errors.sale_price = 'Sale price cannot be higher than base price';
  }
  
  return errors;
};
```

#### 2. API Call Implementation

```typescript
const createProduct = async (productData: CreateProductRequest): Promise<ProductResponse> => {
  try {
    const response = await fetch('/api/v1/products', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${getAccessToken()}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(productData),
    });
    
    const result = await response.json();
    
    if (!response.ok) {
      throw new APIError(result.message, response.status, result.errors);
    }
    
    return result.data;
  } catch (error) {
    console.error('Failed to create product:', error);
    throw error;
  }
};
```

#### 3. Form Component Example (React)

```typescript
const CreateProductForm: React.FC = () => {
  const [formData, setFormData] = useState<CreateProductRequest>({
    name: '',
    sku: '',
    base_price: 0,
    track_inventory: true,
    stock_quantity: 0,
    low_stock_threshold: 10,
  });
  
  const [errors, setErrors] = useState<ValidationErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate form
    const validationErrors = validateProductForm(formData);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }
    
    setIsSubmitting(true);
    try {
      const newProduct = await createProduct(formData);
      
      // Success handling
      showSuccessMessage('Product created successfully!');
      navigateToProductList();
      
    } catch (error) {
      handleAPIError(error);
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {/* Form fields implementation */}
    </form>
  );
};
```

---

## Product Retrieval (Get)

### API Endpoint

```
GET /api/v1/products/{id}?include=category,variants,images
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string (UUID) | Yes | Product ID |
| `include` | string | No | Comma-separated list of related data to include |

### Implementation

```typescript
const getProduct = async (productId: string, include?: string[]): Promise<ProductResponse> => {
  try {
    const params = new URLSearchParams();
    if (include && include.length > 0) {
      params.append('include', include.join(','));
    }
    
    const url = `/api/v1/products/${productId}${params.toString() ? '?' + params.toString() : ''}`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${getAccessToken()}`,
        'Content-Type': 'application/json',
      },
    });
    
    const result = await response.json();
    
    if (!response.ok) {
      throw new APIError(result.message, response.status);
    }
    
    return result.data;
  } catch (error) {
    console.error('Failed to fetch product:', error);
    throw error;
  }
};
```

### Usage Example

```typescript
// Get basic product data
const product = await getProduct('550e8400-e29b-41d4-a716-446655440000');

// Get product with related data
const productWithDetails = await getProduct(
  '550e8400-e29b-41d4-a716-446655440000',
  ['category', 'variants', 'images']
);
```

---

## Product Listing

### API Endpoint

```
GET /api/v1/products
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | number | 1 | Page number (1-based) |
| `page_size` | number | 20 | Items per page (1-100) |
| `sort_by` | string | 'created_at' | Sort field: name, created_at, updated_at, base_price |
| `sort_desc` | boolean | true | Sort in descending order |
| `search` | string | - | Search in name, description, SKU |
| `category_id` | string (UUID) | - | Filter by category |
| `brand` | string | - | Filter by brand |
| `status` | string | - | Filter by status: draft, active, inactive, archived |
| `min_price` | number | - | Minimum price filter |
| `max_price` | number | - | Maximum price filter |
| `low_stock` | boolean | - | Filter products with low stock |

### Implementation

```typescript
interface ProductListFilters {
  page?: number;
  page_size?: number;
  sort_by?: 'name' | 'created_at' | 'updated_at' | 'base_price';
  sort_desc?: boolean;
  search?: string;
  category_id?: string;
  brand?: string;
  status?: 'draft' | 'active' | 'inactive' | 'archived';
  min_price?: number;
  max_price?: number;
  low_stock?: boolean;
}

const listProducts = async (filters: ProductListFilters = {}): Promise<ProductListResponse> => {
  try {
    const params = new URLSearchParams();
    
    // Add filters to query parameters
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString());
      }
    });
    
    const response = await fetch(`/api/v1/products?${params.toString()}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${getAccessToken()}`,
        'Content-Type': 'application/json',
      },
    });
    
    const result = await response.json();
    
    if (!response.ok) {
      throw new APIError(result.message, response.status);
    }
    
    return result.data;
  } catch (error) {
    console.error('Failed to fetch products:', error);
    throw error;
  }
};
```

### Usage Examples

```typescript
// Get first page with default settings
const products = await listProducts();

// Get products with search and filters
const filteredProducts = await listProducts({
  search: 'wireless headphones',
  category_id: '550e8400-e29b-41d4-a716-446655440001',
  status: 'active',
  min_price: 50,
  max_price: 500,
  page: 1,
  page_size: 20,
  sort_by: 'name',
  sort_desc: false,
});

// Get low stock products
const lowStockProducts = await listProducts({
  low_stock: true,
  sort_by: 'stock_quantity',
  sort_desc: false,
});
```

---

## Error Handling

### Error Response Structure

```typescript
interface APIErrorResponse {
  success: false;
  message: string;
  errors?: ValidationError[];
  error_code?: string;
}

interface ValidationError {
  field: string;
  message: string;
  code: string;
}
```

### Common Error Scenarios

#### 1. Validation Errors (400)

```typescript
// Example validation error response
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "sku",
      "message": "SKU must be unique",
      "code": "DUPLICATE_SKU"
    },
    {
      "field": "base_price",
      "message": "Base price must be greater than 0",
      "code": "INVALID_PRICE"
    }
  ]
}
```

#### 2. Authentication Errors (401)

```typescript
{
  "success": false,
  "message": "Authentication required",
  "error_code": "UNAUTHORIZED"
}
```

#### 3. Not Found Errors (404)

```typescript
{
  "success": false,
  "message": "Product not found",
  "error_code": "PRODUCT_NOT_FOUND"
}
```

#### 4. Conflict Errors (409)

```typescript
{
  "success": false,
  "message": "Product with this SKU already exists",
  "error_code": "SKU_CONFLICT"
}
```

### Error Handling Implementation

```typescript
class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public errors?: ValidationError[]
  ) {
    super(message);
    this.name = 'APIError';
  }
}

const handleAPIError = (error: APIError) => {
  switch (error.status) {
    case 400:
      // Handle validation errors
      if (error.errors) {
        displayValidationErrors(error.errors);
      } else {
        showErrorMessage(error.message);
      }
      break;
      
    case 401:
      // Handle authentication errors
      redirectToLogin();
      break;
      
    case 404:
      showErrorMessage('Product not found');
      break;
      
    case 409:
      showErrorMessage('A product with this SKU already exists');
      break;
      
    case 500:
      showErrorMessage('An unexpected error occurred. Please try again.');
      break;
      
    default:
      showErrorMessage(error.message || 'An error occurred');
  }
};
```

---

## Implementation Examples

### Complete Product Management Component (React)

```typescript
const ProductManagement: React.FC = () => {
  const [products, setProducts] = useState<ProductResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<ProductListFilters>({});
  const [pagination, setPagination] = useState<PaginationInfo | null>(null);
  
  // Load products
  const loadProducts = useCallback(async () => {
    setLoading(true);
    try {
      const response = await listProducts(filters);
      setProducts(response.products);
      setPagination(response.pagination);
    } catch (error) {
      handleAPIError(error);
    } finally {
      setLoading(false);
    }
  }, [filters]);
  
  // Load products on mount and filter changes
  useEffect(() => {
    loadProducts();
  }, [loadProducts]);
  
  // Handle search
  const handleSearch = (searchTerm: string) => {
    setFilters(prev => ({ ...prev, search: searchTerm, page: 1 }));
  };
  
  // Handle pagination
  const handlePageChange = (page: number) => {
    setFilters(prev => ({ ...prev, page }));
  };
  
  // Handle filter changes
  const handleFilterChange = (newFilters: Partial<ProductListFilters>) => {
    setFilters(prev => ({ ...prev, ...newFilters, page: 1 }));
  };
  
  return (
    <div className="product-management">
      <ProductFilters
        filters={filters}
        onFilterChange={handleFilterChange}
        onSearch={handleSearch}
      />
      
      <ProductList
        products={products}
        loading={loading}
        onProductSelect={handleProductSelect}
      />
      
      {pagination && (
        <Pagination
          currentPage={pagination.page}
          totalPages={pagination.total_pages}
          onPageChange={handlePageChange}
        />
      )}
    </div>
  );
};
```

### Product Form Component

```typescript
const ProductForm: React.FC<{ productId?: string }> = ({ productId }) => {
  const [formData, setFormData] = useState<CreateProductRequest>({
    name: '',
    sku: '',
    base_price: 0,
    track_inventory: true,
    stock_quantity: 0,
    low_stock_threshold: 10,
  });
  
  const [errors, setErrors] = useState<ValidationErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isEditing, setIsEditing] = useState(!!productId);
  
  // Load product data for editing
  useEffect(() => {
    if (productId) {
      loadProductData(productId);
    }
  }, [productId]);
  
  const loadProductData = async (id: string) => {
    try {
      const product = await getProduct(id);
      setFormData({
        name: product.name,
        sku: product.sku,
        description: product.description,
        category_id: product.category_id,
        brand: product.brand,
        tags: product.tags,
        base_price: product.base_price,
        sale_price: product.sale_price,
        cost_price: product.cost_price,
        stock_quantity: product.stock_quantity,
        low_stock_threshold: product.low_stock_threshold,
        track_inventory: product.track_inventory,
        weight: product.weight,
        dimensions_length: product.dimensions_length,
        dimensions_width: product.dimensions_width,
        dimensions_height: product.dimensions_height,
        meta_title: product.meta_title,
        meta_description: product.meta_description,
        slug: product.slug,
      });
    } catch (error) {
      handleAPIError(error);
    }
  };
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate form
    const validationErrors = validateProductForm(formData);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }
    
    setIsSubmitting(true);
    try {
      if (isEditing && productId) {
        await updateProduct(productId, formData);
        showSuccessMessage('Product updated successfully!');
      } else {
        await createProduct(formData);
        showSuccessMessage('Product created successfully!');
      }
      
      // Navigate back to product list
      navigateToProductList();
      
    } catch (error) {
      handleAPIError(error);
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="product-form">
      {/* Form implementation */}
    </form>
  );
};
```

---

## Best Practices

### 1. Performance Optimization

- **Implement pagination**: Always use pagination for product lists
- **Debounce search**: Implement search debouncing to reduce API calls
- **Cache responses**: Cache product data where appropriate
- **Lazy loading**: Load product images and details on demand

### 2. User Experience

- **Loading states**: Show loading indicators during API calls
- **Error feedback**: Provide clear, actionable error messages
- **Form validation**: Validate forms both client-side and server-side
- **Auto-save drafts**: Consider auto-saving form data as drafts

### 3. Data Management

- **Optimistic updates**: Update UI immediately for better perceived performance
- **State management**: Use proper state management (Redux, Zustand, etc.)
- **Data normalization**: Normalize product data in your state store
- **Real-time updates**: Consider WebSocket connections for real-time inventory updates

### 4. Security

- **Input sanitization**: Sanitize all user inputs
- **XSS prevention**: Escape HTML content properly
- **CSRF protection**: Implement CSRF tokens where necessary
- **Rate limiting**: Implement client-side rate limiting for API calls

### 5. Accessibility

- **Keyboard navigation**: Ensure all functionality is keyboard accessible
- **Screen readers**: Use proper ARIA labels and semantic HTML
- **Color contrast**: Maintain proper color contrast ratios
- **Focus management**: Manage focus properly in modals and forms

### 6. Testing

- **Unit tests**: Test individual components and functions
- **Integration tests**: Test API integration
- **E2E tests**: Test complete user workflows
- **Error scenarios**: Test error handling thoroughly

---

## Conclusion

This guide provides a comprehensive foundation for implementing product management in your admin frontend. Remember to:

1. **Always validate data** both client-side and server-side
2. **Handle errors gracefully** with clear user feedback
3. **Implement proper loading states** for better UX
4. **Use TypeScript** for better type safety
5. **Follow the API specifications** exactly as documented
6. **Test thoroughly** including edge cases and error scenarios

For additional support or questions about specific implementation details, refer to the API documentation or contact the backend development team.