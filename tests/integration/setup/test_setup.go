package setup

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	DatabaseURL      string
	BaseURL          string
	MigrationsPath   string
	TestDataPath     string
	CleanupAfterTest bool
}

// TestSetup provides database and environment setup for integration tests
type TestSetup struct {
	Config *TestConfig
	DB     *sql.DB
	Auth   *AuthHelper
}

// LoadTestConfig loads test configuration from environment variables
func LoadTestConfig() *TestConfig {
	config := &TestConfig{
		DatabaseURL:      getEnv("TEST_DATABASE_URL", "postgres://smartseller_user:smartseller_pass@localhost:5432/smartseller_db?sslmode=disable"),
		BaseURL:          getEnv("TEST_BASE_URL", "http://localhost:8090"),
		MigrationsPath:   getEnv("TEST_MIGRATIONS_PATH", "file://../../migrations"),
		TestDataPath:     getEnv("TEST_DATA_PATH", "./testdata"),
		CleanupAfterTest: getEnv("CLEANUP_AFTER_TEST", "false") == "true",
	}
	return config
}

// NewTestSetup creates a new test setup instance
func NewTestSetup() (*TestSetup, error) {
	config := LoadTestConfig()

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	auth := NewAuthHelper(config.BaseURL)

	setup := &TestSetup{
		Config: config,
		DB:     db,
		Auth:   auth,
	}

	return setup, nil
}

// SetupDatabase runs migrations and seeds test data
func (ts *TestSetup) SetupDatabase() error {
	// Run migrations
	if err := ts.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed test data
	if err := ts.seedTestData(); err != nil {
		return fmt.Errorf("failed to seed test data: %w", err)
	}

	return nil
}

// CleanupDatabase drops all tables and recreates them
func (ts *TestSetup) CleanupDatabase() error {
	// Drop all tables
	if err := ts.dropAllTables(); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Re-run migrations
	if err := ts.runMigrations(); err != nil {
		return fmt.Errorf("failed to re-run migrations: %w", err)
	}

	return nil
}

// SetupTest prepares for individual test execution
func (ts *TestSetup) SetupTest(t *testing.T) {
	t.Helper()

	// Begin transaction for test isolation (if needed)
	// This can be implemented if we want transaction-based test isolation

	// Authenticate test user
	if err := ts.Auth.LoginWithTestUser(); err != nil {
		t.Fatalf("Failed to authenticate test user: %v", err)
	}
}

// TeardownTest cleans up after individual test
func (ts *TestSetup) TeardownTest(t *testing.T) {
	t.Helper()

	if ts.Config.CleanupAfterTest {
		// Clean up test data created during the test
		ts.cleanupTestData()
	}
}

// Close closes database connection and cleans up resources
func (ts *TestSetup) Close() error {
	if ts.DB != nil {
		return ts.DB.Close()
	}
	return nil
}

