package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/google/uuid"
)

// TenantType defines different tenant isolation strategies
type TenantType string

const (
	TenantTypeShared   TenantType = "shared"   // Current: Row-level isolation
	TenantTypeSchema   TenantType = "schema"   // Future: Schema per tenant
	TenantTypeDatabase TenantType = "database" // Future: Database per tenant
)

// TenantContext holds tenant information for the current request
type TenantContext struct {
	StorefrontID   uuid.UUID
	StorefrontSlug string
	SellerID       uuid.UUID
	TenantType     TenantType
}

// TenantResolver handles tenant database resolution and management
type TenantResolver interface {
	GetTenantType(ctx context.Context, storefrontID uuid.UUID) (TenantType, error)
	GetDatabaseConnection(ctx context.Context, storefrontID uuid.UUID) (*sql.DB, error)
	GetStorefrontBySlug(ctx context.Context, slug string) (*entity.Storefront, error)
	GetStorefrontByDomain(ctx context.Context, domain string) (*entity.Storefront, error)
	CreateTenantContext(storefront *entity.Storefront) *TenantContext
	
	// Tenant migration and management
	CanMigrateTenant(ctx context.Context, storefrontID uuid.UUID) (bool, TenantType, error)
	MigrateTenant(ctx context.Context, storefrontID uuid.UUID, targetType TenantType) error
	GetTenantStats(ctx context.Context, storefrontID uuid.UUID) (*TenantStats, error)
	
	// Cache management
	InvalidateStorefront(slug string)
	InvalidateStorefrontByID(storefrontID uuid.UUID)
}

// TenantStats holds metrics for tenant migration decisions
type TenantStats struct {
	StorefrontID     uuid.UUID `json:"storefront_id"`
	CustomerCount    int       `json:"customer_count"`
	OrderCount       int       `json:"order_count"`
	ProductCount     int       `json:"product_count"`
	DataSizeBytes    int64     `json:"data_size_bytes"`
	AvgQueryTime     int64     `json:"avg_query_time_ms"`
	QueriesPerSecond float64   `json:"queries_per_second"`
	StorageUsageMB   float64   `json:"storage_usage_mb"`
	ActiveSessions   int       `json:"active_sessions"`
	LastActivityAt   time.Time `json:"last_activity_at"`
}

// tenantResolver is the concrete implementation
type tenantResolver struct {
	sharedDB          *sql.DB
	tenantDBs         map[uuid.UUID]*sql.DB
	config            *TenantConfig
	cache             TenantCache
	storefrontRepo    repository.StorefrontRepository
	mu                sync.RWMutex
	statsCache        map[uuid.UUID]*cachedStats
	statsCacheMu      sync.RWMutex
}

type cachedStats struct {
	stats     *TenantStats
	expiresAt time.Time
}

// TenantConfig holds configuration for tenant resolution
type TenantConfig struct {
	DefaultTenantType        TenantType            `yaml:"default_tenant_type"`
	TenantOverrides         map[string]TenantType `yaml:"tenant_overrides"`
	SharedDatabaseURL       string                `yaml:"shared_database_url"`
	TenantDatabasePattern   string                `yaml:"tenant_database_pattern"`
	MaxConnectionsPerTenant int                   `yaml:"max_connections_per_tenant"`
	MigrationThresholds     MigrationThresholds   `yaml:"migration_thresholds"`
	CacheSettings           CacheSettings         `yaml:"cache_settings"`
}

// MigrationThresholds define when to automatically migrate tenants
type MigrationThresholds struct {
	SchemaThreshold struct {
		CustomerCount    int           `yaml:"customer_count"`
		OrderCount       int           `yaml:"order_count"`
		DataSizeMB       float64       `yaml:"data_size_mb"`
		AvgQueryTime     time.Duration `yaml:"avg_query_time"`
	} `yaml:"schema_threshold"`
	
	DatabaseThreshold struct {
		CustomerCount    int           `yaml:"customer_count"`
		OrderCount       int           `yaml:"order_count"`
		DataSizeMB       float64       `yaml:"data_size_mb"`
		AvgQueryTime     time.Duration `yaml:"avg_query_time"`
		QueriesPerSecond float64       `yaml:"queries_per_second"`
	} `yaml:"database_threshold"`
}

