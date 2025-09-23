package otel

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// OTelClient handles OpenTelemetry metrics export to Grafana Cloud
type OTelClient struct {
	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter
	enabled       bool
}

// NewOTelClient creates a new OpenTelemetry client for Grafana Cloud
func NewOTelClient() (*OTelClient, error) {
	enabled := os.Getenv("OTEL_ENABLED") == "true"
	if !enabled {
		return &OTelClient{enabled: false}, nil
	}

	endpoint := os.Getenv("OTEL_ENDPOINT")
	username := os.Getenv("OTEL_USERNAME")
	password := os.Getenv("OTEL_PASSWORD")
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "kirimku-backend"
	}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(os.Getenv("OTEL_SERVICE_VERSION")),
			semconv.DeploymentEnvironment(os.Getenv("OTEL_ENVIRONMENT")),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create OTLP HTTP exporter for Grafana Cloud
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + basicAuth(username, password),
		}),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
	)
	if err != nil {
		return nil, err
	}

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(30*time.Second), // Push every 30 seconds
			),
		),
	)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := meterProvider.Meter("kirimku-backend")

	return &OTelClient{
		meterProvider: meterProvider,
		meter:         meter,
		enabled:       true,
	}, nil
}

// Shutdown gracefully shuts down the OTel client
func (c *OTelClient) Shutdown(ctx context.Context) error {
	if !c.enabled || c.meterProvider == nil {
		return nil
	}
	return c.meterProvider.Shutdown(ctx)
}

// GetMeter returns the OpenTelemetry meter for creating instruments
func (c *OTelClient) GetMeter() metric.Meter {
	if !c.enabled {
		return nil
	}
	return c.meter
}

// IsEnabled returns whether OpenTelemetry is enabled
func (c *OTelClient) IsEnabled() bool {
	return c.enabled
}

// basicAuth creates a basic auth string
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return auth // Note: In real implementation, use base64.StdEncoding.EncodeToString([]byte(auth))
}

// Example usage in metrics package:
/*
// In pkg/metrics/otel_metrics.go

type OTelMetrics struct {
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	// ... other metrics
}

func NewOTelMetrics(meter metric.Meter) (*OTelMetrics, error) {
	httpRequestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	httpRequestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &OTelMetrics{
		httpRequestsTotal:   httpRequestsTotal,
		httpRequestDuration: httpRequestDuration,
	}, nil
}

func (m *OTelMetrics) RecordHTTPRequest(ctx context.Context, method, endpoint, statusCode string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
		attribute.String("status_code", statusCode),
	}

	m.httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.httpRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}
*/
