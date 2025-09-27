# Warranty System API Specification

## 📖 Overview

Complete REST API specification for the SmartSeller Warranty Management System. This document defines all endpoints for warranty barcode generation, warranty lookup, claim management, and administrative operations.

**Base URL**: `https://api.smartseller.id/v1`  
**Authentication**: JWT Bearer Token  
**Content-Type**: `application/json`

---

## 🔐 Authentication

### Authentication Schemes

#### Bearer Token Authentication
```http
Authorization: Bearer <jwt_token>
```

#### API Key Authentication (for webhooks)
```http
X-API-Key: <api_key>
```

### User Roles & Permissions

| Role | Permissions |
|------|-------------|
| **Customer** | Warranty lookup, claim submission, status tracking |
| **Store Staff** | Barcode activation, basic claim management |
| **Technician** | Assigned claim management, repair updates |
| **Admin** | Full warranty system management |
| **Super Admin** | System configuration, analytics, user management |

---

## 📋 Warranty Barcode Management

### Generate Single Barcode

Generate a single cryptographically secure warranty barcode.

```http
POST /admin/warranty/barcodes
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "storefront_id": "550e8400-e29b-41d4-a716-446655440001", 
    "warranty_period_months": 12,
    "intended_recipient": "Store Manager",
    "distribution_notes": "For new product display"
}
```

**Response 201 Created:**
```json
{
    "success": true,
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "product_id": "550e8400-e29b-41d4-a716-446655440000",
        "storefront_id": "550e8400-e29b-41d4-a716-446655440001",
        "warranty_start_date": "2024-09-27",
        "warranty_end_date": "2025-09-27",
        "warranty_period_months": 12,
        "status": "active",
        "generation_method": "CSPRNG",
        "entropy_bits": 60,
        "generation_time_ms": 2.1,
        "generation_attempt": 1,
        "collision_checked": true,
        "created_at": "2024-09-27T10:30:00Z",
        "created_by": "550e8400-e29b-41d4-a716-446655440003"
    },
    "security_metrics": {
        "entropy_score": "EXCELLENT",
        "uniqueness_verified": true,
        "cryptographic_strength": "HIGH"
    }
}
```

### Generate Batch Barcodes

Generate multiple barcodes in a single batch operation.

```http
POST /admin/warranty/barcodes/batch
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "storefront_id": "550e8400-e29b-41d4-a716-446655440001",
    "quantity": 1000,
    "warranty_period_months": 12,
    "batch_number": "BATCH-2024-Q4-001",
    "intended_recipient": "Distribution Center",
    "distribution_notes": "For Q4 product launch campaign",
    "security_options": {
        "collision_tolerance": "zero",
        "entropy_monitoring": true,
        "performance_tracking": true
    }
}
```

**Response 202 Accepted:**
```json
{
    "success": true,
    "data": {
        "batch_id": "550e8400-e29b-41d4-a716-446655440010",
        "batch_number": "BATCH-2024-Q4-001",
        "requested_quantity": 1000,
        "generated_quantity": 1000,
        "failed_quantity": 0,
        "collision_count": 3,
        "generation_time": "2.45s",
        "status": "completed",
        "download_url": "/admin/warranty/barcodes/batch/550e8400-e29b-41d4-a716-446655440010/download"
    },
    "statistics": {
        "total_possible_combinations": "1.208925819614629174706176e+18",
        "collision_rate": 0.003,
        "success_rate": 100.0,
        "security_score": "EXCELLENT",
        "recommended_action": "continue",
        "average_generation_time": "2.45ms"
    },
    "performance_metrics": {
        "throughput": "408 codes/second",
        "database_operations": 1003,
        "cache_hit_rate": 0.0
    }
}
```

### Download Batch Results

Download generated batch as CSV/Excel file.

```http
GET /admin/warranty/barcodes/batch/{batch_id}/download
Authorization: Bearer <admin_token>
Query Parameters:
- format: csv|excel (default: csv)
- include_qr: true|false (default: false)
```