// CacheSettings configure tenant caching behavior
type CacheSettings struct {
	StorefrontTTL    time.Duration `yaml:"storefront_ttl"`
	StatsTTL         time.Duration `yaml:"stats_ttl"`
	MaxCacheSize     int           `yaml:"max_cache_size"`
	CleanupInterval  time.Duration `yaml:"cleanup_interval"`
}

// NewTenantResolver creates a new tenant resolver instance
func NewTenantResolver(
	sharedDB *sql.DB, 
	config *TenantConfig, 
	cache TenantCache,
	storefrontRepo repository.StorefrontRepository,
) TenantResolver {
	resolver := &tenantResolver{
		sharedDB:       sharedDB,
		tenantDBs:      make(map[uuid.UUID]*sql.DB),
		config:         config,
		cache:          cache,
		storefrontRepo: storefrontRepo,
		statsCache:     make(map[uuid.UUID]*cachedStats),
	}
	
	// Start background cleanup goroutine
	go resolver.startCleanup()
	
	return resolver
}

// GetTenantType determines the appropriate tenant isolation strategy
func (tr *tenantResolver) GetTenantType(ctx context.Context, storefrontID uuid.UUID) (TenantType, error) {
	// Check explicit configuration overrides first
	if tenantType, exists := tr.config.TenantOverrides[storefrontID.String()]; exists {
		return tenantType, nil
	}
	
	// Check automatic migration thresholds
	stats, err := tr.GetTenantStats(ctx, storefrontID)
	if err != nil {
		// If we can't get stats, default to shared
		return tr.config.DefaultTenantType, nil
	}
	
	// Apply automatic migration rules
	if tr.shouldMigrateToDatabase(stats) {
		return TenantTypeDatabase, nil
	}
	
	if tr.shouldMigrateToSchema(stats) {
		return TenantTypeSchema, nil
	}
	
	return tr.config.DefaultTenantType, nil
}

// GetDatabaseConnection returns the appropriate database connection for a tenant
func (tr *tenantResolver) GetDatabaseConnection(ctx context.Context, storefrontID uuid.UUID) (*sql.DB, error) {
	tenantType, err := tr.GetTenantType(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant type: %w", err)
	}
	
	switch tenantType {
	case TenantTypeShared, TenantTypeSchema:
		return tr.sharedDB, nil
	case TenantTypeDatabase:
		return tr.getTenantDatabase(storefrontID)
	default:
		return nil, fmt.Errorf("unsupported tenant type: %s", tenantType)
	}
}

// GetStorefrontBySlug retrieves a storefront by its slug with caching
func (tr *tenantResolver) GetStorefrontBySlug(ctx context.Context, slug string) (*entity.Storefront, error) {
	// Try cache first
	if storefront := tr.cache.GetStorefront(slug); storefront != nil {
		return storefront, nil
	}
	
	// Query from repository
	storefront, err := tr.storefrontRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	
	if storefront != nil {
		// Cache for configured TTL
		tr.cache.SetStorefront(slug, storefront, tr.config.CacheSettings.StorefrontTTL)
	}
	
	return storefront, nil
}

