// Package mockserver provides test utilities for mocking external HTTP services
package mockserver

import (
	"net/http"
	"net/http/httptest"
)

// NewTestServer creates a new mock HTTP server for testing
func NewTestServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

// MockHandler returns a handler that returns the given response
func MockHandler(statusCode int, response []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(response)
	}
}
