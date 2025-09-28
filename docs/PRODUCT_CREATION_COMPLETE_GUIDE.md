# Complete Product Creation Guide

## Overview

This comprehensive guide provides detailed documentation for successfully creating products in the SmartSeller backend system. It covers authentication, API endpoints, request/response structures, validation rules, and complete examples with troubleshooting.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Authentication](#authentication)
3. [API Endpoint](#api-endpoint)
4. [Request Structure](#request-structure)
5. [Response Structure](#response-structure)
6. [Validation Rules](#validation-rules)
7. [Complete Examples](#complete-examples)
8. [Error Handling](#error-handling)
9. [Testing Guide](#testing-guide)
10. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### System Requirements
- Server running on port 8090 (default configuration)
- Valid user account with authentication credentials
- Database connection established
- All required dependencies installed

### Environment Setup
```bash
# Ensure server is running
go run cmd/main.go

# Server should start on port 8090
# Check logs for: "Server starting on port 8090"
```

---

## Authentication

### Step 1: Obtain Access Token

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:
```json
{
  "email_or_phone": "test@example.com",
  "password": "password123"
}
```

**Example cURL**:
```bash
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email_or_phone": "test@example.com",
    "password": "password123"
  }'
```

**Success Response**:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "test@example.com",
      "role": "admin"
    }
  }
}
```

### Step 2: Use Access Token

Include the access token in all subsequent requests:
```bash
Authorization: Bearer YOUR_ACCESS_TOKEN
```

---

## API Endpoint

**Method**: `POST`  
**URL**: `http://localhost:8090/api/v1/products`  
**Content-Type**: `application/json`  
**Authentication**: Required (Bearer Token)

---

## Request Structure

### CreateProductRequest DTO

```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "sku": "string (required)",
  "category_id": "uuid (optional)",
  "brand": "string (optional)",
  "tags": ["string"] (optional),
  "base_price": "decimal (required)",
  "sale_price": "decimal (optional)",
  "cost_price": "decimal (optional)",
  "track_inventory": "boolean (optional, default: true)",
  "stock_quantity": "integer (required)",
  "low_stock_threshold": "integer (optional)",
  "status": "string (optional, default: 'draft')",
  "meta_title": "string (optional)",
  "meta_description": "string (optional)",
  "slug": "string (optional)",
  "weight": "decimal (optional)",
  "dimensions_length": "decimal (optional)",
  "dimensions_width": "decimal (optional)",
  "dimensions_height": "decimal (optional)"
}
```

### Field Details

| Field | Type | Required | Validation | Example |
|-------|------|----------|------------|---------|
| `name` | string | ✅ | 1-255 chars | "Wireless Bluetooth Headphones" |
| `description` | string | ❌ | max 5000 chars | "High-quality wireless headphones..." |
| `sku` | string | ✅ | 3-100 chars, alphanumeric + underscore/hyphen | "WBH-001" |
| `category_id` | uuid | ❌ | valid UUID4 | "550e8400-e29b-41d4-a716-446655440000" |
| `brand` | string | ❌ | max 255 chars | "TechSound" |
| `tags` | array | ❌ | each tag max 50 chars | ["wireless", "bluetooth", "headphones"] |
| `base_price` | decimal | ✅ | min 0 | 199.99 |
| `sale_price` | decimal | ❌ | min 0 | 149.99 |
| `cost_price` | decimal | ❌ | min 0 | 80.00 |
| `track_inventory` | boolean | ❌ | true/false | true |
| `stock_quantity` | integer | ✅ | min 0 | 100 |
| `low_stock_threshold` | integer | ❌ | min 0 | 10 |
| `status` | string | ❌ | draft/active/inactive/archived | "draft" |
| `meta_title` | string | ❌ | max 255 chars | "Best Wireless Headphones" |
| `meta_description` | string | ❌ | max 500 chars | "Discover our premium wireless..." |
| `slug` | string | ❌ | max 255 chars, slug format | "wireless-bluetooth-headphones" |
| `weight` | decimal | ❌ | min 0 | 0.25 |
| `dimensions_length` | decimal | ❌ | min 0 | 20.5 |
| `dimensions_width` | decimal | ❌ | min 0 | 15.2 |
| `dimensions_height` | decimal | ❌ | min 0 | 8.5 |

---

## Response Structure

### Success Response (201 Created)

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7",
    "sku": "WBH-001",
    "name": "Wireless Bluetooth Headphones",
    "description": "High-quality wireless headphones with noise cancellation",
    "category_id": "550e8400-e29b-41d4-a716-446655440000",
    "category": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Electronics",
      "path": "Electronics/Audio/Headphones"
    },
    "brand": "TechSound",
    "tags": ["wireless", "bluetooth", "headphones"],
    "base_price": 199.99,
    "sale_price": 149.99,
    "cost_price": 80.00,
    "effective_price": 149.99,
    "profit_margin": 46.67,
    "track_inventory": true,
    "stock_quantity": 100,
    "low_stock_threshold": 10,
    "is_low_stock": false,
    "status": "draft",
    "meta_title": "Best Wireless Headphones - TechSound",
    "meta_description": "Discover our premium wireless headphones",
    "slug": "wireless-bluetooth-headphones",
    "weight": 0.25,
    "dimensions_length": 20.5,
    "dimensions_width": 15.2,
    "dimensions_height": 8.5,
    "created_by": "550e8400-e29b-41d4-a716-446655440002",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "images": [],
    "variants": []
  }
}
```

---

## Validation Rules

### Required Fields
- `name`: Product name (1-255 characters)
- `sku`: Stock Keeping Unit (3-100 characters, alphanumeric + underscore/hyphen)
- `base_price`: Base price (minimum 0)
- `stock_quantity`: Stock quantity (minimum 0)

### Optional Fields with Validation
- `description`: Maximum 5000 characters
- `category_id`: Must be valid UUID4 format
- `brand`: Maximum 255 characters
- `tags`: Array of strings, each maximum 50 characters
- `sale_price`: Must be minimum 0 if provided
- `cost_price`: Must be minimum 0 if provided
- `low_stock_threshold`: Must be minimum 0 if provided
- `status`: Must be one of: draft, active, inactive, archived
- `meta_title`: Maximum 255 characters
- `meta_description`: Maximum 500 characters
- `slug`: Maximum 255 characters, must be valid slug format
- `weight`: Must be minimum 0 if provided
- `dimensions_*`: Must be minimum 0 if provided

### Business Rules
- SKU must be unique across all products
- If `sale_price` is provided, it's used as `effective_price`
- If `cost_price` is provided, `profit_margin` is calculated automatically
- Default status is "draft" if not specified
- Default `track_inventory` is true if not specified

---

## Complete Examples

### Example 1: Minimal Product Creation

```bash
curl -X POST http://localhost:8090/api/v1/products \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Basic T-Shirt",
    "sku": "TSH-001",
    "base_price": 29.99,
    "stock_quantity": 50
  }'
```

### Example 2: Complete Product Creation

```bash
curl -X POST http://localhost:8090/api/v1/products \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Bluetooth Headphones",
    "description": "High-quality wireless headphones with active noise cancellation, 30-hour battery life, and premium sound quality.",
    "sku": "WBH-001",
    "category_id": "550e8400-e29b-41d4-a716-446655440000",
    "brand": "TechSound",
    "tags": ["wireless", "bluetooth", "headphones", "noise-cancellation"],
    "base_price": 199.99,
    "sale_price": 149.99,
    "cost_price": 80.00,
    "track_inventory": true,
    "stock_quantity": 100,
    "low_stock_threshold": 10,
    "status": "active",
    "meta_title": "Best Wireless Headphones - TechSound WBH-001",
    "meta_description": "Discover our premium wireless headphones with superior sound quality and long battery life.",
    "slug": "wireless-bluetooth-headphones-wbh-001",
    "weight": 0.25,
    "dimensions_length": 20.5,
    "dimensions_width": 15.2,
    "dimensions_height": 8.5
  }'
