package monitoring
package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/kirimku/smartseller-backend/pkg/metrics"
)

// QueryPerformanceMonitor monitors database query performance across tenants
type QueryPerformanceMonitor struct {
	mu                sync.RWMutex
	metricsCollector  interface{} // Using interface{} instead of metrics.Collector
	tenantResolver    tenant.TenantResolver
	queryStats        map[string]*QueryStats
	slowQueryLog      []SlowQueryEntry
	maxSlowQueries    int
	slowQueryThreshold time.Duration
	enabled           bool
}

// QueryStats holds performance statistics for a specific query pattern
type QueryStats struct {
	QueryPattern      string        `json:"query_pattern"`
	TotalExecutions   int64         `json:"total_executions"`
	TotalDuration     time.Duration `json:"total_duration"`
	AverageDuration   time.Duration `json:"average_duration"`
	MinDuration       time.Duration `json:"min_duration"`
	MaxDuration       time.Duration `json:"max_duration"`
	ErrorCount        int64         `json:"error_count"`
	LastExecuted      time.Time     `json:"last_executed"`
	TenantBreakdown   map[uuid.UUID]*TenantQueryStats `json:"tenant_breakdown"`
}

// TenantQueryStats holds per-tenant query statistics
type TenantQueryStats struct {
	StorefrontID    uuid.UUID     `json:"storefront_id"`
	ExecutionCount  int64         `json:"execution_count"`
	AverageDuration time.Duration `json:"average_duration"`
	ErrorCount      int64         `json:"error_count"`
	LastExecuted    time.Time     `json:"last_executed"`
}

// SlowQueryEntry represents a slow query log entry
type SlowQueryEntry struct {
	QueryPattern    string        `json:"query_pattern"`
	ActualQuery     string        `json:"actual_query"`
	Duration        time.Duration `json:"duration"`
	StorefrontID    uuid.UUID     `json:"storefront_id"`
	TenantType      tenant.TenantType `json:"tenant_type"`
	Timestamp       time.Time     `json:"timestamp"`
	ErrorMessage    string        `json:"error_message,omitempty"`
	QueryParams     interface{}   `json:"query_params,omitempty"`
	StackTrace      string        `json:"stack_trace,omitempty"`
}

// MonitoringConfig holds configuration for performance monitoring
type MonitoringConfig struct {
	Enabled             bool          `yaml:"enabled"`
	SlowQueryThreshold  time.Duration `yaml:"slow_query_threshold"`
	MaxSlowQueries      int           `yaml:"max_slow_queries"`
	MetricsInterval     time.Duration `yaml:"metrics_interval"`
	AlertThreshold      time.Duration `yaml:"alert_threshold"`
	EnableStackTrace    bool          `yaml:"enable_stack_trace"`
	LogSlowQueries      bool          `yaml:"log_slow_queries"`
}

// NewQueryPerformanceMonitor creates a new query performance monitor
func NewQueryPerformanceMonitor(
	metricsCollector interface{}, // Using interface{} for now instead of metrics.Collector
	tenantResolver tenant.TenantResolver,
	config MonitoringConfig,
) *QueryPerformanceMonitor {
	monitor := &QueryPerformanceMonitor{
		metricsCollector:   nil, // Set to nil for now
		tenantResolver:     tenantResolver,
		queryStats:         make(map[string]*QueryStats),
		slowQueryLog:       make([]SlowQueryEntry, 0),
		maxSlowQueries:     config.MaxSlowQueries,
		slowQueryThreshold: config.SlowQueryThreshold,
		enabled:            config.Enabled,
	}

	if monitor.maxSlowQueries <= 0 {
		monitor.maxSlowQueries = 1000
	}

	if monitor.slowQueryThreshold <= 0 {
		monitor.slowQueryThreshold = 500 * time.Millisecond
	}

	return monitor
}

