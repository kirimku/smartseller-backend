# SmartSeller API Documentation

This directory contains comprehensive OpenAPI (Swagger) specifications for the SmartSeller backend API.

## 📁 API Specification Files

### Phase 4 Complete API Specifications

- **`complete-api.yaml`** - 🎯 **Master API specification** with all Phase 4 endpoints
- **`customer-api.yaml`** - 👥 Customer management and address operations
- **`storefront-api.yaml`** - 🏪 Storefront creation and management operations

### Legacy API Files

- `openapi.yaml` - Legacy main specification (Phase 1-3)
- `schemas.yaml` - Common schema definitions (legacy)
- `auth-endpoints.yaml` - Authentication endpoints (legacy)
- `user-endpoints.yaml` - User management endpoints (legacy)
- `product-endpoints.yaml` - Product management endpoints (legacy)
- `product-schemas.yaml` - Product-related schemas (legacy)

## 🚀 Phase 4 API Features

Our Phase 4 implementation includes comprehensive REST APIs for:

### Customer Management
- ✅ Customer registration with validation
- ✅ Profile management and updates
- ✅ Customer search and filtering
- ✅ Account activation/deactivation
- ✅ Customer analytics and statistics

### Address Management
- ✅ Create, update, delete addresses
- ✅ Set default addresses
- ✅ Address validation and geocoding
- ✅ Support for multiple address types

### Storefront Operations
- ✅ Create and customize storefronts
- ✅ Unique slug and domain management
- ✅ Theme and branding customization
- ✅ SEO optimization features
- ✅ Business hours and social links
- ✅ Status management and analytics
- ✅ Storefront duplication

## 📖 API Documentation Overview

### Authentication
- **JWT Bearer Token** authentication for protected endpoints
- **Public endpoints** for storefront viewing and validation
- **Role-based access control** for admin operations

### Response Format
All endpoints follow a consistent response format:

**Success Response:**
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... },
  "meta": {
    "request_id": "req_1234567890",
    "timestamp": "2023-01-15T10:30:00Z"
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Operation failed",
  "error": "validation_error",
  "error_detail": "Detailed error information",
  "validation_errors": ["Field-specific errors"],
  "meta": {
    "http_status": 400,
    "request_id": "req_1234567890",
    "timestamp": "2023-01-15T10:30:00Z"
  }
}
```

### Pagination
List endpoints support pagination:
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)

### Rate Limiting
- **Public endpoints**: 100 requests/minute
- **Authenticated endpoints**: 1000 requests/minute
- **Admin endpoints**: 5000 requests/minute

## 🛠️ Usage Instructions

### 1. View Documentation

**Swagger UI:**
```bash
# Serve locally with swagger-ui
swagger-ui-serve complete-api.yaml
```

**Redoc:**
```bash
# Generate static HTML
redoc-cli build complete-api.yaml --output docs/api.html
```

**VS Code:**
Install the "OpenAPI (Swagger) Editor" extension and open any `.yaml` file.

### 2. Validate Specifications

```bash
# Using swagger-codegen
swagger-codegen validate -i complete-api.yaml

# Using openapi-generator
openapi-generator validate -i complete-api.yaml

# Using spectral (recommended)
spectral lint complete-api.yaml
```

### 3. Generate Code

**Go Server:**
```bash
openapi-generator generate \
  -i complete-api.yaml \
  -g go-server \
  -o ./generated/go-server \
  --additional-properties=packageName=smartseller
```

**TypeScript Client:**
```bash
openapi-generator generate \
  -i complete-api.yaml \
  -g typescript-fetch \
  -o ./generated/typescript-client \
  --additional-properties=npmName=smartseller-client
```

**Postman Collection:**
```bash
openapi-generator generate \
  -i complete-api.yaml \
  -g postman-collection \
  -o ./generated/postman
```

## 🎯 Quick Start

1. **Choose your specification:**
   - Use `complete-api.yaml` for the full Phase 4 API
   - Use individual files for specific domains

2. **Import into your tool:**
   - Swagger UI, Postman, Insomnia, or VS Code
   - Most tools auto-detect OpenAPI 3.0.3 format

3. **Start testing:**
   - Authentication endpoints first
   - Then customer registration
   - Create storefronts and manage addresses

## 📋 API Endpoint Summary

### 🔐 Authentication (Legacy - refer to auth-endpoints.yaml)
- `POST /auth/login` - User login
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout

### 👥 Customer Management
- `POST /customers/register` - Register new customer
- `GET /customers/{id}` - Get customer details
- `PUT /customers/{id}` - Update customer profile
- `GET /customers/search` - Search customers
- `POST /customers/{id}/deactivate` - Deactivate account
- `GET /customers/stats` - Customer statistics

### 📍 Address Management
- `GET /customers/{id}/addresses` - List customer addresses
- `POST /customers/{id}/addresses` - Create new address
- `GET /addresses/{id}` - Get address details
- `PUT /addresses/{id}` - Update address
- `DELETE /addresses/{id}` - Delete address
- `POST /customers/{customer_id}/addresses/{address_id}/default` - Set default
- `POST /addresses/validate` - Validate address
- `POST /addresses/geocode` - Geocode address

### 🏪 Storefront Management
- `POST /storefronts` - Create storefront
- `GET /storefronts` - List storefronts
- `GET /storefronts/{id}` - Get storefront details
- `PUT /storefronts/{id}` - Update storefront
- `DELETE /storefronts/{id}` - Delete storefront
- `GET /storefronts/slug/{slug}` - Get by slug (public)
- `GET /customers/{customer_id}/storefronts` - Customer's storefronts
- `PUT /storefronts/{id}/status` - Update status
- `POST /storefronts/{id}/duplicate` - Duplicate storefront
- `GET /storefronts/{id}/analytics` - Storefront analytics

## 🔍 Testing with cURL

**Register a customer:**
```bash
curl -X POST http://localhost:8080/api/v1/customers/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Create a storefront:**
```bash
curl -X POST http://localhost:8080/api/v1/storefronts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "customer_id": "customer-uuid",
    "name": "My Test Store",
    "slug": "my-test-store",
    "domain": "teststore.smartseller.com",
    "theme": "modern"
  }'
```

---

## 📚 Legacy API Documentation (Phase 1-3)

### Files Overview

#### Main Specification
- **`openapi.yaml`** - Main OpenAPI specification file that imports all endpoints and schemas

#### Endpoint Specifications
- **`auth-endpoints.yaml`** - Authentication and authorization endpoints
- **`user-endpoints.yaml`** - User profile and management endpoints
- **`product-endpoints.yaml`** - Product catalog management endpoints (CRUD operations)

#### Schema Definitions
- **`schemas.yaml`** - Common schemas for authentication and user management
- **`product-schemas.yaml`** - Product-specific request/response schemas

### Product Management API (Legacy)
- ✅ **Create Product** - `POST /api/v1/products`
- ✅ **List Products** - `GET /api/v1/products` (with pagination, filtering, search)
- ✅ **Get Product** - `GET /api/v1/products/{id}`
- ✅ **Update Product** - `PUT /api/v1/products/{id}`
- ✅ **Delete Product** - `DELETE /api/v1/products/{id}`

## 📞 Support

For API documentation questions or issues:
- 📧 Email: api-support@smartseller.com  
- 📚 Full docs: https://docs.smartseller.com
- 🐛 Issues: GitHub repository issues section

## 📄 License

This API specification is licensed under the MIT License.