```

### Example 3: Product with Multiple Tags

```bash
curl -X POST http://localhost:8090/api/v1/products \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gaming Mechanical Keyboard",
    "description": "RGB backlit mechanical keyboard with Cherry MX switches",
    "sku": "GMK-001",
    "brand": "GameTech",
    "tags": ["gaming", "mechanical", "keyboard", "rgb", "cherry-mx"],
    "base_price": 129.99,
    "sale_price": 99.99,
    "cost_price": 60.00,
    "stock_quantity": 75,
    "low_stock_threshold": 15,
    "status": "active"
  }'
```

---

## Error Handling

### Common Error Responses

#### 400 Bad Request - Invalid Request Format
```json
{
  "success": false,
  "message": "Invalid request format",
  "errors": {
    "name": "Name is required",
    "sku": "SKU must be between 3 and 100 characters",
    "base_price": "Base price must be greater than or equal to 0"
  }
}
```

#### 401 Unauthorized - Missing/Invalid Token
```json
{
  "success": false,
  "message": "Authentication required",
  "errors": null
}
```

#### 409 Conflict - Duplicate SKU
```json
{
  "success": false,
  "message": "Product with this SKU already exists",
  "errors": {
    "sku": "SKU 'WBH-001' is already in use"
  }
}
```

#### 400 Bad Request - Invalid Category
```json
{
  "success": false,
  "message": "Specified category does not exist",
  "errors": {
    "category_id": "Category with ID '550e8400-e29b-41d4-a716-446655440000' not found"
  }
}
```

#### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Failed to create product",
  "errors": {
    "internal": "Database connection error"
  }
}
```

