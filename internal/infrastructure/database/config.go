package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/logger"
)

type DBConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	DatabaseName      string
	SSLMode           string
	MaxOpenConns      int
	MaxIdleConns      int
	MaxLifetime       time.Duration
	HealthCheckPeriod time.Duration
}

func NewDBConfig() (*DBConfig, error) {
	config := &DBConfig{
		Host:              getEnvWithDefault("DB_HOST", "localhost"),
		Port:              getEnvAsInt("DB_PORT", 5432),
		User:              getEnvWithDefault("DB_USER", "postgres"),
		Password:          os.Getenv("DB_PASSWORD"),
		DatabaseName:      getEnvWithDefault("DB_NAME", "kirimku"),
		SSLMode:           getEnvWithDefault("DB_SSL_MODE", "disable"),
		MaxOpenConns:      getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:      getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
		MaxLifetime:       time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second,
		HealthCheckPeriod: time.Duration(getEnvAsInt("DB_HEALTH_CHECK_PERIOD", 30)) * time.Second,
	}

	return config, config.validate()
}

func (c *DBConfig) validate() error {
	if c.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DatabaseName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	return nil
}

func (c *DBConfig) ConnectDB() (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DatabaseName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.MaxLifetime)

	// Start connection pool monitoring
	go c.monitorConnectionPool(db)

	return db, nil
}

func (c *DBConfig) monitorConnectionPool(db *sqlx.DB) {
	ticker := time.NewTicker(c.HealthCheckPeriod)
	defer ticker.Stop()

	for range ticker.C {
		stats := db.Stats()
		logger.DBLogger().
			Int("max_open_conns", c.MaxOpenConns).
			Int("max_idle_conns", c.MaxIdleConns).
			Int("open_connections", stats.OpenConnections).
			Int("in_use", stats.InUse).
			Int("idle", stats.Idle).
			Dur("max_lifetime", c.MaxLifetime).
			Msg("Database connection pool stats")

		if err := db.Ping(); err != nil {
			logger.ErrorLogger().
				Err(err).
				Msg("Database connection pool health check failed")
		}
	}
}

func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvWithDefault(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
