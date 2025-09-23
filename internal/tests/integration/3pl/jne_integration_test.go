package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/internal/tests/testcases"
)

// TestJNEIntegration tests the JNE client against the mock server
func TestJNEIntegration(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start the mock server (using existing manager)
	mockServerURL := mock_server.StartMockServer(8080)
	defer mock_server.StopMockServer()

	// Reset any previously recorded requests
	mock_server.ResetRecordedRequests()

	// Set up test credentials
	os.Setenv("JNE_USERNAME", "testuser")
	os.Setenv("JNE_API_KEY", "testkey")
	defer func() {
		os.Unsetenv("JNE_USERNAME")
		os.Unsetenv("JNE_API_KEY")
	}()

	// Create test configuration
	cfg := setupTestConfig()

	// Create JNE client with mock server URL
	client := external.NewJNEClient(cfg)

	// Override the base URL to point to our mock server
	if err := client.SetBaseURL(mockServerURL); err != nil {
		t.Fatalf("Failed to set base URL: %v", err)
	}

	// Get test suite
	suite := testcases.GetJNEShippingRateTestSuite()

	// Test summary
	total := len(suite.TestCases)
	passed := 0
	failed := 0

	// Log test suite information
	t.Logf("Running test suite: %s (%s)", suite.Name, suite.Description)
	t.Logf("Total test cases: %d", total)

	// Run each test case
	for _, tc := range suite.TestCases {
		t.Run(fmt.Sprintf("%s - %s", tc.ID, tc.Name), func(t *testing.T) {
			t.Logf("Test case description: %s", tc.Description)

			// Execute the test
			err := tc.Execute(context.Background(), client)

			// Check result
			if tc.Status == testcases.StatusPassed {
				passed++
				t.Logf("✅ PASS: %s", tc.Result)
			} else {
				failed++
				t.Logf("❌ FAIL: %s", tc.Result)
				if err != nil {
					t.Errorf("Error: %v", err)
				}
				t.Fail()
			}

			// Assert test status
			assert.Equal(t, testcases.StatusPassed, tc.Status)

			// Log execution details
			t.Logf("Execution time: %v", tc.Duration)

			// Optional: Verify request was made to the mock server
			jneRequests := mock_server.GetRecordedJNERequests()
			assert.NotEmpty(t, jneRequests, "No JNE requests were recorded")

			// Reset recorded requests for the next test
			mock_server.ResetRecordedRequests()
		})
	}

	// Log summary
	t.Logf("Test suite completed: %d passed, %d failed out of %d total", passed, failed, total)

	// Fail the overall test if any test case failed
	assert.Equal(t, 0, failed, "Some test cases failed")
}

// setupTestConfig creates a test configuration with necessary location data
func setupTestConfig() *config.Config {
	cfg := &config.Config{}

	// Initialize the LocationData with new maps
	cfg.LocationData.PROVINCES = make(map[string]string)
	cfg.LocationData.DISTRICTS = make(map[string]map[string][]string)

	// Add test data
	cfg.LocationData.PROVINCES["Jakarta"] = "DKI Jakarta"
	cfg.LocationData.PROVINCES["Bandung"] = "Jawa Barat"

	cfg.LocationData.DISTRICTS["DKI Jakarta"] = make(map[string][]string)
	cfg.LocationData.DISTRICTS["Jawa Barat"] = make(map[string][]string)

	cfg.LocationData.DISTRICTS["DKI Jakarta"]["Jakarta"] = []string{"Menteng", "Kemayoran"}
	cfg.LocationData.DISTRICTS["Jawa Barat"]["Bandung"] = []string{"Cicendo", "Antapani"}

	return cfg
}
