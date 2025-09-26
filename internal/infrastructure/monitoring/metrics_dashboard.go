package monitoring

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
)

// MetricsDashboard provides HTTP endpoints for performance monitoring dashboard
type MetricsDashboard struct {
	performanceMonitor *QueryPerformanceMonitor
	tenantResolver     tenant.TenantResolver
}

// NewMetricsDashboard creates a new metrics dashboard
func NewMetricsDashboard(
	performanceMonitor *QueryPerformanceMonitor,
	tenantResolver tenant.TenantResolver,
) *MetricsDashboard {
	return &MetricsDashboard{
		performanceMonitor: performanceMonitor,
		tenantResolver:     tenantResolver,
	}
}

// DashboardSummary provides overall performance summary
type DashboardSummary struct {
	TotalQueries      int64                `json:"total_queries"`
	TotalErrors       int64                `json:"total_errors"`
	AverageLatency    float64              `json:"average_latency_ms"`
	SlowQueryCount    int                  `json:"slow_query_count"`
	ActiveTenants     int                  `json:"active_tenants"`
	TopSlowQueries    []SlowQuerySummary   `json:"top_slow_queries"`
	PerformanceByType map[string]TypeStats `json:"performance_by_type"`
	Timestamp         time.Time            `json:"timestamp"`
}

// SlowQuerySummary provides summary of slow queries
type SlowQuerySummary struct {
	QueryPattern    string        `json:"query_pattern"`
	AverageDuration time.Duration `json:"average_duration"`
	ExecutionCount  int64         `json:"execution_count"`
	ErrorRate       float64       `json:"error_rate"`
	LastExecuted    time.Time     `json:"last_executed"`
}

// TypeStats provides statistics by tenant type
type TypeStats struct {
	TenantType     string        `json:"tenant_type"`
	TotalQueries   int64         `json:"total_queries"`
	AverageLatency time.Duration `json:"average_latency"`
	ErrorCount     int64         `json:"error_count"`
	ErrorRate      float64       `json:"error_rate"`
	ActiveTenants  int           `json:"active_tenants"`
}

// TenantPerformance provides per-tenant performance metrics
type TenantPerformance struct {
	StorefrontID   uuid.UUID                    `json:"storefront_id"`
	TenantType     tenant.TenantType            `json:"tenant_type"`
	TotalQueries   int64                        `json:"total_queries"`
	AverageLatency time.Duration                `json:"average_latency"`
	ErrorCount     int64                        `json:"error_count"`
	ErrorRate      float64                      `json:"error_rate"`
	QueryBreakdown map[string]*TenantQueryStats `json:"query_breakdown"`
	SlowQueries    []SlowQueryEntry             `json:"slow_queries"`
	LastActivity   time.Time                    `json:"last_activity"`
}

// RegisterRoutes registers dashboard routes with the Gin router
func (md *MetricsDashboard) RegisterRoutes(r *gin.RouterGroup) {
	dashboard := r.Group("/dashboard")
	{
		dashboard.GET("/summary", md.GetSummary)
		dashboard.GET("/queries", md.GetQueryStats)
		dashboard.GET("/slow-queries", md.GetSlowQueries)
		dashboard.GET("/tenant/:storefront_id", md.GetTenantPerformance)
		dashboard.GET("/tenants", md.GetAllTenantsPerformance)
		dashboard.GET("/health", md.GetHealthStatus)
		dashboard.POST("/clear-stats", md.ClearStats)
	}
}

// GetSummary provides overall dashboard summary
func (md *MetricsDashboard) GetSummary(c *gin.Context) {
	queryStats := md.performanceMonitor.GetQueryStats()
	slowQueries := md.performanceMonitor.GetSlowQueries(10)

	summary := md.calculateSummary(queryStats, slowQueries)

	c.JSON(http.StatusOK, summary)
}

// GetQueryStats returns detailed query statistics
func (md *MetricsDashboard) GetQueryStats(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	queryStats := md.performanceMonitor.GetQueryStats()

	// Sort by total execution count
	type QueryStatWithPattern struct {
		Pattern string      `json:"pattern"`
		Stats   *QueryStats `json:"stats"`
	}

	var sortedStats []QueryStatWithPattern
	for pattern, stats := range queryStats {
		sortedStats = append(sortedStats, QueryStatWithPattern{
			Pattern: pattern,
			Stats:   stats,
		})
	}

	sort.Slice(sortedStats, func(i, j int) bool {
		return sortedStats[i].Stats.TotalExecutions > sortedStats[j].Stats.TotalExecutions
	})

	if limit > len(sortedStats) {
		limit = len(sortedStats)
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": sortedStats[:limit],
		"total":   len(sortedStats),
		"limit":   limit,
	})
}