**Response 200 OK:**
```
Content-Type: text/csv
Content-Disposition: attachment; filename="batch-REX24-001-barcodes.csv"

barcode_number,product_id,warranty_start_date,warranty_end_date,status
REX24A7M9K2P8Q1N5,550e8400-e29b-41d4-a716-446655440000,2024-09-27,2025-09-27,active
REX24H3R7F9L2X8M6,550e8400-e29b-41d4-a716-446655440000,2024-09-27,2025-09-27,active
...
```

### Validate Barcode Format

Validate barcode format without database lookup.

```http
POST /admin/warranty/barcodes/validate-format
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "barcode_number": "REX24A7M9K2P8Q1N5"
}
```

**Response 200 OK:**
```json
{
    "success": true,
    "data": {
        "valid": true,
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "format_analysis": {
            "prefix_valid": true,
            "year_valid": true,
            "character_set_valid": true,
            "length_valid": true,
            "checksum_valid": true
        },
        "metadata": {
            "format": "REX[YY][RANDOM_12]",
            "year": "24",
            "generation_method": "CSPRNG",
            "estimated_entropy": 60
        }
    }
}
```

### Get Generation Statistics

Retrieve barcode generation statistics and analytics.

```http
GET /admin/warranty/barcodes/generation-stats
Authorization: Bearer <admin_token>
Query Parameters:
- period: today|week|month|year|custom
- start_date: YYYY-MM-DD (for custom period)
- end_date: YYYY-MM-DD (for custom period) 
- storefront_id: UUID (optional filter)
- product_id: UUID (optional filter)
```

**Response 200 OK:**
```json
{
    "success": true,
    "data": {
        "summary": {
            "total_generated": 157834,
            "generation_rate": 127.5,
            "collision_count": 47,
            "collision_rate": 0.0298,
            "average_generation_time": "1.8ms",
            "success_rate": 99.997
        },
        "security_status": {
            "entropy_utilization": 0.000013,
            "security_score": "EXCELLENT",
            "recommended_action": "continue",
            "capacity_remaining": "infinite"
        },
        "performance_metrics": {
            "throughput_per_second": 556,
            "database_performance": "optimal",
            "system_load": "low",
            "error_rate": 0.003
        },
        "period_breakdown": [
            {
                "period": "2024-09-27",
                "generated": 15847,
                "collisions": 5,
                "average_time": "1.9ms"
            }
        ]
    }
}
```

---

## 🔍 Warranty Lookup & Validation

### Public Warranty Lookup

Public endpoint for customers to lookup warranty information.

```http
GET /warranty/lookup/{barcode_number}
```

**Response 200 OK:**
```json
{
    "success": true,
    "data": {
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "status": "active",
        "warranty_start_date": "2024-09-27",
        "warranty_end_date": "2025-09-27",
        "warranty_remaining_days": 365,
        "is_expired": false,
        "can_claim_warranty": true,
        "product": {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "name": "iPhone 15 Pro 256GB",
            "brand": "Apple",
            "model": "A2894",
            "category": "Smartphone"
        },
        "storefront": {
            "id": "550e8400-e29b-41d4-a716-446655440001",
            "name": "TechStore Indonesia",
            "contact_email": "warranty@techstore.id",
            "contact_phone": "+62-21-1234-5678",
            "warranty_policy_url": "https://techstore.id/warranty-policy"
        },
        "claim_instructions": {
            "how_to_claim": "Visit our website or call customer service",
            "required_documents": ["Purchase receipt", "Product photos"],
            "processing_time": "3-5 business days",
            "contact_methods": ["online", "phone", "store_visit"]
        }
    }
}
```

**Response 404 Not Found:**
```json
{
    "success": false,
    "error": {
        "code": "WARRANTY_NOT_FOUND",
        "message": "Warranty barcode not found",
        "details": "The provided barcode number does not exist in our system"
    }
}
```

