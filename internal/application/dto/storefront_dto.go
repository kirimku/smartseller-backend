package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// Storefront Management DTOs

// StorefrontCreateRequest represents a storefront creation request
type StorefrontCreateRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Slug        string  `json:"slug" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
	Domain      *string `json:"domain,omitempty"`
	Subdomain   *string `json:"subdomain,omitempty"`
	OwnerEmail  string  `json:"owner_email" binding:"required,email"`
	OwnerName   string  `json:"owner_name" binding:"required,min=1,max=255"`
}

// StorefrontUpdateRequest represents a storefront update request
type StorefrontUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	Domain      *string `json:"domain,omitempty"`
	Subdomain   *string `json:"subdomain,omitempty"`
}

// StorefrontResponse represents a storefront response
type StorefrontResponse struct {
	ID          uuid.UUID                 `json:"id"`
	Name        string                    `json:"name"`
	Slug        string                    `json:"slug"`
	Description *string                   `json:"description,omitempty"`
	Domain      *string                   `json:"domain,omitempty"`
	Subdomain   *string                   `json:"subdomain,omitempty"`
	Status      string                    `json:"status"`
	Settings    entity.StorefrontSettings `json:"settings"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// StorefrontSearchRequest represents a storefront search request
type StorefrontSearchRequest struct {
	Query     string                   `json:"query,omitempty"`
	Status    *entity.StorefrontStatus `json:"status,omitempty"`
	Page      int                      `json:"page" binding:"min=1"`
	PageSize  int                      `json:"page_size" binding:"min=1,max=100"`
	SortBy    string                   `json:"sort_by,omitempty"`
	SortOrder string                   `json:"sort_order,omitempty" binding:"omitempty,oneof=asc desc"`
}

// PaginatedStorefrontResponse represents a paginated storefront response
type PaginatedStorefrontResponse struct {
	Storefronts []*StorefrontResponse `json:"storefronts"`
	Total       int64                 `json:"total"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
	TotalPages  int64                 `json:"total_pages"`
}

// Domain and Configuration DTOs

// DomainConfigRequest represents a domain configuration request
type DomainConfigRequest struct {
	Domain    string  `json:"domain" binding:"required"`
	Subdomain *string `json:"subdomain,omitempty"`
}

// DomainValidationResponse represents a domain validation response
type DomainValidationResponse struct {
	Domain    string    `json:"domain"`
	Valid     bool      `json:"valid"`
	Available bool      `json:"available"`
	Message   string    `json:"message"`
	CheckedAt time.Time `json:"checked_at"`
}

// StorefrontSettingsRequest represents storefront settings update request
type StorefrontSettingsRequest struct {
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

// Storefront Analytics DTOs

// StorefrontStatsRequest represents a storefront statistics request
type StorefrontStatsRequest struct {
	Period    string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// StorefrontStatsResponse represents storefront statistics
type StorefrontStatsResponse struct {
	StorefrontID   uuid.UUID `json:"storefront_id"`
	TotalCustomers int64     `json:"total_customers"`
	TotalOrders    int64     `json:"total_orders"`
	TotalRevenue   float64   `json:"total_revenue"`
	ActiveProducts int64     `json:"active_products"`
	ConversionRate float64   `json:"conversion_rate"`
	Period         string    `json:"period"`
	Timestamp      time.Time `json:"timestamp"`
}

// StorefrontActivityResponse represents storefront activity response
type StorefrontActivityResponse struct {
	StorefrontID uuid.UUID   `json:"storefront_id"`
	Activities   []*Activity `json:"activities"`
	TotalCount   int64       `json:"total_count"`
	Page         int         `json:"page"`
	PageSize     int         `json:"page_size"`
}

// Migration DTOs

// MigrationStatusResponse represents migration status
type MigrationStatusResponse struct {
	StorefrontID uuid.UUID  `json:"storefront_id"`
	Status       string     `json:"status"`
	CurrentType  string     `json:"current_type"`
	TargetType   string     `json:"target_type"`
	Progress     int        `json:"progress"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
}

// Analytics and Metrics DTOs

// CustomerMetricsRequest represents customer metrics request
type CustomerMetricsRequest struct {
	Period      string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Granularity string     `json:"granularity,omitempty" binding:"omitempty,oneof=hour day week month"`
}

