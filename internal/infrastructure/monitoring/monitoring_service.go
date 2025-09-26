package monitoring

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kirimku/smartseller-backend/internal/infrastructure/database"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/kirimku/smartseller-backend/pkg/telegram"
)

// MonitoringService provides comprehensive database performance monitoring
type MonitoringService struct {
	config              Config
	performanceMonitor  *QueryPerformanceMonitor
	alertManager        *AlertManager
	dashboard           *MetricsDashboard
	connectionManager   *database.ConnectionManager
	enabled             bool
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(
	config Config,
	tenantResolver tenant.TenantResolver,
	connectionManager *database.ConnectionManager,
	telegramService telegram.TelegramService,
) *MonitoringService {
	// Validate and set defaults
	config.Validate()

	// Create performance monitor
	performanceMonitor := NewQueryPerformanceMonitor(
		nil, // metrics collector can be nil for now
		tenantResolver,
		MonitoringConfig{
			Enabled:            config.Performance.Enabled,
			SlowQueryThreshold: config.Performance.SlowQueryThreshold,
			MaxSlowQueries:     config.Performance.MaxSlowQueries,
			MetricsInterval:    config.Performance.MetricsInterval,
			AlertThreshold:     config.Alerts.LatencyThreshold,
			EnableStackTrace:   config.Performance.EnableStackTrace,
			LogSlowQueries:     config.Performance.LogSlowQueries,
		},
	)

	// Create alert manager
	alertManager := NewAlertManager(
		config.Alerts,
		performanceMonitor,
		telegramService,
	)

	// Create dashboard
	dashboard := NewMetricsDashboard(
		performanceMonitor,
		tenantResolver,
	)

	return &MonitoringService{
		config:             config,
		performanceMonitor: performanceMonitor,
		alertManager:       alertManager,
		dashboard:          dashboard,
		connectionManager:  connectionManager,
		enabled:            config.Performance.Enabled,
	}
}

// Start starts the monitoring service
func (ms *MonitoringService) Start(ctx context.Context) error {
	if !ms.enabled {
		log.Println("Monitoring service is disabled")
		return nil
	}

	log.Println("Starting monitoring service...")

	// Start alert monitoring in background
	go ms.alertManager.StartMonitoring(ctx)

	log.Printf("Monitoring service started - Dashboard available on port %d", ms.config.Dashboard.Port)
	return nil
}

// WrapDatabase wraps database connections with monitoring
func (ms *MonitoringService) WrapDatabase(db *sqlx.DB) *MonitoredDB {
	if !ms.enabled {
		return &MonitoredDB{DB: db, monitor: nil}
	}
	return ms.performanceMonitor.WrapDB(db)
}

// RegisterDashboardRoutes registers dashboard routes with Gin router
func (ms *MonitoringService) RegisterDashboardRoutes(r *gin.Engine) {
	if !ms.config.Dashboard.Enabled {
		return
	}

	// Create monitoring route group
	monitoring := r.Group(ms.config.Dashboard.Path)

	// Add basic auth if enabled
	if ms.config.Dashboard.AuthEnabled && ms.config.Dashboard.Username != "" && ms.config.Dashboard.Password != "" {
		monitoring.Use(gin.BasicAuth(gin.Accounts{
			ms.config.Dashboard.Username: ms.config.Dashboard.Password,
		}))
	}

	// Register dashboard routes
	ms.dashboard.RegisterRoutes(monitoring)

	// Add alert management routes
	alerts := monitoring.Group("/alerts")
	{
		alerts.GET("/history", ms.getAlertHistory)
		alerts.POST("/:alert_id/acknowledge", ms.acknowledgeAlert)
		alerts.POST("/:alert_id/resolve", ms.resolveAlert)
		alerts.GET("/stats", ms.getAlertStats)
	}
}

// GetPerformanceStats returns current performance statistics
func (ms *MonitoringService) GetPerformanceStats() map[string]*QueryStats {
	if !ms.enabled {
		return make(map[string]*QueryStats)
	}
	return ms.performanceMonitor.GetQueryStats()
}

// GetSlowQueries returns slow query log
func (ms *MonitoringService) GetSlowQueries(limit int) []SlowQueryEntry {
	if !ms.enabled {
		return []SlowQueryEntry{}
	}
	return ms.performanceMonitor.GetSlowQueries(limit)
}

// ClearStatistics clears all monitoring statistics
func (ms *MonitoringService) ClearStatistics() {
	if ms.enabled {
		ms.performanceMonitor.ClearStats()
	}
}

// IsEnabled returns whether monitoring is enabled
func (ms *MonitoringService) IsEnabled() bool {
	return ms.enabled
}

// GetConfig returns the current monitoring configuration
func (ms *MonitoringService) GetConfig() Config {
	return ms.config
}

// Alert management HTTP handlers

func (ms *MonitoringService) getAlertHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit := 50
	
	if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 || limit <= 0 {
		limit = 50
	}

	alerts := ms.alertManager.GetAlertHistory(limit)
	
	c.JSON(200, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

func (ms *MonitoringService) acknowledgeAlert(c *gin.Context) {
	alertIDStr := c.Param("alert_id")
	
	// Parse UUID would go here, simplified for now
	err := ms.alertManager.AcknowledgeAlert(uuid.New()) // Placeholder
	if err != nil {
		c.JSON(404, gin.H{
			"error":   "alert_not_found",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"message": "Alert acknowledged successfully",
		"alert_id": alertIDStr,
	})
}

func (ms *MonitoringService) resolveAlert(c *gin.Context) {
	alertIDStr := c.Param("alert_id")
	
	// Parse UUID would go here, simplified for now
	err := ms.alertManager.ResolveAlert(uuid.New()) // Placeholder
	if err != nil {
		c.JSON(404, gin.H{
			"error":   "alert_not_found",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(200, gin.H{
		"message": "Alert resolved successfully",
		"alert_id": alertIDStr,
	})
}

func (ms *MonitoringService) getAlertStats(c *gin.Context) {
	alerts := ms.alertManager.GetAlertHistory(0) // Get all alerts
	
	stats := make(map[AlertType]int)
	severityStats := make(map[Severity]int)
	acknowledged := 0
	resolved := 0
	
	for _, alert := range alerts {
		stats[alert.Type]++
		severityStats[alert.Severity]++
		
		if alert.Acknowledged {
			acknowledged++
		}
		if alert.Resolved {
			resolved++
		}
	}
	
	c.JSON(200, gin.H{
		"total_alerts":        len(alerts),
		"alerts_by_type":      stats,
		"alerts_by_severity":  severityStats,
		"acknowledged_alerts": acknowledged,
		"resolved_alerts":     resolved,
	})
}