### Admin Warranty Lookup

Detailed warranty information for administrative use.

```http
GET /admin/warranty/barcodes/{barcode_number}
Authorization: Bearer <admin_token>
```

**Response 200 OK:**
```json
{
    "success": true,
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440002",
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "product_id": "550e8400-e29b-41d4-a716-446655440000",
        "storefront_id": "550e8400-e29b-41d4-a716-446655440001",
        "warranty_start_date": "2024-09-27",
        "warranty_end_date": "2025-09-27",
        "warranty_period_months": 12,
        "status": "active",
        "activation_date": "2024-09-27T14:30:00Z",
        "activated_by": "550e8400-e29b-41d4-a716-446655440020",
        "generation_details": {
            "method": "CSPRNG",
            "entropy_bits": 60,
            "generation_attempt": 1,
            "collision_checked": true,
            "batch_id": "550e8400-e29b-41d4-a716-446655440010",
            "batch_number": "BATCH-2024-Q4-001"
        },
        "usage_history": {
            "lookup_count": 15,
            "last_lookup": "2024-09-27T16:45:00Z",
            "claim_count": 0,
            "last_claim": null
        },
        "created_at": "2024-09-27T10:30:00Z",
        "updated_at": "2024-09-27T14:30:00Z",
        "created_by": "550e8400-e29b-41d4-a716-446655440003"
    }
}
```

### Activate Warranty

Activate warranty when product is purchased by customer.

```http
PUT /admin/warranty/barcodes/{barcode_number}/activate
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "customer_id": "550e8400-e29b-41d4-a716-446655440030",
    "purchase_date": "2024-09-27",
    "purchase_receipt": "RCP-2024-001234",
    "activated_by": "550e8400-e29b-41d4-a716-446655440020",
    "notes": "Purchased through store POS system"
}
```

**Response 200 OK:**
```json
{
    "success": true,
    "message": "Warranty activated successfully",
    "data": {
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "status": "active",
        "activation_date": "2024-09-27T14:30:00Z",
        "warranty_start_date": "2024-09-27",
        "warranty_end_date": "2025-09-27"
    }
}
```

---

## 📋 Warranty Claims Management

### Submit Warranty Claim

Customer submits a new warranty claim.

```http
POST /warranty/claims
Content-Type: application/json
```

**Request Body:**
```json
{
    "barcode_number": "REX24A7M9K2P8Q1N5",
    "issue_description": "Screen flickering and touch not responsive after 3 months of use",
    "issue_category": "hardware_defect",
    "issue_date": "2024-09-25T10:00:00Z",
    "severity": "high",
    "customer_info": {
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phone": "+62812345678",
        "preferred_contact": "email"
    },
    "pickup_address": {
        "street": "Jl. Sudirman No. 123",
        "city": "Jakarta",
        "province": "DKI Jakarta",
        "postal_code": "10220",
        "country": "Indonesia",
        "notes": "Office building, 15th floor"
    },
    "attachments": [
        {
            "type": "photo",
            "description": "Product photo showing defect",
            "file_url": "https://uploads.smartseller.id/claims/photo1.jpg"
        },
        {
            "type": "receipt",
            "description": "Purchase receipt",
            "file_url": "https://uploads.smartseller.id/claims/receipt1.pdf"
        }
    ],
    "customer_notes": "Available for pickup Monday-Friday 9-17",
    "preferred_resolution": "repair"
}
```

**Response 201 Created:**
```json
{
    "success": true,
    "message": "Warranty claim submitted successfully",
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440040",
        "claim_number": "WAR-2024-001234",
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "status": "pending",
        "estimated_processing_time": "3-5 business days",
        "claim_date": "2024-09-27T15:00:00Z",
        "tracking_url": "https://warranty.smartseller.id/claims/WAR-2024-001234",
        "next_steps": [
            "Admin will review your claim within 24 hours",
            "You will receive email notification when status changes",
            "Prepare product for pickup if claim is approved"
        ]
    }
}
```