// GetStorefrontByDomain retrieves a storefront by its domain with caching
func (tr *tenantResolver) GetStorefrontByDomain(ctx context.Context, domain string) (*entity.Storefront, error) {
	// Try cache first (using domain as key)
	cacheKey := "domain:" + domain
	if storefront := tr.cache.GetStorefront(cacheKey); storefront != nil {
		return storefront, nil
	}
	
	// Query from repository
	storefront, err := tr.storefrontRepo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	
	if storefront != nil {
		// Cache with domain key
		tr.cache.SetStorefront(cacheKey, storefront, tr.config.CacheSettings.StorefrontTTL)
		// Also cache with slug key
		tr.cache.SetStorefront(storefront.Slug, storefront, tr.config.CacheSettings.StorefrontTTL)
	}
	
	return storefront, nil
}

// CreateTenantContext creates a tenant context for a storefront
func (tr *tenantResolver) CreateTenantContext(storefront *entity.Storefront) *TenantContext {
	tenantType, _ := tr.GetTenantType(context.Background(), storefront.ID)
	
	return &TenantContext{
		StorefrontID:   storefront.ID,
		StorefrontSlug: storefront.Slug,
		SellerID:       storefront.SellerID,
		TenantType:     tenantType,
	}
}

// CanMigrateTenant checks if a tenant can be migrated to a different isolation strategy
func (tr *tenantResolver) CanMigrateTenant(ctx context.Context, storefrontID uuid.UUID) (bool, TenantType, error) {
	stats, err := tr.GetTenantStats(ctx, storefrontID)
	if err != nil {
		return false, TenantTypeShared, err
	}
	
	currentType, err := tr.GetTenantType(ctx, storefrontID)
	if err != nil {
		return false, TenantTypeShared, err
	}
	
	// Check if migration to database is needed
	if currentType != TenantTypeDatabase && tr.shouldMigrateToDatabase(stats) {
		return true, TenantTypeDatabase, nil
	}
	
	// Check if migration to schema is needed
	if currentType == TenantTypeShared && tr.shouldMigrateToSchema(stats) {
		return true, TenantTypeSchema, nil
	}
	
	return false, currentType, nil
}

// MigrateTenant migrates a tenant to a different isolation strategy (placeholder implementation)
func (tr *tenantResolver) MigrateTenant(ctx context.Context, storefrontID uuid.UUID, targetType TenantType) error {
	// This is a complex operation that would involve:
	// 1. Creating new schema/database
	// 2. Copying data
	// 3. Updating configuration
	// 4. Switching connections
	// 5. Cleaning up old data
	
	// For now, just update the override configuration
	tr.config.TenantOverrides[storefrontID.String()] = targetType
	
	// Invalidate caches
	tr.InvalidateStorefrontByID(storefrontID)
	
	return nil // Placeholder implementation
}

// GetTenantStats retrieves comprehensive tenant statistics with caching
func (tr *tenantResolver) GetTenantStats(ctx context.Context, storefrontID uuid.UUID) (*TenantStats, error) {
	// Check cache first
	tr.statsCacheMu.RLock()
	if cached, exists := tr.statsCache[storefrontID]; exists && time.Now().Before(cached.expiresAt) {
		tr.statsCacheMu.RUnlock()
		return cached.stats, nil
	}
	tr.statsCacheMu.RUnlock()
	
	// Query stats from repository
	storefrontStats, err := tr.storefrontRepo.GetStorefrontStats(ctx, storefrontID)
	if err != nil {
		return nil, err
	}
	
	// Convert to TenantStats
	stats := &TenantStats{
		StorefrontID:     storefrontID,
		CustomerCount:    storefrontStats.CustomerCount,
		OrderCount:       storefrontStats.OrderCount,
		AvgQueryTime:     storefrontStats.AvgQueryTime,
		ActiveSessions:   storefrontStats.ActiveSessions,
		// Additional fields would need to be computed from various sources
		DataSizeBytes:    0, // Placeholder
		QueriesPerSecond: 0, // Placeholder
		StorageUsageMB:   0, // Placeholder
		LastActivityAt:   time.Now(),
	}
	
	// Cache the stats
	tr.statsCacheMu.Lock()
	tr.statsCache[storefrontID] = &cachedStats{
		stats:     stats,
		expiresAt: time.Now().Add(tr.config.CacheSettings.StatsTTL),
	}
	tr.statsCacheMu.Unlock()
	
	return stats, nil
}