---

## Testing Guide

### Step-by-Step Testing Process

#### 1. Verify Server Status
```bash
# Check if server is running
curl -I http://localhost:8090/health
# Expected: 200 OK
```

#### 2. Authenticate
```bash
# Get access token
TOKEN=$(curl -s -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email_or_phone": "test@example.com",
    "password": "password123"
  }' | jq -r '.data.access_token')

echo "Token: $TOKEN"
```

#### 3. Create Test Product
```bash
# Create a simple product
curl -X POST http://localhost:8090/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "sku": "TEST-001",
    "base_price": 99.99,
    "stock_quantity": 10
  }' | jq '.'
```

#### 4. Verify Creation
```bash
# List products to verify creation
curl -X GET http://localhost:8090/api/v1/products \
  -H "Authorization: Bearer $TOKEN" | jq '.data.products[] | select(.sku == "TEST-001")'
```

### Automated Testing Script

```bash
#!/bin/bash

# Product Creation Test Script
BASE_URL="http://localhost:8090"

# 1. Authenticate
echo "🔐 Authenticating..."
TOKEN=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email_or_phone": "test@example.com",
    "password": "password123"
  }' | jq -r '.data.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Authentication failed"
  exit 1
fi

echo "✅ Authentication successful"

# 2. Create product
echo "📦 Creating product..."
RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "sku": "TEST-'$(date +%s)'",
    "base_price": 99.99,
    "stock_quantity": 10
  }')

SUCCESS=$(echo $RESPONSE | jq -r '.success')
if [ "$SUCCESS" = "true" ]; then
  PRODUCT_ID=$(echo $RESPONSE | jq -r '.data.id')
  echo "✅ Product created successfully: $PRODUCT_ID"
else
  echo "❌ Product creation failed:"
  echo $RESPONSE | jq '.'
  exit 1
fi

echo "🎉 All tests passed!"
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: Connection Refused
**Error**: `curl: (7) Failed to connect to localhost:8090`

**Solutions**:
1. Check if server is running: `ps aux | grep "go run"`
2. Verify port configuration in logs
3. Check if port 8090 is available: `netstat -tulpn | grep 8090`
4. Restart server: `go run cmd/main.go`

#### Issue 2: Authentication Failed
**Error**: `400 Bad Request - EmailOrPhone is required`

**Solutions**:
1. Use correct field name: `email_or_phone` (not `email`)
2. Verify credentials in `.env` file
3. Check request Content-Type header
4. Ensure JSON format is valid

#### Issue 3: SKU Already Exists
**Error**: `409 Conflict - Product with this SKU already exists`

**Solutions**:
1. Use unique SKU for each product
2. Check existing products: `GET /api/v1/products`
3. Use timestamp in SKU: `TEST-$(date +%s)`

#### Issue 4: Invalid Category ID
**Error**: `400 Bad Request - Specified category does not exist`

**Solutions**:
1. Create category first or omit `category_id`
2. Verify category exists: `GET /api/v1/categories`
3. Use valid UUID format

#### Issue 5: Validation Errors
**Error**: `400 Bad Request - Invalid request format`

**Solutions**:
1. Check all required fields are provided
2. Verify data types (string, number, boolean)
3. Check field length limits
4. Ensure decimal values are properly formatted

### Debug Commands

```bash
# Check server logs
tail -f app.log

