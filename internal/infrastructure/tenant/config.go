package tenant

import (
	"time"
)

// DefaultTenantConfig returns a default configuration for the tenant resolver
func DefaultTenantConfig() *TenantConfig {
	return &TenantConfig{
		DefaultTenantType:        TenantTypeShared,
		TenantOverrides:         make(map[string]TenantType),
		SharedDatabaseURL:       "postgres://localhost:5432/smartseller?sslmode=disable",
		TenantDatabasePattern:   "postgres://localhost:5432/smartseller_tenant_%s?sslmode=disable",
		MaxConnectionsPerTenant: 10,
		MigrationThresholds: MigrationThresholds{
			SchemaThreshold: struct {
				CustomerCount    int           `yaml:"customer_count"`
				OrderCount       int           `yaml:"order_count"`
				DataSizeMB       float64       `yaml:"data_size_mb"`
				AvgQueryTime     time.Duration `yaml:"avg_query_time"`
			}{
				CustomerCount: 1000,
				OrderCount:    5000,
				DataSizeMB:    100.0,
				AvgQueryTime:  100 * time.Millisecond,
			},
			DatabaseThreshold: struct {
				CustomerCount    int           `yaml:"customer_count"`
				OrderCount       int           `yaml:"order_count"`
				DataSizeMB       float64       `yaml:"data_size_mb"`
				AvgQueryTime     time.Duration `yaml:"avg_query_time"`
				QueriesPerSecond float64       `yaml:"queries_per_second"`
			}{
				CustomerCount:    10000,
				OrderCount:       50000,
				DataSizeMB:       1000.0,
				AvgQueryTime:     200 * time.Millisecond,
				QueriesPerSecond: 100.0,
			},
		},
		CacheSettings: CacheSettings{
			StorefrontTTL:   1 * time.Hour,
			StatsTTL:        15 * time.Minute,
			MaxCacheSize:    1000,
			CleanupInterval: 5 * time.Minute,
		},
	}
}

// LoadTenantConfigFromEnv loads tenant configuration from environment variables
func LoadTenantConfigFromEnv() *TenantConfig {
	// This function would load configuration from environment variables
	// For now, return default config
	return DefaultTenantConfig()
}

// TenantMiddlewareConfig holds configuration for tenant middleware
type TenantMiddlewareConfig struct {
	// Header names to check for tenant identification
	TenantHeaders []string `yaml:"tenant_headers"`
	
	// Default tenant to use if none is found
	DefaultTenant string `yaml:"default_tenant"`
	
	// Whether to require tenant identification
	RequireTenant bool `yaml:"require_tenant"`
	
	// Timeout for tenant resolution
	ResolutionTimeout time.Duration `yaml:"resolution_timeout"`
}

// DefaultTenantMiddlewareConfig returns default middleware configuration
func DefaultTenantMiddlewareConfig() *TenantMiddlewareConfig {
	return &TenantMiddlewareConfig{
		TenantHeaders: []string{
			"X-Storefront-Slug",
			"X-Storefront-Domain",
			"X-Tenant-ID",
		},
		DefaultTenant:     "",
		RequireTenant:     true,
		ResolutionTimeout: 5 * time.Second,
	}
}