### Get Claim Status

Check warranty claim status and timeline.

```http
GET /warranty/claims/{claim_number}
```

**Response 200 OK:**
```json
{
    "success": true,
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440040",
        "claim_number": "WAR-2024-001234",
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "status": "in_repair",
        "status_display": "Being Repaired",
        "progress_percentage": 60,
        "estimated_completion": "2024-10-05T17:00:00Z",
        "issue_description": "Screen flickering and touch not responsive",
        "resolution_type": "repair",
        "current_location": "Authorized Repair Center - Jakarta",
        "assigned_technician": {
            "name": "Ahmad Repair Specialist",
            "contact": "ahmad@repaircenter.id",
            "rating": 4.8
        },
        "timeline": [
            {
                "status": "pending",
                "date": "2024-09-27T15:00:00Z",
                "description": "Claim submitted",
                "visible_to_customer": true
            },
            {
                "status": "validated",
                "date": "2024-09-27T18:30:00Z",
                "description": "Claim approved for repair",
                "handled_by": "Admin Team",
                "visible_to_customer": true
            },
            {
                "status": "assigned",
                "date": "2024-09-28T09:15:00Z",
                "description": "Assigned to technician Ahmad",
                "visible_to_customer": true
            },
            {
                "status": "in_repair",
                "date": "2024-09-29T10:00:00Z",
                "description": "Repair work started - replacing display assembly",
                "visible_to_customer": true
            }
        ],
        "next_actions": ["Quality check", "Prepare for shipping"],
        "can_cancel": false,
        "can_update": true
    }
}
```

### Admin Claim Validation

Admin validates or rejects warranty claims.