// RecordQuery records query execution statistics
func (qpm *QueryPerformanceMonitor) RecordQuery(
	ctx context.Context,
	queryPattern string,
	actualQuery string,
	duration time.Duration,
	err error,
	params interface{},
) {
	if !qpm.enabled {
		return
	}

	// Get tenant context
	storefrontID := qpm.getStorefrontFromContext(ctx)
	tenantType := qpm.getTenantTypeFromContext(ctx)

	// Record metrics
	qpm.recordMetrics(queryPattern, duration, err != nil, storefrontID, tenantType)

	// Update query statistics
	qpm.updateQueryStats(queryPattern, duration, err, storefrontID)

	// Check for slow queries
	if duration >= qpm.slowQueryThreshold {
		qpm.recordSlowQuery(SlowQueryEntry{
			QueryPattern: queryPattern,
			ActualQuery:  actualQuery,
			Duration:     duration,
			StorefrontID: storefrontID,
			TenantType:   tenantType,
			Timestamp:    time.Now(),
			ErrorMessage: qpm.getErrorMessage(err),
			QueryParams:  params,
			StackTrace:   qpm.captureStackTrace(),
		})
	}

	// Log slow queries if configured
	if duration >= qpm.slowQueryThreshold {
		log.Printf("SLOW QUERY: %s (%.2fms) - Tenant: %s", 
			queryPattern, 
			float64(duration.Nanoseconds())/1e6,
			storefrontID.String(),
		)
	}
}

// WrapDB wraps a database connection with performance monitoring
func (qpm *QueryPerformanceMonitor) WrapDB(db *sqlx.DB) *MonitoredDB {
	return &MonitoredDB{
		DB:      db,
		monitor: qpm,
	}
}

// MonitoredDB wraps sqlx.DB with performance monitoring
type MonitoredDB struct {
	*sqlx.DB
	monitor *QueryPerformanceMonitor
}

// QueryxContext wraps sqlx QueryxContext with monitoring
func (mdb *MonitoredDB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := mdb.DB.QueryxContext(ctx, query, args...)
	duration := time.Since(start)
	
	mdb.monitor.RecordQuery(ctx, normalizeQuery(query), query, duration, err, args)
	return rows, err
}

// GetContext wraps sqlx GetContext with monitoring
func (mdb *MonitoredDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := mdb.DB.GetContext(ctx, dest, query, args...)
	duration := time.Since(start)
	
	mdb.monitor.RecordQuery(ctx, normalizeQuery(query), query, duration, err, args)
	return err
}

// SelectContext wraps sqlx SelectContext with monitoring
func (mdb *MonitoredDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := mdb.DB.SelectContext(ctx, dest, query, args...)
	duration := time.Since(start)
	
	mdb.monitor.RecordQuery(ctx, normalizeQuery(query), query, duration, err, args)
	return err
}

// ExecContext wraps sqlx ExecContext with monitoring
func (mdb *MonitoredDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := mdb.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)
	
	mdb.monitor.RecordQuery(ctx, normalizeQuery(query), query, duration, err, args)
	return result, err
}

// GetQueryStats returns query statistics
func (qpm *QueryPerformanceMonitor) GetQueryStats() map[string]*QueryStats {
	qpm.mu.RLock()
	defer qpm.mu.RUnlock()
	
	// Deep copy to avoid race conditions
	stats := make(map[string]*QueryStats)
	for pattern, stat := range qpm.queryStats {
		statCopy := *stat
		statCopy.TenantBreakdown = make(map[uuid.UUID]*TenantQueryStats)
		for tenantID, tenantStat := range stat.TenantBreakdown {
			tenantStatCopy := *tenantStat
			statCopy.TenantBreakdown[tenantID] = &tenantStatCopy
		}
		stats[pattern] = &statCopy
	}
	
	return stats
}

// GetSlowQueries returns slow query log
func (qpm *QueryPerformanceMonitor) GetSlowQueries(limit int) []SlowQueryEntry {
	qpm.mu.RLock()
	defer qpm.mu.RUnlock()
	
	if limit <= 0 || limit > len(qpm.slowQueryLog) {
		limit = len(qpm.slowQueryLog)
	}
	
	// Return most recent entries
	start := len(qpm.slowQueryLog) - limit
	if start < 0 {
		start = 0
	}
	
	return qpm.slowQueryLog[start:]
}

// GetTenantStats returns performance statistics for a specific tenant
func (qpm *QueryPerformanceMonitor) GetTenantStats(storefrontID uuid.UUID) map[string]*TenantQueryStats {
	qpm.mu.RLock()
	defer qpm.mu.RUnlock()
	
	tenantStats := make(map[string]*TenantQueryStats)
	for pattern, stat := range qpm.queryStats {
		if tenantStat, exists := stat.TenantBreakdown[storefrontID]; exists {
			tenantStatCopy := *tenantStat
			tenantStats[pattern] = &tenantStatCopy
		}
	}
	
	return tenantStats
}