# Test server health
curl -I http://localhost:8090/health

# Validate JSON format
echo '{"name":"test"}' | jq '.'

# Check authentication
curl -v -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email_or_phone":"test@example.com","password":"password123"}'
```

### Performance Considerations

1. **Request Size**: Keep product descriptions under 5000 characters
2. **Tags**: Limit to reasonable number of tags (< 20)
3. **Batch Operations**: For multiple products, consider batch endpoints
4. **Images**: Use separate image upload endpoints after product creation

---

## Next Steps

After successful product creation, you can:

1. **Add Product Images**: Use image upload endpoints
2. **Create Product Variants**: Add size, color, or other variations
3. **Set Up Categories**: Organize products into categories
4. **Configure Inventory**: Set up stock tracking and alerts
5. **Add SEO Data**: Optimize meta titles and descriptions

---

## API Reference Links

- [Product Management API](./FRONTEND_ADMIN_PRODUCT_MANAGEMENT_GUIDE.md)
- [Authentication API](../api/openapi/auth-endpoints.yaml)
- [Product Schemas](../api/openapi/product-schemas.yaml)
- [Complete API Documentation](../api/openapi/complete-api.yaml)

---

## Product Variants

### Overview

Product variants allow you to create different versions of a product with varying attributes like color, size, material, etc. The SmartSeller system supports dynamic variant options and automatic variant generation.

### Variant System Architecture

1. **Variant Options**: Define available options for a product (e.g., Color: Red, Blue, Green)
2. **Product Variants**: Specific combinations of options (e.g., Color: Red, Size: Large)
3. **Dynamic Options**: JSONB-based flexible option storage
4. **Auto-generation**: Automatic SKU and name generation for variants

---

## Creating Variant Options

### Step 1: Create Variant Options

Before creating variants, you must define the available options for your product.

**Endpoint**: `POST /api/v1/products/{product_id}/variant-options`

**Request Body**:
```json
{
  "option_name": "Color",
  "option_values": ["Red", "Blue", "Green", "Black"],
  "display_name": "Product Color",
  "is_required": true,
  "sort_order": 1
}
```

**Example cURL**:
```bash
curl -X POST http://localhost:8090/api/v1/products/6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7/variant-options \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "Color",
    "option_values": ["Red", "Blue", "Green", "Black"],
    "display_name": "Product Color",
    "is_required": true,
    "sort_order": 1
  }'