```http
PUT /admin/warranty/claims/{claim_id}/validate
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body (Approve):**
```json
{
    "action": "validate",
    "notes": "Approved for repair - hardware defect confirmed within warranty terms",
    "estimated_completion_date": "2024-10-05T17:00:00Z",
    "priority": "high",
    "internal_notes": "VIP customer - expedite processing"
}
```

**Request Body (Reject):**
```json
{
    "action": "reject",
    "rejection_reason": "Physical damage not covered under warranty terms",
    "notes": "Customer can repair at own cost through authorized service center",
    "recommended_actions": ["Contact service center", "Purchase extended warranty"]
}
```

**Response 200 OK:**
```json
{
    "success": true,
    "message": "Claim validated successfully",
    "data": {
        "claim_id": "550e8400-e29b-41d4-a716-446655440040",
        "claim_number": "WAR-2024-001234",
        "status": "validated",
        "validated_at": "2024-09-27T18:30:00Z",
        "validated_by": "Admin Team",
        "estimated_completion_date": "2024-10-05T17:00:00Z",
        "next_action": "assign_technician"
    }
}
```

### Assign Technician

Assign warranty claim to a technician.

```http
PUT /admin/warranty/claims/{claim_id}/assign
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "technician_id": "550e8400-e29b-41d4-a716-446655440050",
    "estimated_completion_date": "2024-10-05T17:00:00Z",
    "priority": "high",
    "notes": "Rush repair requested by customer",
    "special_instructions": "Handle with extra care - premium device"
}
```

### Update Repair Progress

Technician updates repair progress.

```http
PUT /admin/warranty/claims/{claim_id}/repair-progress
Authorization: Bearer <technician_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "status": "in_repair",
    "progress_notes": "Display assembly replaced, conducting quality tests",
    "completion_percentage": 80,
    "estimated_completion": "2024-10-04T15:00:00Z",
    "parts_used": [
        {
            "part_name": "iPhone 15 Pro Display Assembly",
            "part_number": "DISP-IP15P-001",
            "cost": 850000.00,
            "supplier": "Apple Authorized Parts"
        }
    ],
    "labor_hours": 2.5,
    "quality_check_passed": true
}
```

### Complete Repair

Mark repair as completed and ready for shipping.

```http
PUT /admin/warranty/claims/{claim_id}/complete-repair
Authorization: Bearer <technician_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "repair_notes": "Display assembly successfully replaced. All functions tested and working properly.",
    "quality_rating": 5,
    "total_repair_cost": 850000.00,
    "labor_cost": 200000.00,
    "parts_cost": 650000.00,
    "warranty_on_repair": 90,
    "completion_photos": [
        "https://uploads.smartseller.id/repairs/completed1.jpg",
        "https://uploads.smartseller.id/repairs/test-results1.jpg"
    ]
}
```

### Ship Repaired Item

Arrange shipping of repaired product back to customer.

```http
PUT /admin/warranty/claims/{claim_id}/ship
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "shipping_provider": "JNE Express",
    "service_type": "REG",
    "tracking_number": "8912340000012345",
    "estimated_delivery_date": "2024-10-07T17:00:00Z",
    "shipping_cost": 25000.00,
    "package_weight": 0.5,
    "package_dimensions": "20x15x5cm",
    "delivery_instructions": "Call customer before delivery",
    "insurance_value": 15000000.00
}
```

---

## 📊 Analytics & Reporting

### Warranty Analytics Dashboard

Get comprehensive warranty system analytics.

```http
GET /admin/warranty/analytics/dashboard
Authorization: Bearer <admin_token>
Query Parameters:
- period: today|week|month|quarter|year
- storefront_id: UUID (optional)
```

**Response 200 OK:**
```json
{
    "success": true,
    "data": {
        "overview": {
            "total_warranties": 157834,
            "active_warranties": 142567,
            "expired_warranties": 15267,
            "total_claims": 3247,
            "pending_claims": 156,
            "completed_claims": 2891,
            "claim_rate": 2.06
        },
        "performance_metrics": {
            "average_resolution_time": "4.2 days",
            "customer_satisfaction": 4.7,
            "first_time_fix_rate": 94.2,
            "sla_compliance": 97.8,
            "cost_per_claim": 425000.00
        },
        "trends": {
            "claims_trend": "decreasing",
            "resolution_time_trend": "improving",
            "satisfaction_trend": "stable",
            "cost_trend": "decreasing"
        },
        "top_issues": [
            {
                "category": "hardware_defect",
                "count": 1247,
                "percentage": 38.4,
                "avg_resolution_time": "3.8 days"
            },
            {
                "category": "software_issue",
                "count": 892,
                "percentage": 27.5,
                "avg_resolution_time": "2.1 days"
            }
        ]
    }
}
```

### Warranty Claims Report

Generate detailed warranty claims report.

```http
GET /admin/warranty/reports/claims
Authorization: Bearer <admin_token>
Query Parameters:
- start_date: YYYY-MM-DD
- end_date: YYYY-MM-DD
- status: pending|validated|rejected|completed (optional)
- storefront_id: UUID (optional)
- format: json|csv|excel (default: json)
```

### Export Warranty Data

Export warranty data for business intelligence.

```http
GET /admin/warranty/export
Authorization: Bearer <admin_token>
Query Parameters:
- type: warranties|claims|repairs|analytics
- start_date: YYYY-MM-DD
- end_date: YYYY-MM-DD
- format: csv|excel|json
- include_personal_data: true|false (default: false)
```

---

## 🔔 Webhooks & Notifications

### Webhook Configuration

Configure webhooks for warranty events.

```http
POST /admin/warranty/webhooks
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "name": "Claim Status Updates",
    "url": "https://your-system.com/webhooks/warranty-claims",
    "events": [
        "claim.created",
        "claim.validated", 
        "claim.assigned",
        "claim.completed"
    ],
    "secret": "webhook_secret_key",
    "active": true
}
```

### Webhook Event Examples

#### Claim Created Event
```json
{
    "event": "claim.created",
    "timestamp": "2024-09-27T15:00:00Z",
    "data": {
        "claim_id": "550e8400-e29b-41d4-a716-446655440040",
        "claim_number": "WAR-2024-001234",
        "barcode_number": "REX24A7M9K2P8Q1N5",
        "customer_email": "john.doe@example.com",
        "issue_category": "hardware_defect",
        "severity": "high"
    }
}
```

#### Claim Status Changed Event
```json
{
    "event": "claim.status_changed",
    "timestamp": "2024-09-27T18:30:00Z",
    "data": {
        "claim_id": "550e8400-e29b-41d4-a716-446655440040",
        "claim_number": "WAR-2024-001234",
        "previous_status": "pending",
        "new_status": "validated",
        "changed_by": "Admin Team"
    }
}
```

---

## ❌ Error Handling

### Standard Error Response Format

```json
{
    "success": false,
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message",
        "details": "Additional error details",
        "timestamp": "2024-09-27T15:00:00Z",
        "request_id": "req_1234567890"
    }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `WARRANTY_NOT_FOUND` | 404 | Warranty barcode not found |