// CustomerMetricsResponse represents customer metrics response
type CustomerMetricsResponse struct {
	TotalCustomers     int64                 `json:"total_customers"`
	NewCustomers       int64                 `json:"new_customers"`
	ActiveCustomers    int64                 `json:"active_customers"`
	ChurnRate          float64               `json:"churn_rate"`
	CustomerGrowthRate float64               `json:"customer_growth_rate"`
	Metrics            []CustomerMetricPoint `json:"metrics"`
	Period             string                `json:"period"`
	Timestamp          time.Time             `json:"timestamp"`
}

// CustomerMetricPoint represents a single metric data point
type CustomerMetricPoint struct {
	Date   time.Time `json:"date"`
	New    int64     `json:"new"`
	Active int64     `json:"active"`
	Total  int64     `json:"total"`
}

// SegmentationRequest represents customer segmentation request
type SegmentationRequest struct {
	Criteria []SegmentationCriteria `json:"criteria" binding:"required"`
	Limit    int                    `json:"limit,omitempty" binding:"max=1000"`
}

// SegmentationCriteria represents segmentation criteria
type SegmentationCriteria struct {
	Field    string      `json:"field" binding:"required"`
	Operator string      `json:"operator" binding:"required,oneof=equals not_equals greater_than less_than contains"`
	Value    interface{} `json:"value" binding:"required"`
}

// SegmentationResponse represents segmentation response
type SegmentationResponse struct {
	Segments    []CustomerSegment `json:"segments"`
	TotalCount  int64             `json:"total_count"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// CustomerSegment represents a customer segment
type CustomerSegment struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Count       int64       `json:"count"`
	Percentage  float64     `json:"percentage"`
	Customers   []uuid.UUID `json:"customers,omitempty"`
}

// LifetimeValueResponse represents customer lifetime value
type LifetimeValueResponse struct {
	CustomerID        uuid.UUID  `json:"customer_id"`
	LifetimeValue     float64    `json:"lifetime_value"`
	PredictedValue    float64    `json:"predicted_value"`
	TotalOrders       int64      `json:"total_orders"`
	AverageOrderValue float64    `json:"average_order_value"`
	FirstOrderDate    *time.Time `json:"first_order_date,omitempty"`
	LastOrderDate     *time.Time `json:"last_order_date,omitempty"`
	CalculatedAt      time.Time  `json:"calculated_at"`
}

// StorefrontMetricsRequest represents storefront metrics request
type StorefrontMetricsRequest struct {
	Period      string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Granularity string     `json:"granularity,omitempty" binding:"omitempty,oneof=hour day week month"`
}

// StorefrontMetricsResponse represents storefront metrics response
type StorefrontMetricsResponse struct {
	StorefrontID      uuid.UUID               `json:"storefront_id"`
	TotalRevenue      float64                 `json:"total_revenue"`
	TotalOrders       int64                   `json:"total_orders"`
	AverageOrderValue float64                 `json:"average_order_value"`
	ConversionRate    float64                 `json:"conversion_rate"`
	CustomerCount     int64                   `json:"customer_count"`
	ProductCount      int64                   `json:"product_count"`
	Metrics           []StorefrontMetricPoint `json:"metrics"`
	Period            string                  `json:"period"`
	Timestamp         time.Time               `json:"timestamp"`
}

// StorefrontMetricPoint represents a storefront metric data point
type StorefrontMetricPoint struct {
	Date              time.Time `json:"date"`
	Revenue           float64   `json:"revenue"`
	Orders            int64     `json:"orders"`
	Customers         int64     `json:"customers"`
	AverageOrderValue float64   `json:"average_order_value"`
}

// PerformanceRequest represents performance request
type PerformanceRequest struct {
	Metrics   []string   `json:"metrics" binding:"required"`
	Period    string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// PerformanceResponse represents performance response
type PerformanceResponse struct {
	StorefrontID uuid.UUID              `json:"storefront_id"`
	Metrics      map[string]interface{} `json:"metrics"`
	Period       string                 `json:"period"`
	Timestamp    time.Time              `json:"timestamp"`
}

// System Metrics DTOs

// SystemMetricsRequest represents system metrics request
type SystemMetricsRequest struct {
	Components []string   `json:"components,omitempty"`
	Period     string     `json:"period,omitempty" binding:"omitempty,oneof=1h 24h 7d 30d"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
}

