package dto

import (
	"time"
)

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int    `json:"page" validate:"min=1" example:"1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100" example:"20"`
	SortBy   string `json:"sort_by" validate:"omitempty" example:"created_at"`
	SortDir  string `json:"sort_dir" validate:"omitempty,oneof=asc desc" example:"desc"`
}

// Note: PaginationResponse is already defined in admin_user_dto.go

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty" example:"req_123456789"`
	Timestamp time.Time   `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Path      string      `json:"path,omitempty" example:"/api/v1/products"`
	Method    string      `json:"method,omitempty" example:"POST"`
}

// ErrorDetail represents detailed error information
type ErrorDetail struct {
	Code        string                 `json:"code" example:"PRODUCT_NOT_FOUND"`
	Message     string                 `json:"message" example:"Product not found"`
	Details     map[string]interface{} `json:"details,omitempty"`
	UserMessage string                 `json:"user_message,omitempty" example:"The requested product could not be found"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message,omitempty" example:"Operation completed successfully"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty" example:"req_123456789"`
	Timestamp time.Time   `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

// ValidationErrorResponse represents validation error details
type ValidationErrorResponse struct {
	Error     ValidationErrorDetail `json:"error"`
	RequestID string                `json:"request_id,omitempty" example:"req_123456789"`
	Timestamp time.Time             `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Path      string                `json:"path,omitempty" example:"/api/v1/products"`
	Method    string                `json:"method,omitempty" example:"POST"`
}

// ValidationErrorDetail represents detailed validation error information
type ValidationErrorDetail struct {
	Code    string       `json:"code" example:"VALIDATION_FAILED"`
	Message string       `json:"message" example:"Request validation failed"`
	Fields  []FieldError `json:"fields"`
}

// FieldError represents a field-specific validation error
type FieldError struct {
	Field   string `json:"field" example:"name"`
	Value   string `json:"value,omitempty" example:""`
	Message string `json:"message" example:"Name is required"`
	Rule    string `json:"rule" example:"required"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status    string                   `json:"status" example:"healthy"`
	Timestamp time.Time                `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Services  map[string]ServiceHealth `json:"services,omitempty"`
	Version   string                   `json:"version,omitempty" example:"1.0.0"`
	Uptime    string                   `json:"uptime,omitempty" example:"2h30m15s"`
}

// ServiceHealth represents the health status of a service dependency
type ServiceHealth struct {
	Status       string    `json:"status" example:"healthy"`
	ResponseTime string    `json:"response_time,omitempty" example:"15ms"`
	LastChecked  time.Time `json:"last_checked" example:"2023-01-01T00:00:00Z"`
	Error        *string   `json:"error,omitempty"`
}

// MetricsResponse represents system metrics
type MetricsResponse struct {
	RequestCount   int64         `json:"request_count" example:"12345"`
	ErrorCount     int64         `json:"error_count" example:"23"`
	AverageLatency string        `json:"average_latency" example:"125ms"`
	ActiveRequests int           `json:"active_requests" example:"5"`
	SystemMetrics  SystemMetrics `json:"system_metrics"`
	Timestamp      time.Time     `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage" example:"25.5"`
	MemoryUsage float64 `json:"memory_usage" example:"60.2"`
	DiskUsage   float64 `json:"disk_usage" example:"45.8"`
	GoRoutines  int     `json:"goroutines" example:"42"`
}

// BatchResponse represents a response for batch operations
type BatchResponse struct {
	TotalProcessed int            `json:"total_processed" example:"10"`
	SuccessCount   int            `json:"success_count" example:"8"`
	FailureCount   int            `json:"failure_count" example:"2"`
	Failures       []BatchFailure `json:"failures,omitempty"`
	ProcessingTime string         `json:"processing_time" example:"1.5s"`
	Timestamp      time.Time      `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

// BatchFailure represents a failed item in batch processing
type BatchFailure struct {
	Index     int    `json:"index" example:"3"`
	ID        string `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Error     string `json:"error" example:"Validation failed"`
	ErrorCode string `json:"error_code" example:"VALIDATION_FAILED"`
}

// SearchResponse represents a search result response
type SearchResponse struct {
	Results      []interface{}      `json:"results"`
	TotalResults int                `json:"total_results" example:"1234"`
	SearchTime   string             `json:"search_time" example:"45ms"`
	Query        string             `json:"query" example:"wireless headphones"`
	Suggestions  []string           `json:"suggestions,omitempty"`
	Facets       map[string][]Facet `json:"facets,omitempty"`
	Pagination   PaginationResponse `json:"pagination"`
}

// Facet represents a search facet
type Facet struct {
	Value string `json:"value" example:"Electronics"`
	Count int    `json:"count" example:"42"`
}

// BulkImportResponse represents the response from bulk import operations
type BulkImportResponse struct {
	ImportID         string          `json:"import_id" example:"imp_123456789"`
	Status           string          `json:"status" example:"completed"`
	TotalRecords     int             `json:"total_records" example:"1000"`
	ProcessedRecords int             `json:"processed_records" example:"950"`
	SuccessCount     int             `json:"success_count" example:"925"`
	FailureCount     int             `json:"failure_count" example:"25"`
	SkippedCount     int             `json:"skipped_count" example:"50"`
	Failures         []ImportFailure `json:"failures,omitempty"`
	ProcessingTime   string          `json:"processing_time" example:"5m30s"`
	StartedAt        time.Time       `json:"started_at" example:"2023-01-01T00:00:00Z"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty" example:"2023-01-01T00:05:30Z"`
}

// ImportFailure represents a failed record in import operations
type ImportFailure struct {
	RowNumber int                    `json:"row_number" example:"15"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error" example:"Invalid SKU format"`
	ErrorCode string                 `json:"error_code" example:"INVALID_SKU"`
}

// ExportResponse represents the response from export operations
type ExportResponse struct {
	ExportID       string    `json:"export_id" example:"exp_123456789"`
	Status         string    `json:"status" example:"completed"`
	Format         string    `json:"format" example:"csv"`
	TotalRecords   int       `json:"total_records" example:"1000"`
	FileSize       int64     `json:"file_size" example:"1048576"`
	DownloadURL    string    `json:"download_url" example:"https://example.com/exports/file.csv"`
	ExpiresAt      time.Time `json:"expires_at" example:"2023-01-08T00:00:00Z"`
	ProcessingTime string    `json:"processing_time" example:"2m15s"`
	CreatedAt      time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
}

// AuditLogResponse represents an audit log entry
type AuditLogResponse struct {
	ID         string                 `json:"id" example:"audit_123456789"`
	UserID     string                 `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Action     string                 `json:"action" example:"product.create"`
	Resource   string                 `json:"resource" example:"product"`
	ResourceID string                 `json:"resource_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	OldData    map[string]interface{} `json:"old_data,omitempty"`
	NewData    map[string]interface{} `json:"new_data,omitempty"`
	IPAddress  string                 `json:"ip_address" example:"192.168.1.100"`
	UserAgent  string                 `json:"user_agent" example:"Mozilla/5.0..."`
	Timestamp  time.Time              `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}