| `CLAIM_NOT_FOUND` | 404 | Warranty claim not found |
| `WARRANTY_EXPIRED` | 422 | Warranty has expired |
| `CLAIM_ALREADY_EXISTS` | 422 | Claim already submitted for this warranty |
| `INVALID_STATUS_TRANSITION` | 422 | Cannot change claim to requested status |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_SERVER_ERROR` | 500 | Server error occurred |

### Validation Error Example

```json
{
    "success": false,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Request validation failed",
        "details": {
            "field_errors": {
                "issue_description": "Issue description is required",
                "customer_email": "Invalid email format",
                "warranty_period_months": "Must be between 1 and 60 months"
            }
        },
        "timestamp": "2024-09-27T15:00:00Z",
        "request_id": "req_1234567890"
    }
}
```

---

## 🔄 Rate Limiting

### Rate Limit Headers

All API responses include rate limiting information:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
X-RateLimit-Window: 3600
```

### Rate Limits by User Type

| User Type | Rate Limit | Window |
|-----------|------------|--------|
| **Customer** | 100 req/hour | 1 hour |
| **Store Staff** | 500 req/hour | 1 hour |
| **Technician** | 300 req/hour | 1 hour |
| **Admin** | 1000 req/hour | 1 hour |
| **Super Admin** | 2000 req/hour | 1 hour |

### Webhook Rate Limits

- Maximum 10 webhook endpoints per account
- Maximum 1000 webhook calls per hour per endpoint
- Retry failed webhooks up to 3 times with exponential backoff

---

## 📝 Request/Response Examples

### Complete Warranty Claim Workflow Example

#### 1. Customer Submits Claim
```bash
curl -X POST https://api.smartseller.id/v1/warranty/claims \
  -H "Content-Type: application/json" \
  -d '{
    "barcode_number": "REX24A7M9K2P8Q1N5",
    "issue_description": "Screen not working",
    "issue_category": "hardware_defect",
    "customer_info": {
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+62812345678"
    }
  }'
```

#### 2. Admin Validates Claim
```bash
curl -X PUT https://api.smartseller.id/v1/admin/warranty/claims/550e8400-e29b-41d4-a716-446655440040/validate \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "validate",
    "notes": "Approved for repair",
    "estimated_completion_date": "2024-10-05T17:00:00Z"
  }'
```

#### 3. Assign Technician
```bash
curl -X PUT https://api.smartseller.id/v1/admin/warranty/claims/550e8400-e29b-41d4-a716-446655440040/assign \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "technician_id": "550e8400-e29b-41d4-a716-446655440050",
    "priority": "high"
  }'
```

#### 4. Customer Checks Status
```bash
curl https://api.smartseller.id/v1/warranty/claims/WAR-2024-001234
```

This comprehensive API specification provides complete coverage of all warranty system functionality with detailed request/response examples, error handling, and integration guidelines.