// InvalidateStorefront removes a storefront from cache
func (tr *tenantResolver) InvalidateStorefront(slug string) {
	tr.cache.InvalidateStorefront(slug)
}

// InvalidateStorefrontByID removes a storefront from cache by ID
func (tr *tenantResolver) InvalidateStorefrontByID(storefrontID uuid.UUID) {
	// Also clear stats cache
	tr.statsCacheMu.Lock()
	delete(tr.statsCache, storefrontID)
	tr.statsCacheMu.Unlock()
	
	// We'd need to reverse lookup the slug to clear the cache
	// This is a limitation of the current cache design
}

// Helper methods

// getTenantDatabase retrieves or creates a database connection for a specific tenant
func (tr *tenantResolver) getTenantDatabase(storefrontID uuid.UUID) (*sql.DB, error) {
	tr.mu.RLock()
	if db, exists := tr.tenantDBs[storefrontID]; exists {
		tr.mu.RUnlock()
		return db, nil
	}
	tr.mu.RUnlock()
	
	// Create new connection
	tr.mu.Lock()
	defer tr.mu.Unlock()
	
	// Double-check pattern
	if db, exists := tr.tenantDBs[storefrontID]; exists {
		return db, nil
	}
	
	// Generate database URL for this tenant
	dbURL := fmt.Sprintf(tr.config.TenantDatabasePattern, storefrontID.String())
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(tr.config.MaxConnectionsPerTenant)
	db.SetMaxIdleConns(tr.config.MaxConnectionsPerTenant / 4)
	db.SetConnMaxLifetime(time.Hour)
	
	// Test the connection
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping tenant database: %w", err)
	}
	
	tr.tenantDBs[storefrontID] = db
	return db, nil
}

// shouldMigrateToDatabase checks if tenant should be migrated to database isolation
func (tr *tenantResolver) shouldMigrateToDatabase(stats *TenantStats) bool {
	threshold := tr.config.MigrationThresholds.DatabaseThreshold
	
	return stats.CustomerCount > threshold.CustomerCount ||
		   stats.OrderCount > threshold.OrderCount ||
		   stats.StorageUsageMB > threshold.DataSizeMB ||
		   time.Duration(stats.AvgQueryTime)*time.Millisecond > threshold.AvgQueryTime ||
		   stats.QueriesPerSecond > threshold.QueriesPerSecond
}

// shouldMigrateToSchema checks if tenant should be migrated to schema isolation
func (tr *tenantResolver) shouldMigrateToSchema(stats *TenantStats) bool {
	threshold := tr.config.MigrationThresholds.SchemaThreshold
	
	return stats.CustomerCount > threshold.CustomerCount ||
		   stats.OrderCount > threshold.OrderCount ||
		   stats.StorageUsageMB > threshold.DataSizeMB ||
		   time.Duration(stats.AvgQueryTime)*time.Millisecond > threshold.AvgQueryTime
}

// startCleanup starts a background goroutine to cleanup expired cache entries
func (tr *tenantResolver) startCleanup() {
	ticker := time.NewTicker(tr.config.CacheSettings.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		
		// Clean up stats cache
		tr.statsCacheMu.Lock()
		for id, cached := range tr.statsCache {
			if now.After(cached.expiresAt) {
				delete(tr.statsCache, id)
			}
		}
		tr.statsCacheMu.Unlock()
		
		// Clean up database connections that haven't been used
		tr.mu.Lock()
		for storefrontID, db := range tr.tenantDBs {
			// Check if connection is still valid and close inactive ones
			if err := db.Ping(); err != nil {
				db.Close()
				delete(tr.tenantDBs, storefrontID)
			}
		}
		tr.mu.Unlock()
	}
}