// GetSlowQueries returns slow query log
func (md *MetricsDashboard) GetSlowQueries(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	slowQueries := md.performanceMonitor.GetSlowQueries(limit)

	c.JSON(http.StatusOK, gin.H{
		"slow_queries": slowQueries,
		"count":        len(slowQueries),
	})
}

// GetTenantPerformance returns performance metrics for a specific tenant
func (md *MetricsDashboard) GetTenantPerformance(c *gin.Context) {
	storefrontIDStr := c.Param("storefront_id")
	storefrontID, err := uuid.Parse(storefrontIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_storefront_id",
			"message": "Invalid storefront ID format",
		})
		return
	}

	performance := md.calculateTenantPerformance(storefrontID)

	c.JSON(http.StatusOK, performance)
}

// GetAllTenantsPerformance returns performance metrics for all active tenants
func (md *MetricsDashboard) GetAllTenantsPerformance(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	queryStats := md.performanceMonitor.GetQueryStats()
	tenantMap := make(map[uuid.UUID]*TenantPerformance)

	// Aggregate performance by tenant
	for _, stat := range queryStats {
		for storefrontID, tenantStat := range stat.TenantBreakdown {
			if tenantPerf, exists := tenantMap[storefrontID]; exists {
				tenantPerf.TotalQueries += tenantStat.ExecutionCount
				tenantPerf.ErrorCount += tenantStat.ErrorCount
				if tenantStat.LastExecuted.After(tenantPerf.LastActivity) {
					tenantPerf.LastActivity = tenantStat.LastExecuted
				}
			} else {
				tenantMap[storefrontID] = &TenantPerformance{
					StorefrontID:   storefrontID,
					TotalQueries:   tenantStat.ExecutionCount,
					AverageLatency: tenantStat.AverageDuration,
					ErrorCount:     tenantStat.ErrorCount,
					LastActivity:   tenantStat.LastExecuted,
					QueryBreakdown: make(map[string]*TenantQueryStats),
				}
			}
		}
	}

	// Convert to slice and sort by activity
	var tenants []TenantPerformance
	for _, tenantPerf := range tenantMap {
		if tenantPerf.TotalQueries > 0 {
			tenantPerf.ErrorRate = float64(tenantPerf.ErrorCount) / float64(tenantPerf.TotalQueries) * 100
		}
		tenants = append(tenants, *tenantPerf)
	}

	sort.Slice(tenants, func(i, j int) bool {
		return tenants[i].LastActivity.After(tenants[j].LastActivity)
	})

	if limit > len(tenants) {
		limit = len(tenants)
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants[:limit],
		"total":   len(tenants),
		"limit":   limit,
	})
}

// GetHealthStatus returns system health status
func (md *MetricsDashboard) GetHealthStatus(c *gin.Context) {
	queryStats := md.performanceMonitor.GetQueryStats()

	totalQueries := int64(0)
	totalErrors := int64(0)

	for _, stat := range queryStats {
		totalQueries += stat.TotalExecutions
		totalErrors += stat.ErrorCount
	}

	errorRate := float64(0)
	if totalQueries > 0 {
		errorRate = float64(totalErrors) / float64(totalQueries) * 100
	}

	status := "healthy"
	if errorRate > 5.0 {
		status = "degraded"
	}
	if errorRate > 15.0 {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        status,
		"total_queries": totalQueries,
		"total_errors":  totalErrors,
		"error_rate":    errorRate,
		"timestamp":     time.Now(),
	})
}

// ClearStats clears all performance statistics
func (md *MetricsDashboard) ClearStats(c *gin.Context) {
	md.performanceMonitor.ClearStats()

	c.JSON(http.StatusOK, gin.H{
		"message":   "Statistics cleared successfully",
		"timestamp": time.Now(),
	})
}

// Private helper methods

