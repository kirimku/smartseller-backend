package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Business metrics
	transactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transactions_total",
			Help: "Total number of transactions",
		},
		[]string{"status", "payment_method", "courier"},
	)

	transactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "transaction_amount",
			Help:    "Transaction amounts in rupiah",
			Buckets: []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000, 2500000, 5000000},
		},
		[]string{"status", "payment_method"},
	)

	// Wallet metrics
	walletOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wallet_operations_total",
			Help: "Total number of wallet operations",
		},
		[]string{"operation", "status"},
	)

	walletBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wallet_balance_rupiah",
			Help: "Current wallet balances in rupiah",
		},
		[]string{"wallet_id", "user_type"},
	)

	// Database metrics
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table"},
	)

	// External API metrics
	externalAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_api_calls_total",
			Help: "Total number of external API calls",
		},
		[]string{"service", "method", "status_code"},
	)

	externalAPICallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_api_call_duration_seconds",
			Help:    "Duration of external API calls in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 25, 60},
		},
		[]string{"service", "method", "status_code"},
	)

	// System metrics
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of currently active users",
		},
	)

	cacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result"},
	)
)

// MetricsCollector provides methods to record various metrics
type MetricsCollector struct{}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// RecordHTTPRequest records HTTP request metrics
func (m *MetricsCollector) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration.Seconds())
}

// RecordTransaction records transaction metrics
func (m *MetricsCollector) RecordTransaction(status, paymentMethod, courier string, amount float64) {
	transactionsTotal.WithLabelValues(status, paymentMethod, courier).Inc()
	transactionAmount.WithLabelValues(status, paymentMethod).Observe(amount)
}

// RecordWalletOperation records wallet operation metrics
func (m *MetricsCollector) RecordWalletOperation(operation, status string) {
	walletOperationsTotal.WithLabelValues(operation, status).Inc()
}

// UpdateWalletBalance updates wallet balance metric
func (m *MetricsCollector) UpdateWalletBalance(walletID, userType string, balance float64) {
	walletBalance.WithLabelValues(walletID, userType).Set(balance)
}

// RecordDatabaseQuery records database query metrics
func (m *MetricsCollector) RecordDatabaseQuery(operation, table string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateDatabaseConnections updates active database connections metric
func (m *MetricsCollector) UpdateDatabaseConnections(count float64) {
	dbConnectionsActive.Set(count)
}

// RecordExternalAPICall records external API call metrics
func (m *MetricsCollector) RecordExternalAPICall(service, method, statusCode string, duration time.Duration) {
	externalAPICallsTotal.WithLabelValues(service, method, statusCode).Inc()
	externalAPICallDuration.WithLabelValues(service, method, statusCode).Observe(duration.Seconds())
}

// UpdateActiveUsers updates active users metric
func (m *MetricsCollector) UpdateActiveUsers(count float64) {
	activeUsers.Set(count)
}

// RecordCacheOperation records cache operation metrics
func (m *MetricsCollector) RecordCacheOperation(operation, result string) {
	cacheOperations.WithLabelValues(operation, result).Inc()
}

// PrometheusMiddleware creates a Gin middleware for recording HTTP metrics
func PrometheusMiddleware(collector *MetricsCollector) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		endpoint := c.FullPath()

		// Better endpoint labeling for metrics cardinality control
		if endpoint == "" {
			// For unmatched routes, use a more descriptive label based on the request
			if statusCode == 404 {
				endpoint = "not_found"
			} else if method == "OPTIONS" {
				endpoint = "cors_preflight"
			} else {
				endpoint = "unknown"
			}
		}

		collector.RecordHTTPRequest(method, endpoint, strconv.Itoa(statusCode), duration)
	})
}

// GetGlobalMetricsCollector returns a global instance of MetricsCollector
var globalMetricsCollector = NewMetricsCollector()

func GetGlobalMetricsCollector() *MetricsCollector {
	return globalMetricsCollector
}
