package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/domain/model"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external/jnt"
	"github.com/kirimku/smartseller-backend/internal/tests/testcases"
)

// JNTIntegrationTestSuite defines the test suite for JNT client integration tests
type JNTIntegrationTestSuite struct {
	suite.Suite
	client        *external.JNTClient
	mockServerURL string
	config        *config.Config
}

// SetupSuite runs before all tests in the suite
func (s *JNTIntegrationTestSuite) SetupSuite() {
	// Skip integration tests in short mode
	if testing.Short() {
		s.T().Skip("Skipping integration test in short mode")
	}

	// Start the mock server
	s.mockServerURL = mock_server.StartMockServer(8081)

	// Set up test credentials
	os.Setenv("JNT_API_URL", s.mockServerURL)
	os.Setenv("JNT_API_KEY", "test_api_key")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		s.T().Fatalf("Failed to load config: %v", err)
	}
	s.config = cfg

	// Create JNT client
	s.client = external.NewJNTClient(s.config)

	// Override the base URL to point to our mock server
	if err := s.client.SetBaseURL(s.mockServerURL); err != nil {
		s.T().Fatalf("Failed to set base URL: %v", err)
	}

	// Create a test cache
	s.client.SetCache(cache.New(5*time.Minute, 10*time.Minute))
}

// TearDownSuite runs after all tests in the suite
func (s *JNTIntegrationTestSuite) TearDownSuite() {
	// Stop the mock server
	mock_server.StopMockServer()

	// Clean up environment variables
	os.Unsetenv("JNT_API_URL")
	os.Unsetenv("JNT_API_KEY")
}

// SetupTest runs before each test
func (s *JNTIntegrationTestSuite) SetupTest() {
	// Reset recorded requests
	mock_server.ResetRecordedRequests()
}

// TestShippingFeeCalculation tests the shipping fee calculation
func (s *JNTIntegrationTestSuite) TestShippingFeeCalculation() {
	// Create test addresses
	fromAddress := &model.Address{
		Name:     "Sender",
		Phone:    "081234567890",
		City:     "Jakarta",
		Province: "DKI Jakarta",
		District: "Menteng",
		Address:  "Test Address",
	}

	toAddress := &model.Address{
		Name:     "Receiver",
		Phone:    "081234567891",
		City:     "Bandung",
		Province: "Jawa Barat",
		District: "Cicendo",
		Address:  "Test Address",
	}

	// Test shipping fee calculation
	rate, err := s.client.CalculateShippingFee(context.Background(), fromAddress, toAddress, 1, "")

	// Assertions
	s.NoError(err, "Shipping fee calculation should not error")
	s.NotNil(rate, "Shipping rate should not be nil")
	s.Equal("jnt", rate.CourierID, "Courier ID should be jnt")
	s.Equal("jnt_reg", rate.ServiceID, "Service ID should be jnt_reg")
	s.Greater(rate.Price, 0, "Price should be greater than 0")

	// Verify request was made to the mock server
	jntRequests := mock_server.GetRecordedJNTRequests()
	s.NotEmpty(jntRequests, "JNT requests should be recorded")
}

// TestBooking tests the booking functionality
func (s *JNTIntegrationTestSuite) TestBooking() {
	// Create test transaction and addresses
	transaction := &model.Transaction{
		ID:            "TRX12345",
		Amount:        50000,
		Weight:        1,
		InsuranceCost: 0,
		Items: []model.TransactionItem{
			{
				Name:     "Test Item",
				Price:    50000,
				Quantity: 1,
			},
		},
	}

	consignee := &model.Address{
		Name:     "Sender Name",
		Phone:    "081234567890",
		City:     "Jakarta",
		Province: "DKI Jakarta",
		District: "Menteng",
		Address:  "Test Address",
	}

	receiver := &model.Address{
		Name:     "Receiver Name",
		Phone:    "081234567891",
		City:     "Bandung",
		Province: "Jawa Barat",
		District: "Cicendo",
		Address:  "Test Address",
	}

	receipt := &model.JNTReceipt{
		ID:            "1",
		TransactionID: "TRX12345",
		ReceiptNumber: "JNT12345",
		ServiceType:   jnt.ServiceTypeDropoff,
	}

	// Create booking request
	bookingReq := &jnt.BookingRequest{
		Transaction: transaction,
		Consignee:   consignee,
		Receiver:    receiver,
		Receipt:     receipt,
		ServiceType: jnt.ServiceTypeDropoff,
	}

	// Test booking
	booking, err := s.client.Book(context.Background(), bookingReq)

	// Assertions
	s.NoError(err, "Booking should not error")
	s.NotNil(booking, "Booking should not be nil")
	s.Equal("jnt", booking.CourierID, "Courier ID should be jnt")
	s.NotEmpty(booking.BookingCode, "Booking code should not be empty")

	// Verify request was made to the mock server
	jntRequests := mock_server.GetRecordedJNTRequests()
	s.NotEmpty(jntRequests, "JNT requests should be recorded")
}

// TestCancellation tests the cancellation functionality
func (s *JNTIntegrationTestSuite) TestCancellation() {
	// Create cancellation request
	cancellationReq := &jnt.CancellationRequest{
		TransactionID: "TRX12345",
		BookingCode:   "JNT12345",
	}

	// Test cancellation
	cancellation, err := s.client.Cancel(context.Background(), cancellationReq)

	// Assertions
	s.NoError(err, "Cancellation should not error")
	s.NotNil(cancellation, "Cancellation should not be nil")
	s.True(cancellation.Success, "Cancellation should be successful")

	// Verify request was made to the mock server
	jntRequests := mock_server.GetRecordedJNTRequests()
	s.NotEmpty(jntRequests, "JNT requests should be recorded")
}

// TestShippingRateTestSuite tests using the structured test cases
func (s *JNTIntegrationTestSuite) TestShippingRateTestSuite() {
	// Get test suite
	suite := testcases.GetJNTShippingRateTestSuite()

	// Test summary
	total := len(suite.TestCases)
	passed := 0
	failed := 0

	// Log test suite information
	s.T().Logf("Running test suite: %s (%s)", suite.Name, suite.Description)
	s.T().Logf("Total test cases: %d", total)

	// Run each test case
	for _, tc := range suite.TestCases {
		s.T().Run(fmt.Sprintf("%s - %s", tc.ID, tc.Name), func(t *testing.T) {
			t.Logf("Test case description: %s", tc.Description)

			// Execute the test
			err := tc.Execute(context.Background(), s.client)

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

			// Verify request was made to the mock server
			jntRequests := mock_server.GetRecordedJNTRequests()
			if !tc.ExpectError {
				assert.NotEmpty(t, jntRequests, "No JNT requests were recorded")
			}

			// Reset recorded requests for the next test
			mock_server.ResetRecordedRequests()
		})
	}

	// Log summary
	s.T().Logf("Test suite completed: %d passed, %d failed out of %d total", passed, failed, total)

	// Fail the overall test if any test case failed
	assert.Equal(s.T(), 0, failed, "Some test cases failed")
}

// TestJNTClient runs the test suite
func TestJNTClient(t *testing.T) {
	suite.Run(t, new(JNTIntegrationTestSuite))
}
