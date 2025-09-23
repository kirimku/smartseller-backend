package database

import (
	"testing"
)

func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		dbURL       string
		shouldError bool
	}{
		{
			name:        "invalid connection string",
			dbURL:       "invalid-url",
			shouldError: true,
		},
		{
			name:        "empty connection string",
			dbURL:       "",
			shouldError: true,
		},
		{
			name:        "valid connection string but unreachable db",
			dbURL:       "postgres://invalid:invalid@localhost:5432/invalid?sslmode=disable",
			shouldError: true,
		},
		{
			name:        "valid test database connection",
			dbURL:       "postgres://username:password@localhost:5432/kirimku_test?sslmode=disable",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Connect(tt.dbURL)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				if db != nil {
					t.Error("Expected nil db but got a connection")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if db == nil {
					t.Error("Expected a db connection but got nil")
				}
				if db != nil {
					err = db.Ping()
					if err != nil {
						t.Errorf("Database connection is not valid: %v", err)
					}
					db.Close()
				}
			}
		})
	}
}
