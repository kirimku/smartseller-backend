# Kirimku Admin API Documentation

This directory contains OpenAPI 3.1 specifications for the Kirimku Admin API endpoints, organized by functional areas.

## API Structure

### Admin Dashboard APIs
**File:** `admin-dashboard-endpoints.yaml`

Contains admin dashboard and analytics endpoints:
- `/api/v1/admin/dashboard/tier-distribution` - User tier distribution statistics
- `/api/v1/admin/dashboard/cashback-statistics` - Cashback analytics and statistics  
- `/api/v1/admin/dashboard/top-cashback-users` - Top cashback recipients

### Admin Transaction Management APIs
**File:** `admin-transaction-endpoints.yaml`

Contains comprehensive transaction management endpoints:
- `/api/v1/admin/transactions` - List and filter transactions
- `/api/v1/admin/transactions/{id}` - Get detailed transaction information
- `/api/v1/admin/transactions/{id}/state` - Update transaction state
- `/api/v1/admin/transactions/{id}/notes` - Add notes to transactions
- `/api/v1/admin/transactions/{id}/refund` - Process transaction refunds
- `/api/v1/admin/transactions/{id}/cancel` - Cancel transactions ✨ **NEW**

## Featured Endpoints

### 🆕 Transaction Cancellation
The new transaction cancellation endpoint provides robust cancellation capabilities:

```
POST /api/v1/admin/transactions/{id}/cancel
```

**Features:**
- Admin-only access with JWT authentication
- Comprehensive validation (state, tracking status, courier support)
- JNT logistics integration for shipment cancellation
- Automatic wallet refund processing
- Force cancellation override for emergencies
- Detailed audit logging
- Atomic operations with rollback on failure

**Cancellation Rules:**
- ✅ Transaction states: `invoiced`, `paid`
- ✅ Tracking status: `manifested` (not picked up)
- ✅ Supported couriers: JNT (J&T Express)
- ✅ Admin authentication required

**Request Example:**
```json
{
  "reason": "Customer requested cancellation due to address change",
  "force_cancel": false
}
```

**Response Example:**
```json
{
  "status": "success",
  "message": "Transaction cancelled successfully",
  "data": {
    "transaction_id": 123,
    "previous_state": "paid",
    "new_state": "refunded",
    "refund_amount": 25000.0,
    "logistics_cancellation": {
      "success": true,
      "courier": "jnt",
      "booking_code": "KB123456789"
    },
    "processed_at": "2025-07-08T10:30:00Z",
    "processed_by": "admin123"
  }
}
```

## Authentication

All admin endpoints require:
- **Bearer Token Authentication** - Admin JWT token
- **Admin Role** - User must have admin privileges
- **Valid Session** - Token must not be expired

## Error Handling

Standardized error responses across all endpoints:
- `400` - Bad Request (validation errors, invalid data)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient privileges)
- `404` - Not Found (resource not found)
- `409` - Conflict (business rule violations)
- `500` - Internal Server Error (system errors)

## Response Format

All responses follow a consistent structure:

**Success Response:**
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": { ... },
  "meta": {
    "http_status": 200,
    "timestamp": "2025-07-08T10:30:00Z"
  }
}
```

**Error Response:**
```json
{
  "status": "error", 
  "message": "Operation failed",
  "error": { ... },
  "meta": {
    "http_status": 400,
    "timestamp": "2025-07-08T10:30:00Z"
  }
}
```

## Development

### Validation
All endpoints include comprehensive request validation with detailed error messages.

### Testing
Use the provided test scripts:
- `test_transaction_cancellation_api.sh` - Test cancellation endpoint
- `cmd/validate_cancellation/main.go` - Validate service interfaces

### Documentation
Auto-generated API documentation available via OpenAPI/Swagger UI when the service is running.

## Production Ready

The admin API endpoints are production-ready with:
- ✅ Comprehensive validation and error handling
- ✅ Security measures and authentication
- ✅ Audit logging and accountability
- ✅ Atomic operations and data consistency
- ✅ Detailed documentation and examples
