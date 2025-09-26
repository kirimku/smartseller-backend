package monitoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/kirimku/smartseller-backend/pkg/telegram"
)

// AlertManager manages performance alerts
type AlertManager struct {
	mu                 sync.RWMutex
	config             AlertConfig
	performanceMonitor *QueryPerformanceMonitor
	telegramService    telegram.TelegramService
	alertHistory       []Alert
	lastAlertTime      map[string]time.Time
	enabled            bool
}

// Alert represents a performance alert
type Alert struct {
	ID           uuid.UUID         `json:"id"`
	Type         AlertType         `json:"type"`
	Severity     Severity          `json:"severity"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	StorefrontID uuid.UUID         `json:"storefront_id,omitempty"`
	TenantType   tenant.TenantType `json:"tenant_type,omitempty"`
	QueryPattern string            `json:"query_pattern,omitempty"`
	MetricValue  float64           `json:"metric_value"`
	Threshold    float64           `json:"threshold"`
	Timestamp    time.Time         `json:"timestamp"`
	Acknowledged bool              `json:"acknowledged"`
	Resolved     bool              `json:"resolved"`
	ResolvedAt   *time.Time        `json:"resolved_at,omitempty"`
}

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypeSlowQuery     AlertType = "slow_query"
	AlertTypeHighErrorRate AlertType = "high_error_rate"
	AlertTypeHighLatency   AlertType = "high_latency"
	AlertTypeSystemHealth  AlertType = "system_health"
	AlertTypeTenantIssue   AlertType = "tenant_issue"
)

// Severity represents alert severity levels
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// NewAlertManager creates a new alert manager
func NewAlertManager(
	config AlertConfig,
	performanceMonitor *QueryPerformanceMonitor,
	telegramService telegram.TelegramService,
) *AlertManager {
	return &AlertManager{
		config:             config,
		performanceMonitor: performanceMonitor,
		telegramService:    telegramService,
		alertHistory:       make([]Alert, 0),
		lastAlertTime:      make(map[string]time.Time),
		enabled:            config.Enabled,
	}
}

// StartMonitoring starts the alert monitoring process
func (am *AlertManager) StartMonitoring(ctx context.Context) {
	if !am.enabled {
		return
	}

	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			am.checkAlerts()
		}
	}
}

// checkAlerts checks for various alert conditions
func (am *AlertManager) checkAlerts() {
	am.checkSlowQueries()
	am.checkErrorRates()
	am.checkSystemHealth()
	am.checkTenantPerformance()
}

// checkSlowQueries checks for slow query alerts
func (am *AlertManager) checkSlowQueries() {
	slowQueries := am.performanceMonitor.GetSlowQueries(am.config.SlowQueryAlertCount)

	// Check if we have too many slow queries in the last hour
	recentSlowQueries := 0
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	for _, sq := range slowQueries {
		if sq.Timestamp.After(oneHourAgo) {
			recentSlowQueries++
		}
	}

	if recentSlowQueries >= am.config.SlowQueryAlertCount {
		alertKey := "slow_queries_count"
		if am.shouldAlert(alertKey) {
			alert := Alert{
				ID:          uuid.New(),
				Type:        AlertTypeSlowQuery,
				Severity:    SeverityMedium,
				Title:       "High Slow Query Count Detected",
				Description: fmt.Sprintf("Found %d slow queries in the last hour (threshold: %d)", recentSlowQueries, am.config.SlowQueryAlertCount),
				MetricValue: float64(recentSlowQueries),
				Threshold:   float64(am.config.SlowQueryAlertCount),
				Timestamp:   time.Now(),
			}
			am.triggerAlert(alert)
		}
	}
}

// checkErrorRates checks for high error rates
func (am *AlertManager) checkErrorRates() {
	queryStats := am.performanceMonitor.GetQueryStats()

	for pattern, stats := range queryStats {
		if stats.TotalExecutions == 0 {
			continue
		}

		errorRate := float64(stats.ErrorCount) / float64(stats.TotalExecutions) * 100

		if errorRate >= am.config.ErrorRateThreshold {
			alertKey := fmt.Sprintf("error_rate_%s", pattern)
			if am.shouldAlert(alertKey) {
				severity := SeverityMedium
				if errorRate >= 15.0 {
					severity = SeverityHigh
				}
				if errorRate >= 25.0 {
					severity = SeverityCritical
				}

				alert := Alert{
					ID:           uuid.New(),
					Type:         AlertTypeHighErrorRate,
					Severity:     severity,
					Title:        "High Error Rate Detected",
					Description:  fmt.Sprintf("Query pattern '%s' has error rate of %.2f%% (threshold: %.2f%%)", pattern, errorRate, am.config.ErrorRateThreshold),
					QueryPattern: pattern,
					MetricValue:  errorRate,
					Threshold:    am.config.ErrorRateThreshold,
					Timestamp:    time.Now(),
				}
				am.triggerAlert(alert)
			}
		}
	}
}

// checkSystemHealth checks overall system health
func (am *AlertManager) checkSystemHealth() {
	queryStats := am.performanceMonitor.GetQueryStats()

	totalQueries := int64(0)
	totalErrors := int64(0)
	totalDuration := time.Duration(0)

	for _, stats := range queryStats {
		totalQueries += stats.TotalExecutions
		totalErrors += stats.ErrorCount
		totalDuration += stats.TotalDuration
	}

	if totalQueries == 0 {
		return
	}

	// Check overall error rate
	overallErrorRate := float64(totalErrors) / float64(totalQueries) * 100
	if overallErrorRate >= am.config.ErrorRateThreshold {
		alertKey := "system_error_rate"
		if am.shouldAlert(alertKey) {
			severity := SeverityHigh
			if overallErrorRate >= 15.0 {
				severity = SeverityCritical
			}

			alert := Alert{
				ID:          uuid.New(),
				Type:        AlertTypeSystemHealth,
				Severity:    severity,
				Title:       "System Health Degraded",
				Description: fmt.Sprintf("Overall error rate is %.2f%% (threshold: %.2f%%)", overallErrorRate, am.config.ErrorRateThreshold),
				MetricValue: overallErrorRate,
				Threshold:   am.config.ErrorRateThreshold,
				Timestamp:   time.Now(),
			}
			am.triggerAlert(alert)
		}
	}

	// Check average latency
	avgLatency := time.Duration(int64(totalDuration) / totalQueries)
	if avgLatency >= am.config.LatencyThreshold {
		alertKey := "system_latency"
		if am.shouldAlert(alertKey) {
			alert := Alert{
				ID:          uuid.New(),
				Type:        AlertTypeHighLatency,
				Severity:    SeverityMedium,
				Title:       "High System Latency Detected",
				Description: fmt.Sprintf("Average query latency is %v (threshold: %v)", avgLatency, am.config.LatencyThreshold),
				MetricValue: float64(avgLatency.Milliseconds()),
				Threshold:   float64(am.config.LatencyThreshold.Milliseconds()),
				Timestamp:   time.Now(),
			}
			am.triggerAlert(alert)
		}
	}
}

// checkTenantPerformance checks per-tenant performance issues
func (am *AlertManager) checkTenantPerformance() {
	queryStats := am.performanceMonitor.GetQueryStats()
	tenantPerformance := make(map[uuid.UUID]*TenantMetrics)

	// Aggregate tenant metrics
	for _, stats := range queryStats {
		for storefrontID, tenantStats := range stats.TenantBreakdown {
			if metrics, exists := tenantPerformance[storefrontID]; exists {
				metrics.TotalQueries += tenantStats.ExecutionCount
				metrics.TotalErrors += tenantStats.ErrorCount
			} else {
				tenantPerformance[storefrontID] = &TenantMetrics{
					StorefrontID: storefrontID,
					TotalQueries: tenantStats.ExecutionCount,
					TotalErrors:  tenantStats.ErrorCount,
				}
			}
		}
	}

	// Check each tenant's performance
	for storefrontID, metrics := range tenantPerformance {
		if metrics.TotalQueries == 0 {
			continue
		}

		errorRate := float64(metrics.TotalErrors) / float64(metrics.TotalQueries) * 100
		if errorRate >= am.config.ErrorRateThreshold*2 { // Higher threshold for individual tenants
			alertKey := fmt.Sprintf("tenant_error_rate_%s", storefrontID.String())
			if am.shouldAlert(alertKey) {
				alert := Alert{
					ID:           uuid.New(),
					Type:         AlertTypeTenantIssue,
					Severity:     SeverityMedium,
					Title:        "Tenant Performance Issue",
					Description:  fmt.Sprintf("Tenant %s has error rate of %.2f%%", storefrontID.String(), errorRate),
					StorefrontID: storefrontID,
					MetricValue:  errorRate,
					Threshold:    am.config.ErrorRateThreshold * 2,
					Timestamp:    time.Now(),
				}
				am.triggerAlert(alert)
			}
		}
	}
}

// TenantMetrics holds basic tenant performance metrics
type TenantMetrics struct {
	StorefrontID uuid.UUID
	TotalQueries int64
	TotalErrors  int64
}

// shouldAlert checks if enough time has passed since last alert of this type
func (am *AlertManager) shouldAlert(alertKey string) bool {
	am.mu.RLock()
	lastTime, exists := am.lastAlertTime[alertKey]
	am.mu.RUnlock()

	if !exists {
		return true
	}

	return time.Since(lastTime) >= am.config.AlertCooldown
}

// triggerAlert triggers an alert and sends notifications
func (am *AlertManager) triggerAlert(alert Alert) {
	am.mu.Lock()
	am.alertHistory = append(am.alertHistory, alert)
	am.lastAlertTime[am.getAlertKey(alert)] = alert.Timestamp

	// Maintain max history size
	if len(am.alertHistory) > 1000 {
		am.alertHistory = am.alertHistory[100:]
	}
	am.mu.Unlock()

	// Log alert
	log.Printf("ALERT [%s]: %s - %s", alert.Severity, alert.Title, alert.Description)

	// Send notifications
	am.sendNotifications(alert)
}

// sendNotifications sends alert notifications via configured channels
func (am *AlertManager) sendNotifications(alert Alert) {
	// Send Telegram notification
	if am.config.TelegramEnabled {
		message := am.formatTelegramMessage(alert)
		err := am.telegramService.SendAlert(message)
		if err != nil {
			log.Printf("Failed to send Telegram alert: %v", err)
		}
	}

	// TODO: Send email notification if configured
	// if am.config.EmailEnabled {
	//     am.sendEmailAlert(alert)
	// }
}

// formatTelegramMessage formats alert for Telegram
func (am *AlertManager) formatTelegramMessage(alert Alert) string {
	emoji := am.getSeverityEmoji(alert.Severity)

	message := fmt.Sprintf("%s *SmartSeller Alert*\n\n", emoji)
	message += fmt.Sprintf("*Type:* %s\n", alert.Type)
	message += fmt.Sprintf("*Severity:* %s\n", alert.Severity)
	message += fmt.Sprintf("*Title:* %s\n", alert.Title)
	message += fmt.Sprintf("*Description:* %s\n", alert.Description)

	if alert.StorefrontID != uuid.Nil {
		message += fmt.Sprintf("*Tenant:* %s\n", alert.StorefrontID.String())
	}

	if alert.QueryPattern != "" {
		message += fmt.Sprintf("*Query:* `%s`\n", alert.QueryPattern)
	}

	message += fmt.Sprintf("*Value:* %.2f (Threshold: %.2f)\n", alert.MetricValue, alert.Threshold)
	message += fmt.Sprintf("*Time:* %s", alert.Timestamp.Format("2006-01-02 15:04:05"))

	return message
}

// getSeverityEmoji returns emoji for alert severity
func (am *AlertManager) getSeverityEmoji(severity Severity) string {
	switch severity {
	case SeverityLow:
		return "💙"
	case SeverityMedium:
		return "💛"
	case SeverityHigh:
		return "🧡"
	case SeverityCritical:
		return "❤️"
	default:
		return "⚪"
	}
}

// getAlertKey generates a unique key for alert cooldown tracking
func (am *AlertManager) getAlertKey(alert Alert) string {
	key := string(alert.Type)
	if alert.QueryPattern != "" {
		key += "_" + alert.QueryPattern
	}
	if alert.StorefrontID != uuid.Nil {
		key += "_" + alert.StorefrontID.String()
	}
	return key
}

// GetAlertHistory returns recent alerts
func (am *AlertManager) GetAlertHistory(limit int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if limit <= 0 || limit > len(am.alertHistory) {
		limit = len(am.alertHistory)
	}

	// Return most recent alerts
	start := len(am.alertHistory) - limit
	if start < 0 {
		start = 0
	}

	return am.alertHistory[start:]
}

// AcknowledgeAlert marks an alert as acknowledged
func (am *AlertManager) AcknowledgeAlert(alertID uuid.UUID) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i := range am.alertHistory {
		if am.alertHistory[i].ID == alertID {
			am.alertHistory[i].Acknowledged = true
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// ResolveAlert marks an alert as resolved
func (am *AlertManager) ResolveAlert(alertID uuid.UUID) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i := range am.alertHistory {
		if am.alertHistory[i].ID == alertID {
			now := time.Now()
			am.alertHistory[i].Resolved = true
			am.alertHistory[i].ResolvedAt = &now
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}
