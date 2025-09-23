package mockserver

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestMockServerAvailability checks if the external mockserver is running
func TestMockServerAvailability(t *testing.T) {
	mockServerURL := os.Getenv("MOCK_SERVER_URL")
	if mockServerURL == "" {
		mockServerURL = "http://localhost:1080"
	}

	// Try to connect to the MockServer
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(mockServerURL)
	if err != nil {
		t.Fatalf("Cannot connect to MockServer at %s: %v", mockServerURL, err)
	}
	defer resp.Body.Close()

	// Check if MockServer is responding
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("MockServer responded with status %d, expected 200", resp.StatusCode)
	}

	fmt.Printf("MockServer is available at %s\n", mockServerURL)
}

// TestJNTEndpoints verifies the JNT API endpoints are correctly mocked
func TestJNTEndpoints(t *testing.T) {
	mockServerURL := os.Getenv("MOCK_SERVER_URL")
	if mockServerURL == "" {
		mockServerURL = "http://localhost:1080"
	}

	// Test endpoints
	endpoints := []string{
		"/jandt_track/inquiry.action",
		"/jandt_track/order.action",
		"/jandt_track/update.action",
	}

	client := &http.Client{Timeout: 5 * time.Second}
	
	for _, endpoint := range endpoints {
		url := mockServerURL + endpoint
		resp, err := client.Get(url)
		if err != nil {
			t.Fatalf("Error calling %s: %v", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Endpoint %s responded with status %d, expected 200", endpoint, resp.StatusCode)
		}

		fmt.Printf("Endpoint %s is correctly mocked\n", endpoint)
	}
}