func (md *MetricsDashboard) calculateSummary(
	queryStats map[string]*QueryStats,
	slowQueries []SlowQueryEntry,
) DashboardSummary {
	totalQueries := int64(0)
	totalErrors := int64(0)
	totalDuration := time.Duration(0)
	tenantSet := make(map[uuid.UUID]bool)
	typeStats := make(map[string]*TypeStats)

	for _, stat := range queryStats {
		totalQueries += stat.TotalExecutions
		totalErrors += stat.ErrorCount
		totalDuration += stat.TotalDuration

		for tenantID := range stat.TenantBreakdown {
			tenantSet[tenantID] = true
		}
	}

	avgLatency := float64(0)
	if totalQueries > 0 {
		avgLatency = float64(totalDuration.Nanoseconds()) / float64(totalQueries) / 1e6
	}

	// Calculate top slow queries
	topSlowQueries := md.calculateTopSlowQueries(queryStats, 5)

	// Calculate performance by type
	performanceByType := make(map[string]TypeStats)
	for typeName, stats := range typeStats {
		if stats.TotalQueries > 0 {
			stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.TotalQueries) * 100
		}
		performanceByType[typeName] = *stats
	}

	return DashboardSummary{
		TotalQueries:      totalQueries,
		TotalErrors:       totalErrors,
		AverageLatency:    avgLatency,
		SlowQueryCount:    len(slowQueries),
		ActiveTenants:     len(tenantSet),
		TopSlowQueries:    topSlowQueries,
		PerformanceByType: performanceByType,
		Timestamp:         time.Now(),
	}
}

func (md *MetricsDashboard) calculateTopSlowQueries(
	queryStats map[string]*QueryStats,
	limit int,
) []SlowQuerySummary {
	type slowQueryWithPattern struct {
		Pattern string
		Stats   *QueryStats
	}

	var slowQueries []slowQueryWithPattern
	for pattern, stats := range queryStats {
		slowQueries = append(slowQueries, slowQueryWithPattern{
			Pattern: pattern,
			Stats:   stats,
		})
	}

	// Sort by average duration descending
	sort.Slice(slowQueries, func(i, j int) bool {
		return slowQueries[i].Stats.AverageDuration > slowQueries[j].Stats.AverageDuration
	})

	if limit > len(slowQueries) {
		limit = len(slowQueries)
	}

	result := make([]SlowQuerySummary, limit)
	for i, sq := range slowQueries[:limit] {
		errorRate := float64(0)
		if sq.Stats.TotalExecutions > 0 {
			errorRate = float64(sq.Stats.ErrorCount) / float64(sq.Stats.TotalExecutions) * 100
		}

		result[i] = SlowQuerySummary{
			QueryPattern:    sq.Pattern,
			AverageDuration: sq.Stats.AverageDuration,
			ExecutionCount:  sq.Stats.TotalExecutions,
			ErrorRate:       errorRate,
			LastExecuted:    sq.Stats.LastExecuted,
		}
	}

	return result
}

func (md *MetricsDashboard) calculateTenantPerformance(storefrontID uuid.UUID) *TenantPerformance {
	tenantStats := md.performanceMonitor.GetTenantStats(storefrontID)
	slowQueries := md.performanceMonitor.GetSlowQueries(0) // Get all slow queries

	totalQueries := int64(0)
	totalErrors := int64(0)
	totalDuration := time.Duration(0)
	lastActivity := time.Time{}

	for _, stat := range tenantStats {
		totalQueries += stat.ExecutionCount
		totalErrors += stat.ErrorCount
		totalDuration += stat.AverageDuration * time.Duration(stat.ExecutionCount)
		if stat.LastExecuted.After(lastActivity) {
			lastActivity = stat.LastExecuted
		}
	}

	avgLatency := time.Duration(0)
	if totalQueries > 0 {
		avgLatency = time.Duration(int64(totalDuration) / totalQueries)
	}

	errorRate := float64(0)
	if totalQueries > 0 {
		errorRate = float64(totalErrors) / float64(totalQueries) * 100
	}

	// Filter slow queries for this tenant
	var tenantSlowQueries []SlowQueryEntry
	for _, sq := range slowQueries {
		if sq.StorefrontID == storefrontID {
			tenantSlowQueries = append(tenantSlowQueries, sq)
		}
	}

	return &TenantPerformance{
		StorefrontID:   storefrontID,
		TenantType:     tenant.TenantTypeShared, // Would need to resolve actual type
		TotalQueries:   totalQueries,
		AverageLatency: avgLatency,
		ErrorCount:     totalErrors,
		ErrorRate:      errorRate,
		QueryBreakdown: tenantStats,
		SlowQueries:    tenantSlowQueries,
		LastActivity:   lastActivity,
	}
}
