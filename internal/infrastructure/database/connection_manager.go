package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
)

// ConnectionManager manages database connections for multi-tenant architecture
type ConnectionManager struct {
	// Shared database connection (for shared table strategy)
	sharedDB *sqlx.DB

	// Per-tenant database connections (for database-per-tenant strategy)
	tenantDBs map[string]*sqlx.DB
	mutex     sync.RWMutex

	// Configuration
	config *DatabaseConfig

	// Health monitoring
	healthChecker *HealthChecker

	// Connection pool settings
	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	// Shared database configuration
	SharedDatabase *DatabaseConnectionConfig `json:"shared_database"`

	// Per-tenant database configurations
	TenantDatabases map[string]*DatabaseConnectionConfig `json:"tenant_databases"`

	// Connection pool settings
	MaxIdleConns    int           `json:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`

	// Health check settings
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	ConnectionTimeout   time.Duration `json:"connection_timeout"`
}

// DatabaseConnectionConfig represents individual database connection configuration
type DatabaseConnectionConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`

	// Additional connection parameters
	MaxConnections   int           `json:"max_connections"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	ConnectTimeout   time.Duration `json:"connect_timeout"`
	StatementTimeout time.Duration `json:"statement_timeout"`
}

// ConnectionStatus represents the status of a database connection
type ConnectionStatus string

const (
	ConnectionStatusActive       ConnectionStatus = "active"
	ConnectionStatusIdle         ConnectionStatus = "idle"
	ConnectionStatusUnhealthy    ConnectionStatus = "unhealthy"
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
)

// NewConnectionManager creates a new database connection manager
func NewConnectionManager(config *DatabaseConfig) (*ConnectionManager, error) {
	cm := &ConnectionManager{
		tenantDBs:       make(map[string]*sqlx.DB),
		config:          config,
		maxIdleConns:    config.MaxIdleConns,
		maxOpenConns:    config.MaxOpenConns,
		connMaxLifetime: config.ConnMaxLifetime,
		connMaxIdleTime: config.ConnMaxIdleTime,
	}

	// Set default values if not provided
	if cm.maxIdleConns == 0 {
		cm.maxIdleConns = 10
	}
	if cm.maxOpenConns == 0 {
		cm.maxOpenConns = 100
	}
	if cm.connMaxLifetime == 0 {
		cm.connMaxLifetime = time.Hour
	}
	if cm.connMaxIdleTime == 0 {
		cm.connMaxIdleTime = 10 * time.Minute
	}

	// Initialize shared database if configured
	if config.SharedDatabase != nil {
		sharedDB, err := cm.createConnection(config.SharedDatabase, "shared")
		if err != nil {
			return nil, fmt.Errorf("failed to create shared database connection: %w", err)
		}
		cm.sharedDB = sharedDB
	}

	// Initialize tenant databases
	for tenantID, dbConfig := range config.TenantDatabases {
		tenantDB, err := cm.createConnection(dbConfig, tenantID)
		if err != nil {
			continue // Skip failed connections, will be retried later
		}
		cm.tenantDBs[tenantID] = tenantDB
	}

	// Initialize health checker
	cm.healthChecker = NewHealthChecker(cm, config.HealthCheckInterval)

	return cm, nil
}

// GetConnection returns the appropriate database connection for the tenant context
func (cm *ConnectionManager) GetConnection(ctx context.Context, tenantContext *tenant.TenantContext) (*sqlx.DB, error) {
	if tenantContext == nil {
		if cm.sharedDB != nil {
			return cm.sharedDB, nil
		}
		return nil, fmt.Errorf("no tenant context provided and no shared database configured")
	}

	switch tenantContext.TenantType {
	case tenant.TenantTypeShared:
		if cm.sharedDB == nil {
			return nil, fmt.Errorf("shared database not configured for tenant: %s", tenantContext.StorefrontSlug)
		}
		return cm.sharedDB, nil

	case tenant.TenantTypeDatabase:
		cm.mutex.RLock()
		tenantDB, exists := cm.tenantDBs[tenantContext.StorefrontID.String()]
		cm.mutex.RUnlock()

		if !exists {
			// Try to establish connection for new tenant
			return cm.createTenantConnection(tenantContext.StorefrontID.String())
		}

		// Check connection health
		if err := cm.pingConnection(tenantDB); err != nil {
			return cm.recreateTenantConnection(tenantContext.StorefrontID.String())
		}

		return tenantDB, nil

	case tenant.TenantTypeSchema:
		// For separate schema, use shared DB but context contains schema info
		if cm.sharedDB == nil {
			return nil, fmt.Errorf("shared database required for schema-per-tenant isolation")
		}
		return cm.sharedDB, nil

	default:
		return nil, fmt.Errorf("unsupported tenant type: %s", tenantContext.TenantType)
	}
}

// GetSharedConnection returns the shared database connection
func (cm *ConnectionManager) GetSharedConnection() *sqlx.DB {
	return cm.sharedDB
}

// GetTenantConnection returns a specific tenant's database connection
func (cm *ConnectionManager) GetTenantConnection(tenantID string) (*sqlx.DB, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	db, exists := cm.tenantDBs[tenantID]
	return db, exists
}

// AddTenantDatabase adds a new tenant database connection
func (cm *ConnectionManager) AddTenantDatabase(tenantID string, config *DatabaseConnectionConfig) error {
	tenantDB, err := cm.createConnection(config, tenantID)
	if err != nil {
		return fmt.Errorf("failed to create tenant database connection: %w", err)
	}

	cm.mutex.Lock()
	cm.tenantDBs[tenantID] = tenantDB
	cm.mutex.Unlock()

	return nil
}

