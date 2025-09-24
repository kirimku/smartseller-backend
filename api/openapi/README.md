# SmartSeller Backend API Documentation

This directory contains the OpenAPI 3.0 specification for the SmartSeller backend API.

## Files Overview

### Main Specification
- **`openapi.yaml`** - Main OpenAPI specification file that imports all endpoints and schemas

### Endpoint Specifications
- **`auth-endpoints.yaml`** - Authentication and authorization endpoints
- **`user-endpoints.yaml`** - User profile and management endpoints
- **`product-endpoints.yaml`** - Product catalog management endpoints (CRUD operations)

### Schema Definitions
- **`schemas.yaml`** - Common schemas for authentication and user management
- **`product-schemas.yaml`** - Product-specific request/response schemas

## API Documentation Features

### Product Management API
- ✅ **Create Product** - `POST /api/v1/products`
- ✅ **List Products** - `GET /api/v1/products` (with pagination, filtering, search)
- ✅ **Get Product** - `GET /api/v1/products/{id}`
- ✅ **Update Product** - `PUT /api/v1/products/{id}`
- ✅ **Delete Product** - `DELETE /api/v1/products/{id}`

### Key Features
- **Authentication**: Bearer token authentication for all endpoints
- **Validation**: Comprehensive request validation with detailed error responses
- **Pagination**: Built-in pagination support with metadata
- **Filtering**: Advanced filtering options (category, brand, price range, stock status)
- **Search**: Full-text search across product name, description, and SKU
- **Error Handling**: Standardized error responses with user-friendly messages
- **Examples**: Complete request/response examples for all operations

## Viewing the Documentation

### Using Swagger UI
1. Copy the contents of `openapi.yaml` 
2. Go to [Swagger Editor](https://editor.swagger.io/)
3. Paste the content to view the interactive documentation

### Using Postman
1. Import `openapi.yaml` into Postman
2. Generate a collection from the OpenAPI specification
3. Use the generated collection for testing

### Local Swagger UI Setup
```bash
# Using Docker
docker run -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd):/api swaggerapi/swagger-ui

# Using npm
npm install -g swagger-ui-serve
swagger-ui-serve openapi.yaml
```

## API Base URLs

- **Development**: `http://localhost:8080`
- **Staging**: `https://api-staging.smartseller.com`
- **Production**: `https://api.smartseller.com`

## Authentication

All endpoints require Bearer token authentication:

```
Authorization: Bearer <jwt-token>
```

Get your JWT token by calling the `/api/v1/auth/login` endpoint.

## Common Request Headers

```http
Content-Type: application/json
Authorization: Bearer <jwt-token>
Accept: application/json
```

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data varies by endpoint
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Technical error message",
    "user_message": "User-friendly error message",
    "details": {
      // Additional error details
    }
  },
  "request_id": "req_123456789",
  "timestamp": "2025-09-24T10:00:00Z",
  "path": "/api/v1/products",
  "method": "POST"
}
```

## Validation Errors

Validation errors include field-specific details:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Request validation failed",
    "fields": [
      {
        "field": "sku",
        "value": "invalid-sku",
        "message": "SKU must be 3-100 characters, uppercase alphanumeric with hyphens/underscores only",
        "rule": "sku_format"
      }
    ]
  },
  "request_id": "req_123456789",
  "timestamp": "2025-09-24T10:00:00Z",
  "path": "/api/v1/products",
  "method": "POST"
}
```

## Product API Examples

### Create Product
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Bluetooth Headphones",
    "description": "High-quality wireless headphones with noise cancellation",
    "sku": "WBH-001",
    "category_id": "550e8400-e29b-41d4-a716-446655440001",
    "brand": "AudioTech",
    "tags": ["electronics", "audio", "wireless"],
    "base_price": 199.99,
    "sale_price": 149.99,
    "cost_price": 100.00,
    "stock_quantity": 50,
    "low_stock_threshold": 10,
    "track_inventory": true
  }'
```

### List Products with Filters
```bash
curl -X GET "http://localhost:8080/api/v1/products?page=1&page_size=20&search=headphones&brand=AudioTech&sort_by=created_at&sort_desc=true" \
  -H "Authorization: Bearer <token>"
```

### Update Product
```bash
curl -X PUT http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Wireless Bluetooth Headphones",
    "sale_price": 129.99,
    "stock_quantity": 75
  }'
```

## Testing

The API includes comprehensive integration tests located in `/tests/api/`. See the main project README for testing instructions.

## Support

For API support and questions:
- Email: api-support@smartseller.com
- Documentation: This OpenAPI specification
- Issues: GitHub repository issues