```

**Success Response**:
```json
{
  "success": true,
  "message": "Variant option created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "product_id": "6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7",
    "option_name": "Color",
    "option_values": ["Red", "Blue", "Green", "Black"],
    "display_name": "Product Color",
    "is_required": true,
    "sort_order": 1,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Step 2: Create Additional Options

Create multiple variant options for your product:

```bash
# Create Size option
curl -X POST http://localhost:8090/api/v1/products/6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7/variant-options \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "Size",
    "option_values": ["XS", "S", "M", "L", "XL"],
    "display_name": "Product Size",
    "is_required": true,
    "sort_order": 2
  }'
```

---

## Creating Product Variants

### Step 1: Create Individual Variants

**Endpoint**: `POST /api/v1/products/{product_id}/variants`

**Request Body**:
```json
{
  "options": {
    "Color": "Red",
    "Size": "L"
  },
  "price": 199.99,
  "compare_at_price": 249.99,
  "cost_price": 120.00,
  "stock_quantity": 25,
  "track_inventory": true,
  "weight": 0.3,
  "barcode": "1234567890123"
}
```

**Example cURL**:
```bash
curl -X POST http://localhost:8090/api/v1/products/6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7/variants \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {
      "Color": "Red",
      "Size": "L"
    },
    "price": 199.99,
    "compare_at_price": 249.99,
    "cost_price": 120.00,
    "stock_quantity": 25,
    "track_inventory": true,
    "weight": 0.3,
    "barcode": "1234567890123"
  }'
```

**Success Response**:
```json
{
  "success": true,
  "message": "Product variant created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440004",
    "product_id": "6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7",
    "variant_name": "Red Large",
    "variant_sku": "WBH-001-RED-L",
    "options": {
      "Color": "Red",
      "Size": "L"
    },
    "price": 199.99,
    "compare_at_price": 249.99,
    "cost_price": 120.00,
    "profit_margin": 79.99,
    "stock_quantity": 25,
    "track_inventory": true,
    "weight": 0.3,
    "barcode": "1234567890123",
    "is_active": true,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Step 2: Generate Multiple Variants

For products with multiple options, you can generate all combinations automatically:

**Endpoint**: `POST /api/v1/products/{product_id}/variants/generate`

**Request Body**:
```json
{
  "options": {
    "Color": ["Red", "Blue", "Green"],
    "Size": ["S", "M", "L", "XL"]
  },
  "base_price": 199.99,
  "cost_price": 120.00,
  "stock_quantity": 10
}
```

**Example cURL**:
```bash
curl -X POST http://localhost:8090/api/v1/products/6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7/variants/generate \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {
      "Color": ["Red", "Blue", "Green"],
      "Size": ["S", "M", "L", "XL"]
    },
    "base_price": 199.99,
    "cost_price": 120.00,
    "stock_quantity": 10
  }'
```

**Success Response**:
```json
{
  "success": true,
  "message": "Product variants generated successfully",
  "data": {
    "generated_count": 12,
    "variants": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440005",
        "variant_name": "Red Small",
        "variant_sku": "WBH-001-RED-S",
        "options": {"Color": "Red", "Size": "S"},
        "price": 199.99,
        "cost_price": 120.00,
        "stock_quantity": 10
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440006",
        "variant_name": "Red Medium",
        "variant_sku": "WBH-001-RED-M",
        "options": {"Color": "Red", "Size": "M"},
        "price": 199.99,
        "cost_price": 120.00,
        "stock_quantity": 10
      }
    ]
  }
}
```

---

## Variant Validation Rules

### Option Validation
- Option names must match existing variant options for the product
- Option values must be from the predefined list for each option
- All required options must be provided
- Option combinations must be unique per product

### Pricing Validation
- Price must be greater than 0
- Cost price must be greater than or equal to 0
- Compare at price must be greater than price (if provided)

### Inventory Validation
- Stock quantity must be greater than or equal to 0
- Track inventory must be boolean

### SKU Generation
- Variant SKUs are auto-generated: `{PRODUCT_SKU}-{OPTION_VALUES}`
- Example: `WBH-001-RED-L` for Color: Red, Size: L
- SKUs are automatically made unique if conflicts occur

---

## Complete Variant Examples

### Example 1: T-Shirt with Color and Size Options

```bash
# Step 1: Create the base product
curl -X POST http://localhost:8090/api/v1/products \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Cotton T-Shirt",
    "sku": "TSH-001",
    "base_price": 29.99,
    "stock_quantity": 0,
    "track_inventory": false
  }'

# Step 2: Create Color variant option
curl -X POST http://localhost:8090/api/v1/products/{PRODUCT_ID}/variant-options \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "Color",
    "option_values": ["White", "Black", "Navy", "Gray"],
    "display_name": "T-Shirt Color",
    "is_required": true,
    "sort_order": 1
  }'

# Step 3: Create Size variant option
curl -X POST http://localhost:8090/api/v1/products/{PRODUCT_ID}/variant-options \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "Size",
    "option_values": ["XS", "S", "M", "L", "XL", "XXL"],
    "display_name": "T-Shirt Size",
    "is_required": true,
    "sort_order": 2
  }'

