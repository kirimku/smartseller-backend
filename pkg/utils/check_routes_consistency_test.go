package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestExtractRoutes(t *testing.T) {
	content := `
	mux.HandleFunc("/api/v1/users", userHandler)
	mux.HandleFunc("/api/v1/auth/login", loginHandler)
	mux.HandleFunc("/health", healthHandler)
	`

	routes := make(map[string]bool)
	routes["/api/v1/users"] = true
	routes["/api/v1/auth/login"] = true
	routes["/health"] = true

	extractedRoutes := extractRoutes(content)

	assert.Equal(t, len(routes), len(extractedRoutes), "Route count mismatch")
	for _, route := range extractedRoutes {
		assert.True(t, routes[route], "Unexpected route found: %s", route)
	}
}

func TestExtractOpenAPIRoutes(t *testing.T) {
	// Create a temporary OpenAPI file
	content := []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /api/v1/users:
    get:
      summary: Get users
  /api/v1/auth/login:
    post:
      summary: Login
  /health:
    get:
      summary: Health check
`)

	var openapi OpenAPI
	err := yaml.Unmarshal(content, &openapi)
	assert.NoError(t, err)

	routes := extractOpenAPIRoutes(openapi)

	expectedRoutes := []string{
		"/api/v1/users",
		"/api/v1/auth/login",
		"/health",
	}

	assert.Equal(t, len(expectedRoutes), len(routes), "Route count mismatch")
	for _, expected := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected route not found: %s", expected)
	}
}

func TestCheckConsistency(t *testing.T) {
	tests := []struct {
		name          string
		handlerRoutes []string
		openAPIRoutes []string
		expectError   bool
	}{
		{
			name: "matching routes",
			handlerRoutes: []string{
				"/api/v1/users",
				"/api/v1/auth/login",
				"/health",
			},
			openAPIRoutes: []string{
				"/api/v1/users",
				"/api/v1/auth/login",
				"/health",
			},
			expectError: false,
		},
		{
			name: "missing implementation",
			handlerRoutes: []string{
				"/api/v1/users",
				"/health",
			},
			openAPIRoutes: []string{
				"/api/v1/users",
				"/api/v1/auth/login",
				"/health",
			},
			expectError: true,
		},
		{
			name: "undocumented implementation",
			handlerRoutes: []string{
				"/api/v1/users",
				"/api/v1/auth/login",
				"/health",
				"/undocumented",
			},
			openAPIRoutes: []string{
				"/api/v1/users",
				"/api/v1/auth/login",
				"/health",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkConsistency(tt.handlerRoutes, tt.openAPIRoutes) // No return value needed since it prints results
		})
	}
}