// SystemMetricsResponse represents system metrics response
type SystemMetricsResponse struct {
	CPUUsage         float64                `json:"cpu_usage"`
	MemoryUsage      float64                `json:"memory_usage"`
	DiskUsage        float64                `json:"disk_usage"`
	DatabaseMetrics  DatabaseMetrics        `json:"database_metrics"`
	CacheMetrics     CacheMetrics           `json:"cache_metrics"`
	RequestMetrics   RequestMetrics         `json:"request_metrics"`
	ComponentMetrics map[string]interface{} `json:"component_metrics"`
	Timestamp        time.Time              `json:"timestamp"`
}

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
	ActiveConnections int     `json:"active_connections"`
	IdleConnections   int     `json:"idle_connections"`
	QueriesPerSecond  float64 `json:"queries_per_second"`
	AverageQueryTime  float64 `json:"average_query_time"`
	SlowQueries       int64   `json:"slow_queries"`
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	HitRate     float64 `json:"hit_rate"`
	MissRate    float64 `json:"miss_rate"`
	KeyCount    int64   `json:"key_count"`
	MemoryUsage int64   `json:"memory_usage"`
}

// RequestMetrics represents HTTP request metrics
type RequestMetrics struct {
	RequestsPerSecond   float64 `json:"requests_per_second"`
	AverageResponseTime float64 `json:"average_response_time"`
	ErrorRate           float64 `json:"error_rate"`
	ActiveRequests      int     `json:"active_requests"`
}

// TenantUsageRequest represents tenant usage request
type TenantUsageRequest struct {
	Period    string     `json:"period,omitempty" binding:"omitempty,oneof=7d 30d 90d 1y"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// TenantUsageResponse represents tenant usage response
type TenantUsageResponse struct {
	TotalTenants    int64         `json:"total_tenants"`
	ActiveTenants   int64         `json:"active_tenants"`
	ResourceUsage   ResourceUsage `json:"resource_usage"`
	TenantBreakdown []TenantUsage `json:"tenant_breakdown"`
	Period          string        `json:"period"`
	Timestamp       time.Time     `json:"timestamp"`
}

// ResourceUsage represents resource usage metrics
type ResourceUsage struct {
	DatabaseSize   int64   `json:"database_size"`
	StorageSize    int64   `json:"storage_size"`
	BandwidthUsage int64   `json:"bandwidth_usage"`
	RequestCount   int64   `json:"request_count"`
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
}

// TenantUsage represents individual tenant usage
type TenantUsage struct {
	StorefrontID   uuid.UUID     `json:"storefront_id"`
	StorefrontName string        `json:"storefront_name"`
	ResourceUsage  ResourceUsage `json:"resource_usage"`
	LastActivity   time.Time     `json:"last_activity"`
}

// Reporting DTOs

// ReportRequest represents a report generation request
type ReportRequest struct {
	Type      string                 `json:"type" binding:"required,oneof=customers storefronts orders products analytics"`
	Format    string                 `json:"format" binding:"required,oneof=csv xlsx pdf json"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	DateRange *DateRange             `json:"date_range,omitempty"`
	GroupBy   []string               `json:"group_by,omitempty"`
	Metrics   []string               `json:"metrics,omitempty"`
}

// ReportResponse represents a report generation response
type ReportResponse struct {
	ReportID     uuid.UUID  `json:"report_id"`
	Status       string     `json:"status"`
	DownloadURL  *string    `json:"download_url,omitempty"`
	Progress     int        `json:"progress"`
	EstimatedETA *int       `json:"estimated_eta,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// ScheduledReportRequest represents a scheduled report request
type ScheduledReportRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description *string                `json:"description,omitempty"`
	ReportType  string                 `json:"report_type" binding:"required"`
	Format      string                 `json:"format" binding:"required,oneof=csv xlsx pdf"`
	Schedule    string                 `json:"schedule" binding:"required"` // Cron expression
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Recipients  []string               `json:"recipients" binding:"required"`
	Enabled     bool                   `json:"enabled"`
}

// ScheduledReportResponse represents a scheduled report response
type ScheduledReportResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	NextRun   *time.Time `json:"next_run,omitempty"`
	LastRun   *time.Time `json:"last_run,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ReportStatusResponse represents report status response
type ReportStatusResponse struct {
	ReportID     uuid.UUID  `json:"report_id"`
	Status       string     `json:"status"`
	Progress     int        `json:"progress"`
	DownloadURL  *string    `json:"download_url,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
	RecordCount  *int64     `json:"record_count,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}