# Step 4: Generate all variant combinations
curl -X POST http://localhost:8090/api/v1/products/{PRODUCT_ID}/variants/generate \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {
      "Color": ["White", "Black", "Navy", "Gray"],
      "Size": ["XS", "S", "M", "L", "XL", "XXL"]
    },
    "base_price": 29.99,
    "cost_price": 15.00,
    "stock_quantity": 20
  }'
```

### Example 2: Shoes with Color, Size, and Material Options

```bash
# Create Color option
curl -X POST http://localhost:8090/api/v1/products/{PRODUCT_ID}/variant-options \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "Color",
    "option_values": ["Black", "Brown", "White"],
    "display_name": "Shoe Color",
    "is_required": true,
    "sort_order": 1
  }'

# Create Size option
curl -X POST http://localhost:8090/api/v1/products/{PRODUCT_ID}/variant-options \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "Size",
    "option_values": ["7", "8", "9", "10", "11", "12"],
    "display_name": "Shoe Size",
    "is_required": true,
    "sort_order": 2
  }'

# Create Material option
curl -X POST http://localhost:8090/api/v1/products/{PRODUCT_ID}/variant-options \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "Material",
    "option_values": ["Leather", "Canvas", "Synthetic"],
    "display_name": "Shoe Material",
    "is_required": true,
    "sort_order": 3
  }'

# Create specific variant
curl -X POST http://localhost:8090/api/v1/products/{PRODUCT_ID}/variants \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {
      "Color": "Black",
      "Size": "10",
      "Material": "Leather"
    },
    "price": 149.99,
    "compare_at_price": 199.99,
    "cost_price": 80.00,
    "stock_quantity": 15,
    "weight": 1.2,
    "barcode": "1234567890124"
  }'
```

---

## Variant Error Handling

### Common Variant Errors

#### 400 Bad Request - Invalid Option Values
```json
{
  "success": false,
  "message": "Invalid variant options",
  "errors": {
    "Color": "Value 'Purple' is not valid for option 'Color'. Valid values: [Red, Blue, Green, Black]",
    "Size": "Option 'Size' is required but not provided"
  }
}
```

#### 409 Conflict - Duplicate Variant
```json
{
  "success": false,
  "message": "Variant with these options already exists",
  "errors": {
    "options": "A variant with Color: Red, Size: L already exists for this product"
  }
}
```

#### 404 Not Found - Product Not Found
```json
{
  "success": false,
  "message": "Product not found",
  "errors": {
    "product_id": "Product with ID '6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7' not found"
  }
}
```

#### 400 Bad Request - No Variant Options Defined
```json
{
  "success": false,
  "message": "No variant options defined for this product",
  "errors": {
    "variant_options": "Please create variant options before creating variants"
  }
}
```

---

## Variant Testing Guide

### Step-by-Step Variant Testing

#### 1. Test Variant Option Creation
```bash
# Test creating Color option
TOKEN=$(curl -s -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email_or_phone": "test@example.com", "password": "password123"}' | jq -r '.data.access_token')

PRODUCT_ID="6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7"

curl -X POST http://localhost:8090/api/v1/products/$PRODUCT_ID/variant-options \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "TestColor",
    "option_values": ["Red", "Blue"],
    "display_name": "Test Color",
    "is_required": true,
    "sort_order": 1
  }' | jq '.'
```

#### 2. Test Variant Creation
```bash
# Test creating variant with valid options
curl -X POST http://localhost:8090/api/v1/products/$PRODUCT_ID/variants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {"TestColor": "Red"},
    "price": 99.99,
    "stock_quantity": 10
  }' | jq '.'
```

#### 3. Test Invalid Variant Creation
```bash
# Test creating variant with invalid option value
curl -X POST http://localhost:8090/api/v1/products/$PRODUCT_ID/variants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {"TestColor": "Purple"},
    "price": 99.99,
    "stock_quantity": 10
  }' | jq '.'
```

#### 4. Test Duplicate Variant Creation
```bash
# Test creating duplicate variant
curl -X POST http://localhost:8090/api/v1/products/$PRODUCT_ID/variants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {"TestColor": "Red"},
    "price": 99.99,
    "stock_quantity": 10
  }' | jq '.'
```

### Automated Variant Testing Script

```bash
#!/bin/bash

