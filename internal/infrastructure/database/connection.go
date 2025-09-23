package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

// Connect establishes a connection to the database and runs migrations
func Connect(dataSourceName string) (*sqlx.DB, error) {
	var err error
	
	// Open connection with sqlx
	db, err = sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	// Run migrations using the underlying *sql.DB
	if err := runMigrations(db.DB); err != nil {
		log.Printf("Warning: Error running migrations: %v", err)
	}

	return db, nil
}

// GetDB returns the database connection
func GetDB() *sqlx.DB {
	return db
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/infrastructure/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %v", err)
	}

	return nil
}