// runMigrations executes database migrations
func (ts *TestSetup) runMigrations() error {
	driver, err := postgres.WithInstance(ts.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(ts.Config.MigrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// dropAllTables drops all tables in the database
func (ts *TestSetup) dropAllTables() error {
	// Get all table names
	rows, err := ts.DB.Query(`
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public' AND tablename != 'schema_migrations'
	`)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	// Drop all tables
	for _, table := range tables {
		if _, err := ts.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// seedTestData loads initial test data
func (ts *TestSetup) seedTestData() error {
	// Create test users
	if err := ts.createTestUsers(); err != nil {
		return fmt.Errorf("failed to create test users: %w", err)
	}

	// Create test categories
	if err := ts.createTestCategories(); err != nil {
		return fmt.Errorf("failed to create test categories: %w", err)
	}

	return nil
}

// createTestUsers creates test users for authentication
func (ts *TestSetup) createTestUsers() error {
	testUsers := []struct {
		email    string
		password string
		role     string
	}{
		{"testuser@example.com", "testpassword123", "user"},
		{"admin@example.com", "adminpassword123", "admin"},
	}

	for _, user := range testUsers {
		_, err := ts.DB.Exec(`
			INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, crypt($2, gen_salt('bf')), $3, NOW(), NOW())
			ON CONFLICT (email) DO NOTHING
		`, user.email, user.password, user.role)

		if err != nil {
			return fmt.Errorf("failed to create test user %s: %w", user.email, err)
		}
	}

	return nil
}

// createTestCategories creates test product categories
func (ts *TestSetup) createTestCategories() error {
	categories := []struct {
		name        string
		slug        string
		description string
		parentSlug  *string
	}{
		{"Electronics", "electronics", "Electronic devices and accessories", nil},
		{"Clothing", "clothing", "Clothing and fashion items", nil},
		{"Books", "books", "Books and educational materials", nil},
		{"Smartphones", "smartphones", "Mobile phones and accessories", stringPtr("electronics")},
		{"Laptops", "laptops", "Laptops and computers", stringPtr("electronics")},
		{"Men's Clothing", "mens-clothing", "Clothing for men", stringPtr("clothing")},
		{"Women's Clothing", "womens-clothing", "Clothing for women", stringPtr("clothing")},
	}

	// Create root categories first
	for _, cat := range categories {
		if cat.parentSlug == nil {
			_, err := ts.DB.Exec(`
				INSERT INTO product_categories (id, name, slug, description, created_at, updated_at)
				VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
				ON CONFLICT (slug) DO NOTHING
			`, cat.name, cat.slug, cat.description)

			if err != nil {
				return fmt.Errorf("failed to create category %s: %w", cat.name, err)
			}
		}
	}

	// Create child categories
	for _, cat := range categories {
		if cat.parentSlug != nil {
			_, err := ts.DB.Exec(`
				INSERT INTO product_categories (id, name, slug, description, parent_id, created_at, updated_at)
				SELECT gen_random_uuid(), $1, $2, $3, pc.id, NOW(), NOW()
				FROM product_categories pc
				WHERE pc.slug = $4
				ON CONFLICT (slug) DO NOTHING
			`, cat.name, cat.slug, cat.description, *cat.parentSlug)

			if err != nil {
				return fmt.Errorf("failed to create child category %s: %w", cat.name, err)
			}
		}
	}

	return nil
}

// cleanupTestData removes data created during tests
func (ts *TestSetup) cleanupTestData() {
	// Clean up in reverse order of dependencies
	tables := []string{
		"product_images",
		"product_variants",
		"product_variant_options",
		"products",
		// Don't clean categories and users as they're seed data
	}

	for _, table := range tables {
		if _, err := ts.DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at > NOW() - INTERVAL '1 hour'", table)); err != nil {
			log.Printf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
}

// GetTestProductCategory returns a test category ID by slug
func (ts *TestSetup) GetTestProductCategory(slug string) (string, error) {
	var categoryID string
	err := ts.DB.QueryRow("SELECT id FROM product_categories WHERE slug = $1", slug).Scan(&categoryID)
	if err != nil {
		return "", fmt.Errorf("failed to get category %s: %w", slug, err)
	}
	return categoryID, nil
}

// CreateTestProduct creates a test product and returns its ID
func (ts *TestSetup) CreateTestProduct(name, sku string, price float64, categorySlug string) (string, error) {
	categoryID, err := ts.GetTestProductCategory(categorySlug)
	if err != nil {
		return "", err
	}

	var productID string
	err = ts.DB.QueryRow(`
		INSERT INTO products (id, name, sku, price, category_id, status, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 'active', NOW(), NOW())
		RETURNING id
	`, name, sku, price, categoryID).Scan(&productID)

	if err != nil {
		return "", fmt.Errorf("failed to create test product: %w", err)
	}

	return productID, nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func stringPtr(s string) *string {
	return &s
}

// TestSuiteSetup sets up the entire test suite
func TestSuiteSetup() (*TestSetup, error) {
	setup, err := NewTestSetup()
	if err != nil {
		return nil, fmt.Errorf("failed to create test setup: %w", err)
	}

	if err := setup.SetupDatabase(); err != nil {
		setup.Close()
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return setup, nil
}

// TestSuiteTeardown cleans up after the entire test suite
func TestSuiteTeardown(setup *TestSetup) error {
	if setup != nil {
		if err := setup.CleanupDatabase(); err != nil {
			log.Printf("Warning: failed to cleanup database: %v", err)
		}
		return setup.Close()
	}
	return nil
}