// RemoveTenantDatabase removes a tenant database connection
func (cm *ConnectionManager) RemoveTenantDatabase(tenantID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	db, exists := cm.tenantDBs[tenantID]
	if !exists {
		return fmt.Errorf("tenant database not found: %s", tenantID)
	}

	if err := db.Close(); err != nil {
		// Log error but continue with removal
	}

	delete(cm.tenantDBs, tenantID)
	return nil
}

// BeginTx begins a transaction with the appropriate database connection
func (cm *ConnectionManager) BeginTx(ctx context.Context, tenantContext *tenant.TenantContext) (*sqlx.Tx, error) {
	db, err := cm.GetConnection(ctx, tenantContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

// ExecuteInTransaction executes a function within a database transaction
func (cm *ConnectionManager) ExecuteInTransaction(ctx context.Context, tenantContext *tenant.TenantContext, fn func(*sqlx.Tx) error) error {
	tx, err := cm.BeginTx(ctx, tenantContext)
	if err != nil {
		return err
	}
	defer tx.Rollback() // This is safe to call even after commit

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// GetConnectionStats returns statistics about database connections
func (cm *ConnectionManager) GetConnectionStats() map[string]*ConnectionStats {
	stats := make(map[string]*ConnectionStats)

	// Shared database stats
	if cm.sharedDB != nil {
		stats["shared"] = cm.getDBStats(cm.sharedDB, "shared", true)
	}

	// Tenant database stats
	cm.mutex.RLock()
	for tenantID, db := range cm.tenantDBs {
		stats[tenantID] = cm.getDBStats(db, tenantID, false)
	}
	cm.mutex.RUnlock()

	return stats
}

// GetHealthStatus returns the health status of all connections
func (cm *ConnectionManager) GetHealthStatus() *HealthStatus {
	if cm.healthChecker != nil {
		return cm.healthChecker.GetStatus()
	}
	return &HealthStatus{
		Status:      "unknown",
		LastChecked: time.Now(),
		Connections: make(map[string]*ConnectionHealth),
	}
}

// Close closes all database connections
func (cm *ConnectionManager) Close() error {
	var errors []error

	// Stop health checker
	if cm.healthChecker != nil {
		cm.healthChecker.Stop()
	}

	// Close shared connection
	if cm.sharedDB != nil {
		if err := cm.sharedDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close shared database: %w", err))
		}
	}

	// Close tenant connections
	cm.mutex.Lock()
	for tenantID, db := range cm.tenantDBs {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close tenant database %s: %w", tenantID, err))
		}
	}
	cm.tenantDBs = make(map[string]*sqlx.DB)
	cm.mutex.Unlock()

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	return nil
}

// Private methods

func (cm *ConnectionManager) createConnection(config *DatabaseConnectionConfig, identifier string) (*sqlx.DB, error) {
	dsn := cm.buildDSN(config)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxIdleConns(cm.maxIdleConns)
	db.SetMaxOpenConns(config.MaxConnections)
	if config.MaxConnections == 0 {
		db.SetMaxOpenConns(cm.maxOpenConns)
	}
	db.SetConnMaxLifetime(cm.connMaxLifetime)
	db.SetConnMaxIdleTime(cm.connMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func (cm *ConnectionManager) buildDSN(config *DatabaseConnectionConfig) string {
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
		sslMode,
	)

	// Add additional connection parameters
	if config.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(config.ConnectTimeout.Seconds()))
	}
	if config.StatementTimeout > 0 {
		dsn += fmt.Sprintf(" statement_timeout=%d", int(config.StatementTimeout.Milliseconds()))
	}

	return dsn
}

func (cm *ConnectionManager) createTenantConnection(tenantID string) (*sqlx.DB, error) {
	// This would typically load tenant database configuration from a configuration service
	// For now, return an error indicating the tenant is not configured
	return nil, fmt.Errorf("tenant database not configured: %s", tenantID)
}

func (cm *ConnectionManager) recreateTenantConnection(tenantID string) (*sqlx.DB, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Close existing connection
	if oldDB, exists := cm.tenantDBs[tenantID]; exists {
		oldDB.Close()
		delete(cm.tenantDBs, tenantID)
	}

	// This would recreate the connection using stored configuration
	// For now, return an error
	return nil, fmt.Errorf("failed to recreate tenant connection: %s", tenantID)
}

func (cm *ConnectionManager) pingConnection(db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

func (cm *ConnectionManager) getDBStats(db *sqlx.DB, identifier string, isShared bool) *ConnectionStats {
	sqlStats := db.Stats()

	return &ConnectionStats{
		Identifier:         identifier,
		IsShared:           isShared,
		OpenConnections:    sqlStats.OpenConnections,
		InUse:              sqlStats.InUse,
		Idle:               sqlStats.Idle,
		WaitCount:          sqlStats.WaitCount,
		WaitDuration:       sqlStats.WaitDuration,
		MaxIdleClosed:      sqlStats.MaxIdleClosed,
		MaxLifetimeClosed:  sqlStats.MaxLifetimeClosed,
		MaxOpenConnections: sqlStats.MaxOpenConnections,
	}
}

// Supporting types

type ConnectionStats struct {
	Identifier         string        `json:"identifier"`
	IsShared           bool          `json:"is_shared"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
	MaxOpenConnections int           `json:"max_open_connections"`
}

type HealthStatus struct {
	Status      string                       `json:"status"`
	LastChecked time.Time                    `json:"last_checked"`
	Connections map[string]*ConnectionHealth `json:"connections"`
}

type ConnectionHealth struct {
	Status       ConnectionStatus `json:"status"`
	LastChecked  time.Time        `json:"last_checked"`
	ResponseTime time.Duration    `json:"response_time"`
	Error        string           `json:"error,omitempty"`
}
