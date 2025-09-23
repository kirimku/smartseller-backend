package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Set up test environment variables
	testEnv := map[string]string{
		"SESSION_KEY":          "test-session-key",
		"DB_HOST":              "localhost",
		"DB_PORT":              "5432",
		"DB_USER":              "testuser",
		"DB_PASSWORD":          "testpass",
		"DB_NAME":              "testdb",
		"DB_SSL_MODE":          "disable",
		"DB_MAX_OPEN_CONNS":    "25",
		"DB_MAX_IDLE_CONNS":    "25",
		"DB_CONN_MAX_LIFETIME": "300",
	}

	// Set environment variables
	for k, v := range testEnv {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Call LoadConfig
	err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	// Verify database configuration
	if AppConfig.Database.MaxOpenConns != 25 {
		t.Errorf("Expected MaxOpenConns to be 25, got %v", AppConfig.Database.MaxOpenConns)
	}
	if AppConfig.Database.MaxIdleConns != 25 {
		t.Errorf("Expected MaxIdleConns to be 25, got %v", AppConfig.Database.MaxIdleConns)
	}
	expectedLifetime := time.Duration(300) * time.Second
	if AppConfig.Database.MaxLifetime != expectedLifetime {
		t.Errorf("Expected MaxLifetime to be %v, got %v", expectedLifetime, AppConfig.Database.MaxLifetime)
	}
}