# Variant Testing Script
BASE_URL="http://localhost:8090"
PRODUCT_ID="6b946aee-e2a2-4e5b-b9bb-65b4f86e96a7"

# 1. Authenticate
echo "🔐 Authenticating..."
TOKEN=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email_or_phone": "test@example.com", "password": "password123"}' | jq -r '.data.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Authentication failed"
  exit 1
fi
echo "✅ Authentication successful"

# 2. Create variant option
echo "🎨 Creating variant option..."
OPTION_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/products/$PRODUCT_ID/variant-options \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "option_name": "TestColor",
    "option_values": ["Red", "Blue", "Green"],
    "display_name": "Test Color",
    "is_required": true,
    "sort_order": 1
  }')

OPTION_SUCCESS=$(echo $OPTION_RESPONSE | jq -r '.success')
if [ "$OPTION_SUCCESS" = "true" ]; then
  echo "✅ Variant option created successfully"
else
  echo "❌ Variant option creation failed:"
  echo $OPTION_RESPONSE | jq '.'
fi

# 3. Create variant
echo "📦 Creating variant..."
VARIANT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/products/$PRODUCT_ID/variants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {"TestColor": "Red"},
    "price": 99.99,
    "cost_price": 50.00,
    "stock_quantity": 10
  }')

VARIANT_SUCCESS=$(echo $VARIANT_RESPONSE | jq -r '.success')
if [ "$VARIANT_SUCCESS" = "true" ]; then
  VARIANT_ID=$(echo $VARIANT_RESPONSE | jq -r '.data.id')
  echo "✅ Variant created successfully: $VARIANT_ID"
else
  echo "❌ Variant creation failed:"
  echo $VARIANT_RESPONSE | jq '.'
fi

# 4. Test invalid variant
echo "🚫 Testing invalid variant creation..."
INVALID_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/products/$PRODUCT_ID/variants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "options": {"TestColor": "Purple"},
    "price": 99.99,
    "stock_quantity": 10
  }')

INVALID_SUCCESS=$(echo $INVALID_RESPONSE | jq -r '.success')
if [ "$INVALID_SUCCESS" = "false" ]; then
  echo "✅ Invalid variant correctly rejected"
else
  echo "❌ Invalid variant was incorrectly accepted"
fi

echo "🎉 Variant testing completed!"
```

---

## Variant Troubleshooting

### Common Variant Issues and Solutions

#### Issue 1: Option Value Not Valid
**Error**: `Value 'Purple' is not valid for option 'Color'`

**Solutions**:
1. Check existing option values: Query variant options for the product
2. Use only predefined values from the option definition
3. Update variant option to include new values if needed

#### Issue 2: Variant Already Exists
**Error**: `Variant with these options already exists`

**Solutions**:
1. Check existing variants for the product
2. Use different option combinations
3. Update existing variant instead of creating new one

#### Issue 3: Required Option Missing
**Error**: `Option 'Size' is required but not provided`

**Solutions**:
1. Include all required options in the request
2. Check variant option definitions for required flags
3. Provide values for all options marked as required

#### Issue 4: No Variant Options Defined
**Error**: `No variant options defined for this product`

**Solutions**:
1. Create variant options before creating variants
2. Ensure variant options are properly saved
3. Check product ID is correct

### Debug Commands for Variants

```bash
# Check existing variant options for a product
curl -X GET http://localhost:8090/api/v1/products/{PRODUCT_ID}?include=variant_options \
  -H "Authorization: Bearer $TOKEN" | jq '.data.variant_options'

# Check existing variants for a product
curl -X GET http://localhost:8090/api/v1/products/{PRODUCT_ID}?include=variants \
  -H "Authorization: Bearer $TOKEN" | jq '.data.variants'

# Query database for variant options
psql -h localhost -U smartseller_user -d smartseller_db \
  -c "SELECT option_name, option_values FROM product_variant_options WHERE product_id = '{PRODUCT_ID}';"

# Query database for variants
psql -h localhost -U smartseller_user -d smartseller_db \
  -c "SELECT variant_name, variant_sku, options FROM product_variants WHERE product_id = '{PRODUCT_ID}';"
```

---

*Last Updated: $(date)*
*Version: 1.1*