// ClearStats clears all statistics (useful for testing)
func (qpm *QueryPerformanceMonitor) ClearStats() {
	qpm.mu.Lock()
	defer qpm.mu.Unlock()
	
	qpm.queryStats = make(map[string]*QueryStats)
	qpm.slowQueryLog = make([]SlowQueryEntry, 0)
}

// Private helper methods

func (qpm *QueryPerformanceMonitor) getStorefrontFromContext(ctx context.Context) uuid.UUID {
	if storefrontID, ok := ctx.Value("storefront_id").(uuid.UUID); ok {
		return storefrontID
	}
	return uuid.Nil
}

func (qpm *QueryPerformanceMonitor) getTenantTypeFromContext(ctx context.Context) tenant.TenantType {
	if tenantType, ok := ctx.Value("tenant_type").(tenant.TenantType); ok {
		return tenantType
	}
	return tenant.TenantTypeShared
}

func (qpm *QueryPerformanceMonitor) recordMetrics(
	queryPattern string,
	duration time.Duration,
	hasError bool,
	storefrontID uuid.UUID,
	tenantType tenant.TenantType,
) {
	// Metrics collection is disabled for now
	// TODO: Implement metrics collection when metrics package is available
	return
}

func (qpm *QueryPerformanceMonitor) updateQueryStats(
	queryPattern string,
	duration time.Duration,
	err error,
	storefrontID uuid.UUID,
) {
	qpm.mu.Lock()
	defer qpm.mu.Unlock()

	stat, exists := qpm.queryStats[queryPattern]
	if !exists {
		stat = &QueryStats{
			QueryPattern:    queryPattern,
			TenantBreakdown: make(map[uuid.UUID]*TenantQueryStats),
			MinDuration:     duration,
			MaxDuration:     duration,
		}
		qpm.queryStats[queryPattern] = stat
	}

	// Update overall stats
	stat.TotalExecutions++
	stat.TotalDuration += duration
	stat.AverageDuration = time.Duration(int64(stat.TotalDuration) / stat.TotalExecutions)
	stat.LastExecuted = time.Now()

	if duration < stat.MinDuration {
		stat.MinDuration = duration
	}
	if duration > stat.MaxDuration {
		stat.MaxDuration = duration
	}

	if err != nil {
		stat.ErrorCount++
	}

	// Update tenant-specific stats
	if storefrontID != uuid.Nil {
		tenantStat, exists := stat.TenantBreakdown[storefrontID]
		if !exists {
			tenantStat = &TenantQueryStats{
				StorefrontID: storefrontID,
			}
			stat.TenantBreakdown[storefrontID] = tenantStat
		}

		tenantStat.ExecutionCount++
		// Calculate running average
		if tenantStat.ExecutionCount == 1 {
			tenantStat.AverageDuration = duration
		} else {
			tenantStat.AverageDuration = time.Duration(
				(int64(tenantStat.AverageDuration)*(tenantStat.ExecutionCount-1) + int64(duration)) / tenantStat.ExecutionCount,
			)
		}
		tenantStat.LastExecuted = time.Now()

		if err != nil {
			tenantStat.ErrorCount++
		}
	}
}

func (qpm *QueryPerformanceMonitor) recordSlowQuery(entry SlowQueryEntry) {
	qpm.mu.Lock()
	defer qpm.mu.Unlock()

	qpm.slowQueryLog = append(qpm.slowQueryLog, entry)

	// Maintain maximum size
	if len(qpm.slowQueryLog) > qpm.maxSlowQueries {
		qpm.slowQueryLog = qpm.slowQueryLog[1:]
	}
}

func (qpm *QueryPerformanceMonitor) getErrorMessage(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func (qpm *QueryPerformanceMonitor) captureStackTrace() string {
	// Simple stack trace capture - in production you might want to use runtime.Stack()
	return "stack trace capture not implemented"
}

// normalizeQuery normalizes query patterns for consistent tracking
func normalizeQuery(query string) string {
	// Simple query normalization - replace parameters with placeholders
	// In production, you might want more sophisticated normalization
	return query
}